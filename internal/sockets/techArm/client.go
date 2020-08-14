package techArm

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
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

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 100

	devUpdate = time.Second * 1
)

var UserLogoutTech chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientTechArm информация о подключившемся пользователе
type ClientTechArm struct {
	hub  *HubTechArm
	conn *websocket.Conn
	send chan armResponse

	armInfo ArmInfo
}

//readPump обработчик чтения сокета
func (c *ClientTechArm) readPump(db *sqlx.DB) {

	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{
		var tempCrosses = make([]CrossInfo, 0)
		crosses := getCross(c.armInfo.Region, db)
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
		devices := getDevice(db)
		for _, dev := range devices {
			for _, area := range c.armInfo.Area {
				tArea, _ := strconv.Atoi(area)
				if dev.Area == tArea && dev.Region == c.armInfo.Region {
					tempDevises = append(tempDevises, dev)
				}
			}
		}
		resp.Data[typeDevices] = tempDevises
		resp.Data["gps"] = GPSInfo
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
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /techArm |Message: %v \n", c.armInfo.ip, c.armInfo.Login, err.Error())
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
				arm.User = c.armInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        c.armInfo.Login,
					TCPType:     tcpConnect.TypeDispatch,
					Idevice:     arm.ID,
					Data:        arm,
					From:        tcpConnect.FromTechArmSoc,
					CommandType: typeDButton,
				}
				mess.SendToTCPServer()
			}
		case typeGPS:
			{
				gps := comm.ChangeProtocol{}
				_ = json.Unmarshal(p, &gps)
				gps.User = c.armInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        c.armInfo.Login,
					TCPType:     tcpConnect.TypeChangeProtocol,
					Idevice:     gps.ID,
					Data:        gps,
					From:        tcpConnect.FromTechArmSoc,
					CommandType: typeGPS,
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
