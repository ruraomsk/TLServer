package mainMap

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock/mainCross"
	"github.com/JanFant/TLServer/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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

	crossTick = time.Second * 5
)

var UserLogoutGS chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientMainMap информация о подключившемся пользователе
type ClientMainMap struct {
	hub  *HubMainMap
	conn *websocket.Conn
	send chan mapResponse

	cInfo *clientInfo
}

//clientInfo информация о клиенту
type clientInfo struct {
	login    string     //логин
	tokenStr string     //строка токена пользователя
	token    *jwt.Token //токен
	ip       []string   //разделенный ip
}

//readPump обработчик чтения сокета
func (c *ClientMainMap) readPump(db *sqlx.DB, gc *gin.Context) {
	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

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
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /map |Message: %v \n", c.cInfo.ip[0], c.cInfo.login, err.Error())
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
				account := &data.Account{}
				_ = json.Unmarshal(p, &account)

				var (
					resp  = newMapMess(typeLogin, nil)
					token *jwt.Token
				)
				resp.Data, token = logIn(account.Login, account.Password, c.conn.RemoteAddr().String(), db)
				if _, ok := resp.Data["message"]; !ok {
					c.cInfo.login = fmt.Sprint(resp.Data["login"])
					c.cInfo.tokenStr = fmt.Sprint(resp.Data["token"])
					c.cInfo.token = token
				}
				c.send <- resp
			}
		case typeChangeAccount:
			{
				account := &data.Account{}
				_ = json.Unmarshal(p, &account)
				var (
					resp  = newMapMess(typeLogin, nil)
					token *jwt.Token
				)
				resp.Data, token = logIn(account.Login, account.Password, c.conn.RemoteAddr().String(), db)
				if _, ok := resp.Data["message"]; !ok {
					//делаем выход из аккаунта
					respLO := newMapMess(typeLogOut, nil)
					status := logOut(c.cInfo.login, db)
					if status {
						logOutSockets(c.cInfo.login)
						c.cInfo.login = fmt.Sprint(resp.Data["login"])
						c.cInfo.tokenStr = fmt.Sprint(resp.Data["token"])
						c.cInfo.token = token
					}
					c.send <- respLO
				}
				c.send <- resp
			}
		case typeLogOut: //отправка default
			{
				if c.cInfo.login != "" {
					resp := newMapMess(typeLogOut, nil)
					status := logOut(c.cInfo.login, db)
					if status {
						resp.Data["authorizedFlag"] = false
						logOutSockets(c.cInfo.login)
					}
					c.cInfo.login = ""
					c.cInfo.tokenStr = ""
					c.cInfo.token = nil
					c.send <- resp
				}
			}
		case typeCheckConn: //отправка default
			{
				resp := newMapMess(typeCheckConn, nil)
				statusDB := false
				_, err := db.Exec(`SELECT * FROM public.accounts;`)
				if err == nil {
					statusDB = true
				}
				resp.Data["statusBD"] = statusDB
				var tcpPackage = tcpConnect.TCPMessage{
					TCPType:     tcpConnect.TypeState,
					User:        c.cInfo.login,
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
				fmt.Println("asdasd")
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
