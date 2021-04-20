package mainCross

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/crossSock"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	agspudge "github.com/ruraomsk/ag-server/pudge"
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

var ChangeState chan tcpConnect.TCPMessage
var GetCrossUserForMap chan bool
var UserLogoutCross chan string
var CrossUsersForMap chan []crossSock.CrossInfo
var ArmDeleted chan tcpConnect.TCPMessage

//Run запуск хаба для mainCross
func (h *HubCross) Run() {
	ChangeState = make(chan tcpConnect.TCPMessage)
	GetCrossUserForMap = make(chan bool)
	UserLogoutCross = make(chan string)
	crossSock.CrossUsersForDisplay = make(chan []crossSock.CrossInfo)
	CrossUsersForMap = make(chan []crossSock.CrossInfo)
	crossSock.GetCrossUsersForDisplay = make(chan bool)
	crossSock.DiscCrossUsers = make(chan []crossSock.CrossInfo)
	ArmDeleted = make(chan tcpConnect.TCPMessage)

	type crossUpdateInfo struct {
		Idevice  int             `json:"idevice"`
		Status   data.TLSostInfo `json:"status"`
		State    agspudge.Cross  `json:"state"`
		stateStr string
	}
	globArrCross := make(map[int]crossUpdateInfo)
	globArrDevice := make(map[int]device.DevInfo)

	updateTicker := time.NewTicker(readCrossTick)
	checkValidityTicker := time.NewTicker(checkTokensValidity)
	defer func() {
		updateTicker.Stop()
		checkValidityTicker.Stop()
	}()

	for {
		select {
		case <-updateTicker.C:
			{
				if len(h.clients) > 0 {
					aPos := make([]int, 0)
					arrayCross := make(map[int]crossUpdateInfo)
					arrayDevice := make(map[int]device.DevInfo)
					for client := range h.clients {
						if len(aPos) == 0 {
							aPos = append(aPos, client.crossInfo.Idevice)
							continue
						}
						for _, a := range aPos {
							if a == client.crossInfo.Idevice {
								break
							}
							aPos = append(aPos, client.crossInfo.Idevice)
						}
					}
					//выполняем если хоть что-то есть
					if len(aPos) > 0 {
						//запрос статуса и state
						query, args, err := sqlx.In("SELECT idevice, status, state FROM public.cross WHERE idevice IN (?)", aPos)
						if err != nil {
							logger.Error.Println("|Message: cross socket cant make IN ", err.Error())
							continue
						}
						db, id := data.GetDBX()
						query = db.Rebind(query)
						rows, err := db.Queryx(query, args...)
						if err != nil {
							logger.Error.Println("|Message: db not respond", err.Error())
							continue
						}
						for rows.Next() {
							var tempCR crossUpdateInfo
							_ = rows.Scan(&tempCR.Idevice, &tempCR.Status.Num, &tempCR.stateStr)
							data.CacheInfo.Mux.Lock()
							tempCR.Status.Description = data.CacheInfo.MapTLSost[tempCR.Status.Num].Description
							tempCR.Status.Control = data.CacheInfo.MapTLSost[tempCR.Status.Num].Control
							data.CacheInfo.Mux.Unlock()
							tempCR.State, _ = crossSock.ConvertStateStrToStruct(tempCR.stateStr)
							arrayCross[tempCR.Idevice] = tempCR
						}
						data.FreeDBX(id)
						for idevice, newData := range arrayCross {
							if oldData, ok := globArrCross[idevice]; ok {
								//если запись есть нужно сравнить и если есть разница отправить изменения
								if oldData.State.PK != newData.State.PK || oldData.State.NK != newData.State.NK || oldData.State.CK != newData.State.CK || oldData.Status.Num != newData.Status.Num {
									for client := range h.clients {
										if client.crossInfo.Idevice == newData.Idevice {
											msg := newCrossMess(typeCrossUpdate, nil)
											msg.Data["idevice"] = newData.Idevice
											msg.Data["status"] = newData.Status
											msg.Data["state"] = newData.State
											client.send <- msg
										}
									}
								}
							} else {
								//если не существует старой записи ее нужно отправить
								for client := range h.clients {
									if client.crossInfo.Idevice == newData.Idevice {
										msg := newCrossMess(typeCrossUpdate, nil)
										msg.Data["idevice"] = newData.Idevice
										msg.Data["status"] = newData.Status
										msg.Data["state"] = newData.State
										client.send <- msg
									}
								}
							}
						}

						device.GlobalDevices.Mux.Lock()
						for key, c := range device.GlobalDevices.MapDevices {
							arrayDevice[key] = c
						}
						device.GlobalDevices.Mux.Unlock()

						for idevice, newDev := range arrayDevice {
							if oldDev, ok := globArrDevice[idevice]; ok {
								if oldDev.Controller.StatusConnection != newDev.Controller.StatusConnection || oldDev.Controller.Status.Ethernet != newDev.Controller.Status.Ethernet {
									for client := range h.clients {
										if client.crossInfo.Idevice == newDev.Controller.ID {
											msg := newCrossMess(typeCrossConnection, nil)
											msg.Data["scon"] = newDev.Controller.StatusConnection
											msg.Data["eth"] = newDev.Controller.Status.Ethernet
											client.send <- msg
										}
									}
								}
							} else {
								for client := range h.clients {
									if client.crossInfo.Idevice == newDev.Controller.ID {
										msg := newCrossMess(typeCrossConnection, nil)
										msg.Data["scon"] = newDev.Controller.StatusConnection
										msg.Data["eth"] = newDev.Controller.Status.Ethernet
										client.send <- msg
									}
								}
							}
						}

						globArrCross = arrayCross
						globArrDevice = arrayDevice
					}
				}
			}
		case dkInfo := <-tcpConnect.ChanChangeDk:
			{
				for client := range h.clients {
					if client.crossInfo.Idevice == dkInfo.Idevice {
						msg := newCrossMess(typePhase, nil)
						msg.Data["idevice"] = dkInfo.Idevice
						msg.Data["dk"] = dkInfo.DK
						client.send <- msg
					}
				}
			}
		case client := <-h.register:
			{
				var regStatus = true
				//проверка на существование такого перекрестка (сбос если нету)
				_, err := crossSock.GetNewState(client.crossInfo.Pos)
				if err != nil {
					close(client.send)
					_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errCrossDoesntExist))
					_ = client.conn.Close()
					regStatus = false
					client.regStatus <- regStatus
					continue
				}

				//проверка открыт ли у этого пользователя такой перекресток
				//for hubClient := range h.clients {
				//	if client.crossInfo.Pos == hubClient.crossInfo.Pos && client.crossInfo.AccInfo.Login == hubClient.crossInfo.AccInfo.Login {
				//		close(client.send)
				//		_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errDoubleOpeningDevice))
				//		_ = client.conn.Close()
				//		regStatus = false
				//		client.regStatus <- regStatus
				//		break
				//	}
				//}
				//if !regStatus {
				//	continue
				//}
				//кромешный пи**** с созданием нормального клиента
				resp, Idevice, description := takeCrossInfo(client.crossInfo.Pos)
				client.crossInfo.Idevice = Idevice
				client.crossInfo.Description = description
				resp.Data["access"] = false
				if (fmt.Sprint(resp.Data["region"]) == client.crossInfo.AccInfo.Region) || (client.crossInfo.AccInfo.Region == "*") {
					resp.Data["access"] = data.AccessCheck(client.crossInfo.AccInfo.Login, 4)
				}
				delete(resp.Data, "region")

				//если роль пришедшего Viewer то влаг ему не ставим
				flagEdit := false
				if client.crossInfo.AccInfo.Role != "Viewer" {
					for hClient := range h.clients {
						if hClient.crossInfo.Pos == client.crossInfo.Pos && hClient.crossInfo.Edit {
							flagEdit = true
							break
						}
					}
					if !flagEdit {
						resp.Data["edit"] = true
						client.crossInfo.Edit = true
					} else {
						resp.Data["edit"] = false
					}
					//если есть полномочия запишим что он на перекрестке
					device.GlobalDevEdit.Mux.Lock()
					tDev := device.GlobalDevEdit.MapDevices[client.crossInfo.Idevice]
					arm := comm.CommandARM{
						User:    client.crossInfo.Login,
						ID:      Idevice,
						Command: 4,
						Params:  1,
					}
					var mess = tcpConnect.TCPMessage{
						User:        client.crossInfo.AccInfo.Login,
						TCPType:     tcpConnect.TypeDispatch,
						Idevice:     arm.ID,
						Data:        arm,
						From:        tcpConnect.FromCrossSoc,
						CommandType: typeDButton,
						Pos:         client.crossInfo.Pos,
					}
					mess.SendToTCPServer()

					tDev.TurnOnFlag = true
					tDev.BusyCount++

					device.GlobalDevEdit.MapDevices[client.crossInfo.Idevice] = tDev
					device.GlobalDevEdit.Mux.Unlock()
				}

				client.regStatus <- regStatus

				h.clients[client] = true
				//отправим собранные данные клиенту
				client.send <- resp

				//если пользователь занял светофор на управление отправить на карту список всех управляемых светофоров
				if client.crossInfo.Edit {
					CrossUsersForMap <- h.usersList()
				}
			}
		case client := <-h.unregister:
			{
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					_ = client.conn.Close()
					//if client.crossInfo.Edit {

					//если есть полномочия запишим что он на перекрестке
					device.GlobalDevEdit.Mux.Lock()
					tDev := device.GlobalDevEdit.MapDevices[client.crossInfo.Idevice]
					tDev.BusyCount--
					//if tDev.BusyCount == 0 && tDev.TurnOnFlag == true {
					arm := comm.CommandARM{
						User:    client.crossInfo.Login,
						ID:      client.crossInfo.Idevice,
						Command: 4,
						Params:  0,
					}
					var mess = tcpConnect.TCPMessage{
						User:        client.crossInfo.AccInfo.Login,
						TCPType:     tcpConnect.TypeDispatch,
						Idevice:     arm.ID,
						Data:        arm,
						From:        tcpConnect.FromCrossSoc,
						CommandType: typeDButton,
						Pos:         client.crossInfo.Pos,
					}
					mess.SendToTCPServer()

					//	tDev.TurnOnFlag = false
					//}
					device.GlobalDevEdit.MapDevices[client.crossInfo.Idevice] = tDev
					device.GlobalDevEdit.Mux.Unlock()

					for aClient := range h.clients {
						if (aClient.crossInfo.Pos == client.crossInfo.Pos) && (aClient.crossInfo.AccInfo.Role != "Viewer") {
							aClient.crossInfo.Edit = true
							resp := newCrossMess(typeChangeEdit, nil)
							resp.Data["edit"] = true
							aClient.send <- resp
							break
						}
					}
					//отправить на мапу подключенные устройства которые редактируют
					CrossUsersForMap <- h.usersList()

					//}
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
		//отправить на мапу подключенные устройства которые редактируют
		case <-GetCrossUserForMap:
			{
				CrossUsersForMap <- h.usersList()
			}
		case <-crossSock.GetCrossUsersForDisplay:
			{
				crossSock.CrossUsersForDisplay <- h.usersList()
			}
		case dCrInfo := <-crossSock.DiscCrossUsers: //ok
			{
				for _, dCr := range dCrInfo {
					for client := range h.clients {
						if client.crossInfo.Pos == dCr.Pos && client.crossInfo.Login == dCr.Login {
							msg := newCrossMess(typeClose, nil)
							msg.Data["message"] = "закрытие администратором"
							client.send <- msg
						}
					}
				}
			}
		case msgD := <-ArmDeleted: //ok
			{
				for client := range h.clients {
					if client.crossInfo.Pos == msgD.Pos {
						msg := newCrossMess(typeClose, nil)
						msg.Data["message"] = "перекресток удален"
						client.send <- msg
					}
				}
			}
		case login := <-UserLogoutCross:
			{
				for client := range h.clients {
					if client.crossInfo.AccInfo.Login == login {
						msg := newCrossMess(typeClose, nil)
						msg.Data["message"] = "пользователь вышел из системы"
						client.send <- msg
					}
				}
			}
		case <-checkValidityTicker.C:
			{
				for client := range h.clients {
					if client.crossInfo.AccInfo.Valid() != nil {
						msg := newCrossMess(typeClose, nil)
						msg.Data["message"] = "вышло время сеанса пользователя"
						client.send <- msg
					}
				}
			}
		case msg := <-ChangeState: //ok
			{
				resp := newCrossMess(typeStateChange, nil)
				var uState agspudge.UserCross
				raw, _ := json.Marshal(msg.Data)
				_ = json.Unmarshal(raw, &uState)
				resp.Data["state"] = uState.State
				resp.Data["user"] = msg.User
				for client := range h.clients {
					if client.crossInfo.Pos == msg.Pos {
						client.send <- resp
					}
				}
			}
		case msg := <-sockets.DispatchMessageFromAnotherPlace:
			{
				for client := range h.clients {
					if client.crossInfo.Idevice == msg.Idevice {
						//так сформировано (п.с. ну а че...)
						_ = client.conn.WriteJSON(msg.Data)
					}
				}
			}
		case msg := <-tcpConnect.TCPRespCrossSoc:
			{
				resp := newCrossMess(typeDButton, nil)
				resp.Data["status"] = msg.Status
				if msg.Status {
					resp.Data["command"] = msg.Data
				}
				for client := range h.clients {
					if client.crossInfo.Idevice == msg.Idevice {
						client.send <- resp
					}
				}
			}
		}
	}
}

func (h *HubCross) usersList() []crossSock.CrossInfo {
	var temp = make([]crossSock.CrossInfo, 0)
	for client := range h.clients {
		if client.crossInfo.Edit {
			temp = append(temp, *client.crossInfo)
		}
	}
	return temp
}
