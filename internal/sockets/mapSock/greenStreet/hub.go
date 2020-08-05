package greenStreet

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/mapSock"
	"github.com/jmoiron/sqlx"
	"time"
)

//HubGStreet структура хаба для GStreet
type HubGStreet struct {
	clients    map[*ClientGS]bool
	broadcast  chan gSResponse
	register   chan *ClientGS
	unregister chan *ClientGS
}

//NewGSHub создание хаба
func NewGSHub() *HubGStreet {
	return &HubGStreet{
		broadcast:  make(chan gSResponse),
		clients:    make(map[*ClientGS]bool),
		register:   make(chan *ClientGS),
		unregister: make(chan *ClientGS),
	}
}

//Run запуск хаба для xctrl
func (h *HubGStreet) Run(db *sqlx.DB) {

	crossReadTick := time.NewTicker(crossPeriod)
	defer crossReadTick.Stop()

	oldTFs := mapSock.SelectTL(db)
	for {
		select {
		case <-crossReadTick.C:
			{
				if len(h.clients) > 0 {
					newTFs := mapSock.SelectTL(db)
					if len(newTFs) != len(oldTFs) {
						resp := newGSMess(typeRepaint, nil)
						resp.Data["tflight"] = newTFs
						data.CacheArea.Mux.Lock()
						resp.Data["areaZone"] = data.CacheArea.Areas
						data.CacheArea.Mux.Unlock()
						for client := range h.clients {
							client.send <- resp
						}
					} else {
						var (
							tempTF   []data.TrafficLights
							flagFill = false
						)
						for _, nTF := range newTFs {
							for _, oTF := range oldTFs {
								if oTF.Idevice == nTF.Idevice {
									var flagAdd = false
									if oTF.Sost.Num != nTF.Sost.Num {
										flagAdd = true
									}
									if oTF.Subarea != nTF.Subarea {
										flagAdd = true
										flagFill = true
									}
									if flagAdd {
										tempTF = append(tempTF, nTF)
										break
									}
								}
							}
						}
						if len(tempTF) > 0 {
							resp := newGSMess(typeTFlight, nil)
							if flagFill {
								data.FillMapAreaZone()
								data.CacheArea.Mux.Lock()
								resp.Data["areaZone"] = data.CacheArea.Areas
								data.CacheArea.Mux.Unlock()
							}
							resp.Data["tflight"] = tempTF
							for client := range h.clients {
								client.send <- resp
							}
						}
					}
					oldTFs = newTFs
				}
			}
		case client := <-h.register:
			{
				h.clients[client] = true

				fmt.Printf("gStreet reg: ")
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

				fmt.Println("gStreet unReg: ")
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
		case login := <-mapSock.UserLogoutGS:
			{
				resp := newGSMess(typeClose, nil)
				resp.Data["message"] = "пользователь вышел из системы"
				for client := range h.clients {
					if client.login == login {
						client.send <- resp
					}
				}
			}
		case msg := <-tcpConnect.TCPRespGS:
			{
				resp := newGSMess(typeDButton, nil)
				resp.Data["status"] = msg.Status
				if msg.Status {
					resp.Data["command"] = msg.Data
					var message = sockets.DBMessage{Data: resp, Idevice: msg.Idevice}
					sockets.DispatchMessageFromAnotherPlace <- message
				}
				for client := range h.clients {
					if client.login == msg.User {
						client.send <- resp
					}
				}
			}
		}
	}
}
