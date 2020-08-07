package mainMap

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/internal/sockets/maps"
	"github.com/JanFant/TLServer/logger"
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

	login string
	ip    string
}

//readPump обработчик чтения сокета
func (c *ClientMainMap) readPump(db *sqlx.DB, gc *gin.Context) {
	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{
		flag, tk := checkToken(gc, db)
		resp := newMapMess(typeMapInfo, maps.MapOpenInfo(db))
		if flag {
			login := tk.Login
			role := tk.Role
			resp.Data["role"] = tk.Role
			resp.Data["manageFlag"], _ = data.AccessCheck(login, role, 2)
			resp.Data["logDeviceFlag"], _ = data.AccessCheck(login, role, 5)
			resp.Data["techArmFlag"], _ = data.AccessCheck(login, role, 7)
			resp.Data["gsFlag"], _ = data.AccessCheck(login, role, 8)
			resp.Data["description"] = tk.Description
			resp.Data["authorizedFlag"] = true
			resp.Data["region"] = tk.Region
			var areaMap = make(map[string]string)
			for _, area := range tk.Area {
				var tempA data.AreaInfo
				tempA.SetAreaInfo(tk.Region, area)
				areaMap[tempA.Num] = tempA.NameArea
			}
			resp.Data["area"] = areaMap
			data.CacheArea.Mux.Lock()
			resp.Data["areaZone"] = data.CacheArea.Areas
			data.CacheArea.Mux.Unlock()
			c.login = login
		}
		c.send <- resp
	}
	crossSock.GetCrossUserForMap <- true
	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.unregister <- c
			break
		}
		//ну отправка и отправка
		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /map |Message: %v \n", c.ip, c.login, err.Error())
			resp := newMapMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
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
				resp := newMapMess(typeLogin, nil)
				resp.Data = logIn(account.Login, account.Password, c.conn.RemoteAddr().String(), db)
				if _, ok := resp.Data["message"]; !ok {
					c.login = fmt.Sprint(resp.Data["login"])
				}
				c.send <- resp
			}
		case typeChangeAccount:
			{
				account := &data.Account{}
				_ = json.Unmarshal(p, &account)
				resp := newMapMess(typeLogin, nil)
				resp.Data = logIn(account.Login, account.Password, c.conn.RemoteAddr().String(), db)
				if _, ok := resp.Data["message"]; !ok {
					//делаем выход из аккаунта
					respLO := newMapMess(typeLogOut, nil)
					status := logOut(c.login, db)
					if status {
						logOutSockets(c.login)
						c.login = fmt.Sprint(resp.Data["login"])
					}
					c.send <- respLO
				}
				c.send <- resp
			}
		case typeLogOut: //отправка default
			{
				if c.login != "" {
					resp := newMapMess(typeLogOut, nil)
					status := logOut(c.login, db)
					if status {
						resp.Data["authorizedFlag"] = false
						logOutSockets(c.login)
					}
					c.login = ""
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
					User:        c.login,
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
