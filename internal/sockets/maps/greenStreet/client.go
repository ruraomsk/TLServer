package greenStreet

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/accToken"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/routeGS"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/maps"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
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
	checkTokensValidity = time.Minute * 1
)

//ClientGS информация о подключившемся пользователе
type ClientGS struct {
	hub  *HubGStreet
	conn *websocket.Conn
	send chan gSResponse

	cInfo   *accToken.Token
	devices []int
}

//readPump обработчик чтения сокета
func (c *ClientGS) readPump(db *sqlx.DB) {
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{
		resp := newGSMess(typeMapInfo, maps.MapOpenInfo(db))
		resp.Data["routes"] = getAllModes(db)
		data.CacheArea.Mux.Lock()
		resp.Data["areaZone"] = data.CacheArea.Areas
		data.CacheArea.Mux.Unlock()
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
			resp := newGSMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
			continue
		}
		switch typeSelect {
		case typeCreateRout:
			{
				temp := routeGS.Route{}
				_ = json.Unmarshal(p, &temp)
				resp := newGSMess(typeCreateRout, nil)
				err := temp.Create(db)
				if err != nil {
					resp.Data[typeError] = errCantWriteInBD
					c.send <- resp
				} else {
					resp.Data["route"] = temp
					c.hub.broadcast <- resp
				}
			}
		case typeUpdateRout:
			{
				temp := routeGS.Route{}
				_ = json.Unmarshal(p, &temp)
				resp := newGSMess(typeUpdateRout, nil)
				err := temp.Update(db)
				if err != nil {
					resp.Data[typeError] = errCantWriteInBD
					c.send <- resp
				} else {
					resp.Data["route"] = temp
					c.hub.broadcast <- resp
				}
			}
		case typeDeleteRout:
			{
				temp := routeGS.Route{}
				_ = json.Unmarshal(p, &temp)
				resp := newGSMess(typeDeleteRout, nil)
				err := temp.Delete(db)
				if err != nil {
					resp.Data[typeError] = errCantDeleteFromBD
					c.send <- resp
				} else {
					resp.Data["route"] = temp
					c.hub.broadcast <- resp
				}
			}
		case typeJump: //отправка default
			{
				location := &data.Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := newGSMess(typeJump, nil)
				resp.Data["boxPoint"] = box
				c.send <- resp
			}
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
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
		default:
			{
				fmt.Println("asdasd")
				resp := newGSMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientGS) writePump() {
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
