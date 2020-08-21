package mainMap

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets/crossSock/mainCross"
	"github.com/JanFant/TLServer/internal/sockets/maps"
	"github.com/jmoiron/sqlx"
	"time"
)

//HubMainMap структура хаба для xctrl
type HubMainMap struct {
	clients    map[*ClientMainMap]bool
	broadcast  chan mapResponse
	register   chan *ClientMainMap
	unregister chan *ClientMainMap
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

//Run запуск хаба для xctrl
func (h *HubMainMap) Run(db *sqlx.DB) {

	UserLogoutGS = make(chan string)
	data.AccAction = make(chan string)
	crossReadTick := time.NewTicker(crossTick)
	defer crossReadTick.Stop()

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
							for _, oTF := range oldTFs {
								if oTF.Idevice == nTF.Idevice && oTF.Description != nTF.Description {
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
					flag, tk := checkToken(client.cInfo.tokenStr, client.cInfo.ip, db)
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
						client.cInfo.login = login
					}
					client.send <- resp
				}

				h.clients[client] = true

				fmt.Printf("mainMap reg: ")
				for hClient := range h.clients {
					fmt.Printf("%v ", hClient.cInfo.login)
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

				fmt.Printf("mainMap unReg: ")
				for hClient := range h.clients {
					fmt.Printf("%v ", hClient.cInfo.login)
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
					if client.cInfo.login == login {
						client.send <- respLO
					}
				}
				logOutSockets(login)
			}
		}
	}
}
