package xctrl

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/logger"
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

	stateTime = time.Second * 20
)

var UserLogoutXctrl chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientXctrl информация о подключившемся пользователе
type ClientXctrl struct {
	hub  *HubXctrl
	conn *websocket.Conn
	send chan MessXctrl

	login string
	ip    string
}

//readPump обработчик чтения сокета
func (c *ClientXctrl) readPump(db *sqlx.DB) {

	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	{
		allXctrl, err := getXctrl(db)
		if err != nil {
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /charPoint |Message: %v \n", c.ip, c.login, err.Error())
			resp := newXctrlMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errGetXctrl}
			c.send <- resp
			goto next
		}
		resp := newXctrlMess(typeXctrlInfo, nil)
		//собираю в кучу регионы для отображения
		chosenRegion := make(map[string]string)
		data.CacheInfo.Mux.Lock()
		for first, second := range data.CacheInfo.MapRegion {
			chosenRegion[first] = second
		}
		delete(chosenRegion, "*")
		resp.Data["regionInfo"] = chosenRegion

		//собираю в кучу районы для отображения
		chosenArea := make(map[string]map[string]string)
		for first, second := range data.CacheInfo.MapArea {
			chosenArea[first] = make(map[string]string)
			chosenArea[first] = second
		}
		delete(chosenArea, "Все регионы")
		data.CacheInfo.Mux.Unlock()
		resp.Data["areaInfo"] = chosenArea
		resp.Data[typeXctrlInfo] = allXctrl
		c.send <- resp
	}

next:

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.unregister <- c
			break
		}
		//ну отправка и отправка
		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /charPoint |Message: %v \n", c.ip, c.login, err.Error())
			resp := newXctrlMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
			continue
		}
		switch typeSelect {
		case typeXctrlChange:
			{
				temp := struct {
					SType string           `json:"type"`
					State []xcontrol.State `json:"state"`
				}{}
				_ = json.Unmarshal(p, &temp)
				err := changeXctrl(temp.State, db)
				if err != nil {
					logger.Error.Printf("|IP: %v |Login: %v |Resource: /charPoint |Message: %v \n", c.ip, c.login, err.Error())
					resp := newXctrlMess(typeError, nil)
					resp.Data["message"] = ErrorMessage{Error: errChangeXctrl}
					c.send <- resp
					continue
				}
				resp := newXctrlMess(typeXctrlChange, nil)
				resp.Data["message"] = "ok"
				c.send <- resp
				//обновим данные
				respI := newXctrlMess(typeXctrlReInfo, nil)
				allXctrl, _ := getXctrl(db)
				respI.Data[typeXctrlInfo] = allXctrl
				c.send <- respI
			}
		case typeXctrlCreate:
			{
				temp := struct {
					SType string         `json:"type"`
					State xcontrol.State `json:"state"`
				}{}
				_ = json.Unmarshal(p, &temp)
				err := createXctrl(temp.State, db)
				if err != nil {
					logger.Error.Printf("|IP: %v |Login: %v |Resource: /charPoint |Message: %v \n", c.ip, c.login, err.Error())
					resp := newXctrlMess(typeError, nil)
					resp.Data["message"] = ErrorMessage{Error: errChangeXctrl}
					c.send <- resp
					continue
				}
				resp := newXctrlMess(typeXctrlCreate, nil)
				resp.Data["message"] = "ok"
				c.send <- resp
				//обновим данные
				respI := newXctrlMess(typeXctrlReInfo, nil)
				allXctrl, _ := getXctrl(db)
				respI.Data[typeXctrlInfo] = allXctrl
				c.send <- respI
			}
		case typeXctrlDelete:
			{
				temp := struct {
					SType string         `json:"type"`
					State xcontrol.State `json:"state"`
				}{}
				_ = json.Unmarshal(p, &temp)
				err := deleteXctrl(temp.State, db)
				if err != nil {
					logger.Error.Printf("|IP: %v |Login: %v |Resource: /charPoint |Message: %v \n", c.ip, c.login, err.Error())
					resp := newXctrlMess(typeError, nil)
					resp.Data["message"] = ErrorMessage{Error: errChangeXctrl}
					c.send <- resp
					continue
				}
				resp := newXctrlMess(typeXctrlDelete, nil)
				resp.Data["message"] = "ok"
				c.send <- resp
				//обновим данные
				respI := newXctrlMess(typeXctrlReInfo, nil)
				allXctrl, _ := getXctrl(db)
				respI.Data[typeXctrlInfo] = allXctrl
				c.send <- respI
			}
		case typeGetArea:
			{
				temp := struct {
					Region int `json:"region"`
					Area   int `json:"area"`
				}{}
				_ = json.Unmarshal(p, &temp)
				tfLight, err := getSubAreaTF(temp.Region, temp.Area, db)
				if err != nil {
					logger.Error.Printf("|IP: %v |Login: %v |Resource: /charPoint |Message: %v \n", c.ip, c.login, err.Error())
					resp := newXctrlMess(typeError, nil)
					resp.Data["message"] = ErrorMessage{Error: errChangeXctrl}
					c.send <- resp
					continue
				}
				resp := newXctrlMess(typeGetArea, nil)
				resp.Data["tflight"] = tfLight
				c.send <- resp
			}
		default:
			{
				fmt.Println("asdasd")
				resp := newXctrlMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
				continue
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientXctrl) writePump() {
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
