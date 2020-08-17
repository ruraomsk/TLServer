package chat

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
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

var UserLogoutChat chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientChat информация о подключившемся пользователе
type ClientChat struct {
	hub        *HubChat
	conn       *websocket.Conn
	send       chan chatResponse
	clientInfo clientInfo
}

//readPump обработчик чтения сокета
func (c *ClientChat) readPump(db *sqlx.DB) {
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	//выгрузить список доступных пользователей
	{
		users, err := getAllUsers(c.hub, db)
		if err != nil {
			resp := newChatMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errNoAccessWithDatabase}
			c.send <- resp
		} else {
			resp := newChatMess(typeAllUsers, nil)
			resp.Data["users"] = users
			c.send <- resp
		}
	}
	//выгрузить архив сообщений за последний день
	{
		var arc = ArchiveMessages{TimeStart: time.Now(), TimeEnd: time.Now().AddDate(0, 0, -1), To: globalMessage}
		err := arc.takeArchive(db)
		if err != nil {
			resp := newChatMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errNoAccessWithDatabase}
			c.send <- resp
		} else {
			resp := newChatMess(typeArchive, nil)
			resp.Data[typeArchive] = arc
			c.send <- resp
		}
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
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /cross |Message: %v \n", c.clientInfo.ip, c.clientInfo.login, err.Error())
			resp := newChatMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
			continue
		}
		switch typeSelect {
		case typeMessage:
			{
				var mF Message
				_ = json.Unmarshal(p, &mF)
				if err := mF.SaveMessage(db); err != nil {
					resp := newChatMess(typeError, nil)
					resp.Data["message"] = ErrorMessage{Error: errNoAccessWithDatabase}
					c.send <- resp
				} else {
					resp := newChatMess(typeMessage, nil)
					resp.Data["message"] = mF.Message
					resp.Data["time"] = mF.Time
					resp.Data["from"] = mF.From
					resp.Data["to"] = mF.To
					resp.to = mF.To
					resp.from = mF.From
					c.hub.broadcast <- resp
				}
			}
		default:
			{
				fmt.Println(typeSelect)
				resp := newChatMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientChat) writePump() {
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
