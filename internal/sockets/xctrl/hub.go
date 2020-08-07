package xctrl

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/xcontrol"
	"reflect"
	"time"
)

//HubXctrl структура хаба для xctrl
type HubXctrl struct {
	clients    map[*ClientXctrl]bool
	broadcast  chan MessXctrl
	register   chan *ClientXctrl
	unregister chan *ClientXctrl
}

//NewXctrlHub создание хаба
func NewXctrlHub() *HubXctrl {
	return &HubXctrl{
		broadcast:  make(chan MessXctrl),
		clients:    make(map[*ClientXctrl]bool),
		register:   make(chan *ClientXctrl),
		unregister: make(chan *ClientXctrl),
	}
}

//Run запуск хаба для xctrl
func (h *HubXctrl) Run(db *sqlx.DB) {
	UserLogoutXctrl = make(chan string)

	updateTicker := time.NewTicker(stateTime)
	defer updateTicker.Stop()
	oldXctrl, _ := getXctrl(db)

	for {
		select {
		case <-updateTicker.C:
			{
				if len(h.clients) > 0 {
					newXctrl, _ := getXctrl(db)
					var tempXctrl []xcontrol.State
					for _, nX := range newXctrl {
						flagNew := true
						for _, oX := range oldXctrl {
							if oX.Region == nX.Region && oX.Area == nX.Area && oX.SubArea == nX.SubArea {
								flagNew = false
								if !reflect.DeepEqual(nX.Calculates, oX.Calculates) ||
									!reflect.DeepEqual(nX.Status, oX.Status) ||
									!reflect.DeepEqual(nX.Strategys, oX.Strategys) ||
									!reflect.DeepEqual(nX.XNumber, oX.XNumber) ||
									!reflect.DeepEqual(nX.PKLast, oX.PKLast) ||
									!reflect.DeepEqual(nX.PKNow, oX.PKNow) ||
									!reflect.DeepEqual(nX.LastTime, oX.LastTime) ||
									!reflect.DeepEqual(nX.Switch, oX.Switch) ||
									!reflect.DeepEqual(nX.Release, oX.Release) ||
									!reflect.DeepEqual(nX.Step, oX.Step) {
									tempXctrl = append(tempXctrl, nX)
								}
								break
							}
						}
						if flagNew {
							tempXctrl = append(tempXctrl, nX)
						}
					}
					oldXctrl = newXctrl
					if len(tempXctrl) > 0 {
						for client := range h.clients {
							resp := newXctrlMess(typeXctrlUpdate, nil)
							resp.Data[typeXctrlUpdate] = tempXctrl
							client.send <- resp
						}
					}
				}
			}
		case client := <-h.register:
			{
				h.clients[client] = true

				fmt.Printf("xctrl reg: ")
				for client := range h.clients {
					fmt.Printf("%v ", client.login)
				}
				fmt.Printf("\n")
			}
		case client := <-h.unregister:
			{
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					_ = client.conn.Close()
				}

				fmt.Printf("xctrl unReg: ")
				for client := range h.clients {
					fmt.Printf("%v ", client.login)
				}
				fmt.Printf("\n")
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
		case login := <-UserLogoutXctrl:
			{
				resp := newXctrlMess(typeClose, nil)
				resp.Data["message"] = "пользователь вышел из системы"
				for client := range h.clients {
					if client.login == login {
						client.send <- resp
					}
				}
			}
		}
	}
}