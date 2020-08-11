package controlCross

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/comm"
	agspudge "github.com/ruraomsk/ag-server/pudge"
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
)

//ClientControlCr информация о подключившемся пользователе
type ClientControlCr struct {
	hub       *HubControlCross
	conn      *websocket.Conn
	send      chan ControlSokResponse
	regStatus chan bool
	crossInfo crossSock.CrossInfo
}

var UserLogoutCrControl chan string

//readPump обработчик чтения сокета
func (c *ClientControlCr) readPump(db *sqlx.DB) {
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.unregister <- c
			break
		}
		//ну отправка и отправка
		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /cross |Message: %v \n", c.crossInfo.Ip, c.crossInfo.Login, err.Error())
			resp := newControlMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
		}
		switch typeSelect {
		case typeSendB: //отправка state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				var (
					userCross = agspudge.UserCross{User: c.crossInfo.Login, State: temp.State}
					mess      = tcpConnect.TCPMessage{
						User:        c.crossInfo.Login,
						TCPType:     tcpConnect.TypeState,
						Idevice:     temp.State.IDevice,
						Data:        userCross,
						From:        tcpConnect.FromCrControlSoc,
						CommandType: typeSendB,
						Pos:         c.crossInfo.Pos,
					}
				)
				mess.SendToTCPServer()
			}
		case typeCheckB: //проверка state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := checkCrossData(temp.State)
				c.send <- resp
			}
		case typeCreateB: //создание перекрестка
			{
				temp := struct {
					Type  string         `json:"type"`
					State agspudge.Cross `json:"state"`
					Z     int            `json:"z"`
				}{}
				_ = json.Unmarshal(p, &temp)
				resp := newControlMess(typeCreateB, nil)
				resp.Data = createCrossData(temp.State, c.crossInfo.Pos, c.crossInfo.Login, temp.Z, db)
				c.send <- resp
			}
		case typeDeleteB: //удаление state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				userCross := agspudge.UserCross{User: c.crossInfo.Login, State: temp.State}
				userCross.State.IDevice = -1
				mess := tcpConnect.TCPMessage{
					User:        c.crossInfo.Login,
					TCPType:     tcpConnect.TypeState,
					Idevice:     temp.State.IDevice,
					Data:        userCross,
					From:        tcpConnect.FromCrControlSoc,
					CommandType: typeDeleteB,
					Pos:         c.crossInfo.Pos,
				}
				mess.SendToTCPServer()
			}
		case typeUpdateB: //обновление state
			{
				resp := newControlMess(typeUpdateB, nil)
				resp, _, _ = takeControlInfo(c.crossInfo.Pos, db)
				resp.Data["edit"] = c.crossInfo.Edit
				c.send <- resp
			}
		case typeEditInfoB: //информация о пользователях что редактируют перекресток
			{
				resp := newControlMess(typeEditInfoB, nil)

				type usersEdit struct {
					User string `json:"user"`
					Edit bool   `json:"edit"`
				}
				var users []usersEdit
				//нужно ха этим посмотреть но проблем не должно быть
				for client := range c.hub.clients {
					if client.crossInfo.Pos == c.crossInfo.Pos {
						temp := usersEdit{User: client.crossInfo.Login, Edit: client.crossInfo.Edit}
						users = append(users, temp)
					}
				}
				resp.Data["users"] = users
				c.send <- resp
			}
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = c.crossInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        c.crossInfo.Login,
					TCPType:     tcpConnect.TypeDispatch,
					Idevice:     arm.ID,
					Data:        arm,
					From:        tcpConnect.FromCrControlSoc,
					CommandType: typeDButton,
					Pos:         c.crossInfo.Pos,
				}
				mess.SendToTCPServer()

			}
		default:
			{
				fmt.Println(typeSelect)
				resp := newControlMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientControlCr) writePump() {
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
