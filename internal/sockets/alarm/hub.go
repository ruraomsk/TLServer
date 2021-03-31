package alarm

import (
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"time"
)

//HubAlarm структура хаба для alarm
type HubAlarm struct {
	clients    map[*ClientAlarm]bool
	broadcast  chan alarmResponse
	register   chan *ClientAlarm
	unregister chan *ClientAlarm
	db         *sqlx.DB
}

//NewAlarmHub создание хаба
func NewAlarmHub() *HubAlarm {
	return &HubAlarm{
		broadcast:  make(chan alarmResponse),
		clients:    make(map[*ClientAlarm]bool),
		register:   make(chan *ClientAlarm),
		unregister: make(chan *ClientAlarm),
	}
}

//Run запуск хаба для techArm
func (h *HubAlarm) Run(db *sqlx.DB) {
	h.db = db
	UserLogoutAlarm = make(chan string)
	sockets.DispatchMessageFromAnotherPlace = make(chan sockets.DBMessage, 50)

	readCrossTick := time.NewTicker(devUpdate)
	checkValidityTicker := time.NewTicker(checkTokensValidity)
	defer func() {
		readCrossTick.Stop()
		checkValidityTicker.Stop()
	}()

	for {
		select {
		case <-readCrossTick.C:
			{
				if len(h.clients) == 0 {
					continue
				}
				for client := range h.clients {
					client.makeResponse()
				}
			}
		case client := <-h.register:
			{
				h.clients[client] = true
			}
		case client := <-h.unregister:
			{
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					_ = client.conn.Close()
				}
			}
		case mess := <-h.broadcast:
			{
				for client := range h.clients {
					select {
					case client.send <- mess:
					default:
						delete(h.clients, client)
						close(client.send)
					}
				}
			}
		case login := <-UserLogoutAlarm:
			{
				resp := newAlarmMess(typeClose, nil)
				resp.Data["message"] = "пользователь вышел из системы"
				for client := range h.clients {
					if client.armInfo.AccInfo.Login == login {
						client.send <- resp
					}
				}
			}
		case <-checkValidityTicker.C:
			{
				for client := range h.clients {
					if client.armInfo.AccInfo.Valid() != nil {
						msg := newAlarmMess(typeClose, nil)
						msg.Data["message"] = "вышло время сеанса пользователя"
						client.send <- msg
					}
				}
			}
		}
	}
}
