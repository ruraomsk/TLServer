package alarm

import (
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"sort"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	devUpdate           = time.Second * 1
	checkTokensValidity = time.Minute * 1
)

var UserLogoutAlarm chan string //канал для закрытия сокетов, пользователя который вышел из системы

//ClientTechArm информация о подключившемся пользователе
type ClientAlarm struct {
	hub  *HubAlarm
	conn *websocket.Conn
	send chan alarmResponse

	armInfo   *Info
	CrossRing *CrossRing
}

//makeResponse всегда готовит изменение и ответ для клиента
func (c *ClientAlarm) makeResponse() {
	c.CrossRing.Ring = false
	change := false
	for _, nc := range getCross(c.armInfo.Region, c.hub.db) {
		_, is := c.CrossRing.CrossInfo[key(nc.Region, nc.Area, nc.ID)]
		if !nc.Control {
			if !is {
				//Первый раз попадает в мапу
				c.CrossRing.Ring = true
				c.CrossRing.CrossInfo[key(nc.Region, nc.Area, nc.ID)] = nc
				change = true
			}
			continue
		}
		//Есть управление
		if !is {
			continue
		}
		//Была прошлый раз неисправна
		change = true
		delete(c.CrossRing.CrossInfo, key(nc.Region, nc.Area, nc.ID))
	}
	if !change {
		return
	}
	var crossResponse = CrossResponse{Ring: c.CrossRing.Ring, CrossInfo: make([]*CrossInfo, 0)}
	for _, cr := range c.CrossRing.CrossInfo {
		crossResponse.CrossInfo = append(crossResponse.CrossInfo, cr)
	}
	sort.Slice(crossResponse.CrossInfo, func(i int, j int) bool {
		return crossResponse.CrossInfo[i].Time.After(crossResponse.CrossInfo[j].Time)
	})

	resp := newAlarmMess(typeRingData, nil)
	resp.Data[typeRingData] = crossResponse
	c.send <- resp
}

//readPump обработчик чтения сокета
func (c *ClientAlarm) readPump(db *sqlx.DB) {

	//если нужно указать лимит пакета
	//c.conn.SetReadLimit(maxMessageSize)

	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	{
		c.makeResponse()
	}
}

//writePump обработчик записи в сокет
func (c *ClientAlarm) writePump() {
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
