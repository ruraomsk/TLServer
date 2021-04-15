package techArm

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	"strconv"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	devUpdate           = time.Second * 1
	checkTokensValidity = time.Minute * 1
)

var UserLogoutTech chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientTechArm информация о подключившемся пользователе
type ClientTechArm struct {
	hub  *HubTechArm
	conn *websocket.Conn
	send chan armResponse

	armInfo *ArmInfo
}

//readPump обработчик чтения сокета
func (c *ClientTechArm) readPump() {
	//db := data.GetDB("ClientTechArm")
	//defer data.FreeDB("ClientTechArm")
	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{
		var tempCrosses = make([]CrossInfo, 0)
		crosses := getCross(c.armInfo.Region)
		for _, cross := range crosses {
			for _, area := range c.armInfo.Area {
				tArea, _ := strconv.Atoi(area)
				if cross.Area == tArea {
					tempCrosses = append(tempCrosses, cross)
				}
			}
		}
		resp := newArmMess(typeArmInfo, nil)
		resp.Data[typeCrosses] = tempCrosses

		var tempDevises = make([]DevInfo, 0)
		devices := getDevice()
		for _, dev := range devices {
			for _, area := range c.armInfo.Area {
				tArea, _ := strconv.Atoi(area)
				if dev.Area == tArea && dev.Region == c.armInfo.Region {
					tempDevises = append(tempDevises, dev)
				}
			}
		}
		resp.Data[typeDevices] = tempDevises
		resp.Data["gprs"] = GPRSInfo
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
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /techArm |Message: %v \n", c.armInfo.AccInfo.IP, c.armInfo.AccInfo.Login, err.Error())
			resp := newArmMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
			continue
		}
		switch typeSelect {
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = c.armInfo.AccInfo.Login
				reset := false
				if arm.Command == 4 {
					device.GlobalDevEdit.Mux.Lock()
					tDev := device.GlobalDevEdit.MapDevices[arm.ID]
					if arm.Params == 1 {
						tDev.TurnOnFlag = true
					} else if arm.Params == 0 {
						tDev.TurnOnFlag = false
						tDev.BusyCount = 0
						reset = true
					}
					device.GlobalDevEdit.MapDevices[arm.ID] = tDev
					device.GlobalDevEdit.Mux.Unlock()
					if reset {
						arm.Params = -1
						var mess = tcpConnect.TCPMessage{
							User:        arm.User,
							TCPType:     tcpConnect.TypeDispatch,
							From:        tcpConnect.FromGsSoc,
							CommandType: typeDButton,
							Pos:         sockets.PosInfo{},
						}
						mess.SendToTCPServer()
					}
				}
				var mess = tcpConnect.TCPMessage{
					User:        c.armInfo.AccInfo.Login,
					TCPType:     tcpConnect.TypeDispatch,
					Idevice:     arm.ID,
					Data:        arm,
					From:        tcpConnect.FromTechArmSoc,
					CommandType: typeDButton,
				}
				mess.SendToTCPServer()
			}
		case typeGPRS:
			{
				var temp = struct {
					Type string              `json:"type"`
					Gprs comm.ChangeProtocol `json:"gprs"`
				}{}

				_ = json.Unmarshal(p, &temp)
				temp.Gprs.User = c.armInfo.AccInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        c.armInfo.AccInfo.Login,
					TCPType:     tcpConnect.TypeChangeProtocol,
					Idevice:     temp.Gprs.ID,
					Data:        temp.Gprs,
					From:        tcpConnect.FromTechArmSoc,
					CommandType: typeGPRS,
				}
				mess.SendToTCPServer()
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientTechArm) writePump() {
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
