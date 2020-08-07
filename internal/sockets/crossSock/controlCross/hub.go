package controlCross

import (
	"github.com/jmoiron/sqlx"
	"time"
)

//HubCross структура хаба для cross
type HubCross struct {
	clients    map[*ClientCross]bool
	broadcast  chan crossResponse
	register   chan *ClientCross
	unregister chan *ClientCross
}

//NewCrossHub создание хаба
func NewCrossHub() *HubCross {
	return &HubCross{
		broadcast:  make(chan crossResponse),
		clients:    make(map[*ClientCross]bool),
		register:   make(chan *ClientCross),
		unregister: make(chan *ClientCross),
	}
}

//Run запуск хаба для xctrl
func (h *HubCross) Run(db *sqlx.DB) {

	updateTicker := time.NewTicker(time.Second * 20)
	defer updateTicker.Stop()

	for {
		select {
		case <-updateTicker.C:
			{

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
		}
	}
}
