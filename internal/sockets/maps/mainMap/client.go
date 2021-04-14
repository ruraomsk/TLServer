package mainMap

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/crossSock/mainCross"
	"github.com/ruraomsk/TLServer/logger"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	crossTick           = time.Second * 5
	checkTokensValidity = time.Minute * 1
)

var UserLogoutGS chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientMainMap информация о подключившемся пользователе
type ClientMainMap struct {
	hub  *HubMainMap
	conn *websocket.Conn
	send chan mapResponse

	cInfo    *accToken.Token
	rawToken string
	cookie   string
}

//readPump обработчик чтения сокета
func (c *ClientMainMap) readPump() {
	//если нужно указать лимит пакета
	db := data.GetDB("ClientMainMap")
	defer data.FreeDB("ClientMainMap")
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	mainCross.GetCrossUserForMap <- true
	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.unregister <- c
			break
		}
		//ну отправка и отправка
		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /map |Message: %v \n", c.cInfo.IP, c.cInfo.Login, err.Error())
			resp := newMapMess(typeError, nil)
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
				resp := newMapMess(typeJump, nil)
				resp.Data["boxPoint"] = box
				c.send <- resp
			}
		case typeLogin: //отправка default
			{
				var (
					account  = &data.Account{}
					token    *accToken.Token
					tokenStr string
				)
				_ = json.Unmarshal(p, &account)
				resp := newMapMess(typeLogin, nil)
				resp.Data, token, tokenStr = logIn(account.Login, account.Password, c.conn.RemoteAddr().String(), db)
				if token != nil {
					//делаем выход из аккаунта
					for client := range c.hub.clients {
						if client.cInfo.Login == account.Login {
							logOutSockets(account.Login)
							respLO := newMapMess(typeLogOut, nil)
							client.send <- respLO
							break
						}
					}
					c.cInfo = token
					c.cookie = tokenStr
				}
				c.send <- resp
			}
		case typeChangeAccount:
			{
				var (
					account  = &data.Account{}
					token    *accToken.Token
					tokenStr string
				)
				_ = json.Unmarshal(p, &account)
				resp := newMapMess(typeLogin, nil)
				resp.Data, token, tokenStr = logIn(account.Login, account.Password, c.conn.RemoteAddr().String(), db)
				if token != nil {
					//делаем выход из аккаунта
					respLO := newMapMess(typeLogOut, nil)
					status := logOut(c.cInfo.Login, db)
					if status {
						logOutSockets(c.cInfo.Login)
						c.cInfo = token
						c.cookie = tokenStr
					}
					c.send <- respLO
				}
				c.send <- resp
			}
		case typeLogOut: //отправка default
			{
				if c.cInfo.Login != "" {
					resp := newMapMess(typeLogOut, nil)
					status := logOut(c.cInfo.Login, db)
					if status {
						resp.Data["authorizedFlag"] = false
						logOutSockets(c.cInfo.Login)
					}
					c.cInfo = new(accToken.Token)
					c.cookie = ""
					c.send <- resp
				}
			}
		case typeCheckConn: //отправка default
			{
				resp := newMapMess(typeCheckConn, nil)
				statusDB := false
				if data.GetDB("CheckConn") != nil {
					statusDB = true
					data.FreeDB("CheckConn")
				}
				resp.Data["statusBD"] = statusDB
				var tcpPackage = tcpConnect.TCPMessage{
					TCPType:     tcpConnect.TypeState,
					User:        c.cInfo.Login,
					Idevice:     -1,
					Data:        0,
					From:        tcpConnect.FromMapSoc,
					CommandType: typeDButton,
				}
				tcpPackage.SendToTCPServer()

				c.send <- resp
			}
		default:
			{
				resp := newMapMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientMainMap) writePump() {
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
