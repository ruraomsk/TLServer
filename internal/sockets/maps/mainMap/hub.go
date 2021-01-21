package mainMap

import (
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/accToken"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/internal/sockets/crossSock/mainCross"
	"github.com/JanFant/TLServer/internal/sockets/maps"
	"github.com/jmoiron/sqlx"
	"time"
)

//HubMainMap структура хаба для mainMap
type HubMainMap struct {
	clients    map[*ClientMainMap]bool //карта клиентов
	broadcast  chan mapResponse        //струтура сообщения mainMap
	register   chan *ClientMainMap     //канал для регистрации пользователя
	unregister chan *ClientMainMap     //канал для удаления пользователя
}

//NewMainMapHub создание хаба
func NewMainMapHub() *HubMainMap {
	return &HubMainMap{
		broadcast:  make(chan mapResponse),
		clients:    make(map[*ClientMainMap]bool),
		register:   make(chan *ClientMainMap),
		unregister: make(chan *ClientMainMap),
	}
}

//Run запуск хаба для mainMap
func (h *HubMainMap) Run(db *sqlx.DB) {
	UserLogoutGS = make(chan string, 5)
	data.AccAction = make(chan string, 50)
	license.LogOutAllFromLicense = make(chan bool)
	checkValidityTicker := time.NewTicker(checkTokensValidity)
	crossReadTick := time.NewTicker(crossTick)
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
						resp := newMapMess(typeRepaint, nil)
						resp.Data["tflight"] = newTFs
						data.FillMapAreaZone()
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
									if oTF.Sost.Num != nTF.Sost.Num || oTF.Description != nTF.Description || oTF.Points != nTF.Points {
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
							resp := newMapMess(typeTFlight, nil)
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
				{
					flag, tk := checkToken(client.cookie, client.cInfo.IP, db)
					resp := newMapMess(typeMapInfo, maps.MapOpenInfo(db))
					if flag {
						resp.Data["role"] = tk.Role
						resp.Data["access"] = data.AccessCheck(tk.Login, 2, 5, 6, 7, 8, 9)
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
						client.cInfo = tk
					}
					client.send <- resp
				}

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
		case crossUsers := <-mainCross.CrossUsersForMap:
			{
				resp := newMapMess(typeEditCrossUsers, nil)
				resp.Data["editCrossUsers"] = crossUsers
				for client := range h.clients {
					client.send <- resp
				}
			}
		case msg := <-tcpConnect.TCPRespMap:
			{
				resp := newMapMess(typeCheckConn, nil)
				resp.Data["statusS"] = msg.Status
				for client := range h.clients {
					client.send <- resp
				}
			}
		case login := <-data.AccAction:
			{
				respLO := newMapMess(typeLogOut, nil)
				status := logOut(login, db)
				if status {
					respLO.Data["authorizedFlag"] = false
				}
				for client := range h.clients {
					if client.cInfo.Login == login {
						client.send <- respLO
					}
				}
				logOutSockets(login)
			}
		case <-checkValidityTicker.C:
			{
				for client := range h.clients {
					if client.cookie != "" {
						if client.cInfo.Valid() != nil {
							resp := newMapMess(typeLogOut, nil)
							status := logOut(client.cInfo.Login, db)
							if status {
								resp.Data["authorizedFlag"] = false
							}
							client.cInfo = new(accToken.Token)
							client.cookie = ""
							client.send <- resp
						}
					}
				}
			}
		case <-license.LogOutAllFromLicense:
			{
				for client := range h.clients {
					data.AccAction <- client.cInfo.Login
				}
			}
		}
	}
}
