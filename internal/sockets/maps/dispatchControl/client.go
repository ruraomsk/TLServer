package dispatchControl

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/maps"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	crossPeriod         = time.Second * 5
	devicePeriod        = time.Second * 5
	checkTokensValidity = time.Minute * 1
)

//ClientDC информация о подключившемся пользователе
type ClientDC struct {
	hub        *HubDispCtrl
	conn       *websocket.Conn
	send       chan dCResponse
	cInfo      *accToken.Token
	devices    []int
	sendPhases bool
}
type Phase struct {
	Device int `json:"device"`
	Phase  int `json:"phase"`
}

//readPump обработчик чтения сокета
func (c *ClientDC) readPump() {
	//db := data.GetDB("ClientGS")
	//defer data.FreeDB("ClientGS")
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{
		resp := newDCMess(typeMapInfo, maps.MapOpenInfo())
		data.CacheArea.Mux.Lock()
		resp.Data["areaZone"] = data.CacheArea.Areas
		data.CacheArea.Mux.Unlock()
		if c.sendPhases {
			resp.Data[typePhases] = getPhases(c.devices)
		}
		c.send <- resp
	}

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.unregister <- c
			break
		}
		//ну отправка и отправка
		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /greenStreet |Message: %v \n", c.cInfo.IP, c.cInfo.Login, err.Error())
			resp := newDCMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
			continue
		}
		switch typeSelect {
		case typeJump: //отправка default
			{
				location := &data.Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := newDCMess(typeJump, nil)
				resp.Data["boxPoint"] = box
				c.send <- resp
			}
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				if arm.Command == 4 && arm.Params == 1 {
					found := false
					for _, d := range c.devices {
						if d == arm.ID {
							found = true
							break
						}
					}
					if !found {
						c.devices = append(c.devices, arm.ID)
						c.sendPhases = true
						//logger.Debug.Printf("добавил в device %v ",c.devices)
					}
				}
				if arm.Command == 4 && arm.Params == 0 {
					for i, d := range c.devices {
						if d == arm.ID {
							if len(c.devices) <= 1 {
								c.devices = make([]int, 0)
							} else {
								if i+1 >= len(c.devices) {
									c.devices = c.devices[:i]
								} else {
									c.devices = append(c.devices[:i], c.devices[i+1:]...)
								}
							}
							break
						}
					}
					//logger.Debug.Printf("убавил в device %v ",c.devices)
					if len(c.devices) == 0 {
						c.sendPhases = false
					}
				}
				arm.User = c.cInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        arm.User,
					TCPType:     tcpConnect.TypeDispatch,
					Idevice:     arm.ID,
					Data:        arm,
					From:        tcpConnect.FromGsSoc,
					CommandType: typeDButton,
					Pos:         sockets.PosInfo{},
				}
				mess.SendToTCPServer()
			}
		case typeRoute:
			{
				execRoute := executeRoute{}
				_ = json.Unmarshal(p, &execRoute)

				arm := comm.CommandARM{Command: 4, User: c.cInfo.Login}
				var mess = tcpConnect.TCPMessage{
					User:        c.cInfo.Login,
					TCPType:     tcpConnect.TypeDispatch,
					From:        tcpConnect.FromGsSoc,
					CommandType: typeDButton,
					Pos:         sockets.PosInfo{},
				}
				if execRoute.TurnOn {
					c.sendPhases = true
					c.devices = execRoute.Devices
					//logger.Debug.Printf("client devs %v",c.devices)
					arm.Params = 1
					device.GlobalDevEdit.Mux.Lock()
					for _, dev := range execRoute.Devices {
						tDev := device.GlobalDevEdit.MapDevices[dev]
						if tDev.BusyCount == 0 || tDev.TurnOnFlag == false {
							arm.ID = dev
							mess.Idevice = arm.ID
							mess.Data = arm
							mess.SendToTCPServer()
							tDev.TurnOnFlag = true
						}
						tDev.BusyCount++
						device.GlobalDevEdit.MapDevices[dev] = tDev
					}
					device.GlobalDevEdit.Mux.Unlock()
				} else {
					c.sendPhases = false
					arm.Params = 0
					device.GlobalDevEdit.Mux.Lock()
					for _, dev := range c.devices {
						tDev := device.GlobalDevEdit.MapDevices[dev]
						tDev.BusyCount--
						if tDev.BusyCount == 0 && tDev.TurnOnFlag == true {
							arm.ID = dev
							mess.Idevice = arm.ID
							mess.Data = arm
							mess.SendToTCPServer()
							tDev.TurnOnFlag = false
						}
						device.GlobalDevEdit.MapDevices[dev] = tDev
					}
					device.GlobalDevEdit.Mux.Unlock()
					c.devices = make([]int, 0)
				}
			}
		default:
			{
				resp := newDCMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientDC) writePump() {
	pingTick := time.NewTicker(pingPeriod)
	defer func() {
		pingTick.Stop()
	}()
	for {
		select {
		case mess, ok := <-c.send:
			{
				_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "канал был закрыт"))
					return
				}
				_ = c.conn.WriteJSON(mess)
				// Add queued chat messages to the current websocket message.
				n := len(c.send)
				for i := 0; i < n; i++ {
					_ = c.conn.WriteJSON(<-c.send)
				}
			}
		case <-pingTick.C:
			{
				_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}
}
