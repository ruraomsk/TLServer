package greenStreet

import (
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/maps"
	"github.com/ruraomsk/TLServer/internal/sockets/maps/mainMap"
	"github.com/ruraomsk/ag-server/comm"
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

//Run запуск хаба для GS
func (h *HubGStreet) Run(db *sqlx.DB) {

	crossReadTick := time.NewTicker(crossPeriod)
	checkValidityTicker := time.NewTicker(checkTokensValidity)
	defer func() {
		crossReadTick.Stop()
		checkValidityTicker.Stop()
	}()

	oldTFs := maps.SelectTL(db)
	for {
		select {
		case <-crossReadTick.C:
			{
				if len(h.clients) > 0 {
					newTFs := maps.SelectTL(db)
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
							var flagAdd = true
							for _, oTF := range oldTFs {
								if oTF.Idevice == nTF.Idevice {
									flagAdd = false
									if oTF.Sost.Num != nTF.Sost.Num || oTF.Description != nTF.Description {
										flagAdd = true
									}
									if oTF.Subarea != nTF.Subarea {
										flagAdd = true
										flagFill = true
									}
									break
								}
							}
							if flagAdd {
								tempTF = append(tempTF, nTF)
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
			}
		case client := <-h.unregister:
			{
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					_ = client.conn.Close()
				}

				if len(client.devices) != 0 {
					arm := comm.CommandARM{Command: 4, Params: 0, User: client.cInfo.Login}
					var mess = tcpConnect.TCPMessage{
						User:        client.cInfo.Login,
						TCPType:     tcpConnect.TypeDispatch,
						From:        tcpConnect.FromGsSoc,
						CommandType: typeDButton,
						Pos:         sockets.PosInfo{},
					}

					device.GlobalDevEdit.Mux.Lock()
					for _, dev := range client.devices {
						tDev := device.GlobalDevEdit.MapDevices[dev]
						tDev.BusyCount--
						if tDev.BusyCount == 0 && tDev.TurnOnFlag == true {
							arm.ID = dev
							mess.Idevice = arm.ID
							mess.Data = arm
							mess.SendToTCPServer()
							tDev.TurnOnFlag = false
						}
						device.GlobalDevEdit.MapDevices[dev] = tDev
					}
					device.GlobalDevEdit.Mux.Unlock()
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
		case login := <-mainMap.UserLogoutGS:
			{
				resp := newGSMess(typeClose, nil)
				resp.Data["message"] = "пользователь вышел из системы"
				for client := range h.clients {
					if client.cInfo.Login == login {
						client.send <- resp
					}
				}
			}
		case <-checkValidityTicker.C:
			{
				for client := range h.clients {
					if client.cInfo.Valid() != nil {
						msg := newGSMess(typeClose, nil)
						msg.Data["message"] = "вышло время сеанса пользователя"
						client.send <- msg
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
					if client.cInfo.Login == msg.User {
						client.send <- resp
					}
				}
			}
		}
	}
}
