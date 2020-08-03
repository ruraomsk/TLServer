package xctrl

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/xcontrol"
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
)

var UserLogoutXctrl chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientXctrl информация о подключившемся пользователе
type ClientXctrl struct {
	hub  *HubXctrl
	conn *websocket.Conn
	send chan MessXctrl

	login string
}

//readPump обработчик чтения сокета
func (c *ClientXctrl) readPump(db *sqlx.DB) {
	//defer func() {
	//	c.hub.unregister <- c
	//	_ = c.conn.Close()
	//}()

	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{
		allXctrl, err := getXctrl(db)
		if err != nil {
			resp := newXctrlMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errGetXctrl}
			c.send <- resp
		}
		resp := newXctrlMess(typeXctrlInfo, nil)
		resp.Data[typeXctrlInfo] = allXctrl
		c.send <- resp
	}

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			//resp := newXctrlMess(typeClose, nil)
			//c.send <- resp
			c.hub.unregister <- c
			break
		}
		//ну отправка и отправка
		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			resp := newXctrlMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
		}
		switch typeSelect {
		case typeXctrlChange:
			{
				temp := struct {
					SType string           `json:"type"`
					State []xcontrol.State `json:"state"`
				}{}
				_ = json.Unmarshal(p, &temp)
				err := writeXctrl(temp.State, db)
				if err != nil {
					resp := newXctrlMess(typeError, nil)
					resp.Data["message"] = ErrorMessage{Error: errChangeXctrl}
					c.send <- resp
				}
				resp := newXctrlMess(typeXctrlChange, nil)
				resp.Data["message"] = "ok"
				c.send <- resp
			}
		default:
			{
				fmt.Println("asdasd")
				resp := newXctrlMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientXctrl) writePump(db *sqlx.DB) {
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

				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				_ = json.NewEncoder(w).Encode(mess)

				// Add queued chat messages to the current websocket message.
				n := len(c.send)
				for i := 0; i < n; i++ {
					_, _ = w.Write([]byte{'\n'})
					_ = json.NewEncoder(w).Encode(mess)
				}

				if err := w.Close(); err != nil {
					return
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
