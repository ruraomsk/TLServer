package controlCross

import (
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
	maxMessageSize = 1024 * 100
)

//ClientCross информация о подключившемся пользователе
type ClientCross struct {
	hub  *HubCross
	conn *websocket.Conn
	send chan crossResponse

	login string
}

//readPump обработчик чтения сокета
func (c *ClientCross) readPump(db *sqlx.DB) {
	//defer func() {
	//	c.hub.unregister <- c
	//	_ = c.conn.Close()
	//}()

	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{

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
			logger.Error.Printf("|IP: - |Login: %v |Resource: /charPoint |Message: %v \n", c.login, err.Error())
			resp := newCrossMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
		}
		switch typeSelect {
		default:
			{
				fmt.Println("asdasd")
				resp := newCrossMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientCross) writePump(db *sqlx.DB) {
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
