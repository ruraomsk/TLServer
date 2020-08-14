package controlCross

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/internal/sockets/crossSock/mainCross"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

//HubCross структура хаба для cross
type HubControlCross struct {
	clients    map[*ClientControlCr]bool
	broadcast  chan ControlSokResponse
	register   chan *ClientControlCr
	unregister chan *ClientControlCr
}

//NewCrossHub создание хаба
func NewCrossHub() *HubControlCross {
	return &HubControlCross{
		broadcast:  make(chan ControlSokResponse),
		clients:    make(map[*ClientControlCr]bool),
		register:   make(chan *ClientControlCr),
		unregister: make(chan *ClientControlCr),
	}
}

//Run запуск хаба для xctrl
func (h *HubControlCross) Run(db *sqlx.DB) {
	crossSock.GetArmUsersForDisplay = make(chan bool)
	crossSock.CrArmUsersForDisplay = make(chan []crossSock.CrossInfo)
	crossSock.DiscArmUsers = make(chan []crossSock.CrossInfo)
	UserLogoutCrControl = make(chan string)

	for {
		select {
		case client := <-h.register:
			{
				var regStatus = true
				//проверка на существование такого перекрестка (сбос если нету)
				_, err := crossSock.GetNewState(client.crossInfo.Pos, db)
				if err != nil {
					close(client.send)
					_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errCrossDoesntExist))
					_ = client.conn.Close()
					regStatus = false
					client.regStatus <- regStatus
					continue
				}

				//проверка открыт ли у этого пользователя такой перекресток
				for hubClient := range h.clients {
					if client.crossInfo.Pos == hubClient.crossInfo.Pos && client.crossInfo.Login == hubClient.crossInfo.Login {
						close(client.send)
						_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errDoubleOpeningDevice))
						_ = client.conn.Close()
						regStatus = false
						client.regStatus <- regStatus
						break
					}
				}

				if !regStatus {
					continue
				}
				//кромешный пи**** с созданием нормального клиента
				resp, Idevice, description := takeControlInfo(client.crossInfo.Pos, db)
				client.crossInfo.Idevice = Idevice
				client.crossInfo.Description = description
				data.CacheInfo.Mux.Lock()
				resp.Data["areaMap"] = data.CacheInfo.MapArea[data.CacheInfo.MapRegion[client.crossInfo.Pos.Region]]
				data.CacheInfo.Mux.Unlock()

				//если роль пришедшего Viewer то влаг ему не ставим
				flagEdit := false
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

				client.regStatus <- regStatus

				h.clients[client] = true
				//отправим собранные данные клиенту
				client.send <- resp

				fmt.Printf("Cross ARM reg: ")
				for client := range h.clients {
					fmt.Printf("%v ", client.crossInfo)
				}
				fmt.Printf("\n")
			}
		case client := <-h.unregister:
			{
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					_ = client.conn.Close()
					if client.crossInfo.Edit {
						{
							for aClient := range h.clients {
								if (aClient.crossInfo.Pos == client.crossInfo.Pos) && (aClient.crossInfo.Role != "Viewer") {
									//delete(h.clients, aClient)
									aClient.crossInfo.Edit = true
									//h.clients[aClient] = true
									resp := newControlMess(typeChangeEdit, nil)
									resp.Data["edit"] = true
									aClient.send <- resp
									break
								}
							}
						}
					}
				}

				fmt.Printf("Cross ARM UnReg: ")
				for client := range h.clients {
					fmt.Printf("%v ", client.crossInfo)
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
		case <-crossSock.GetArmUsersForDisplay:
			{
				crossSock.CrArmUsersForDisplay <- h.usersList()
			}
		case dArmInfo := <-crossSock.DiscArmUsers:
			{
				for _, dArm := range dArmInfo {
					for client := range h.clients {
						if client.crossInfo.Pos == dArm.Pos && client.crossInfo.Login == dArm.Login {
							msg := newControlMess(typeClose, nil)
							msg.Data["message"] = "закрытие администратором"
							client.send <- msg
						}
					}
				}
			}
		case login := <-UserLogoutCrControl:
			{
				for client := range h.clients {
					if client.crossInfo.Login == login {
						msg := newControlMess(typeClose, nil)
						msg.Data["message"] = "пользователь вышел из системы"
						client.send <- msg
					}
				}
			}
		case msg := <-tcpConnect.TCPRespCrControlSoc:
			{
				resp := newControlMess("", nil)
				switch msg.CommandType {
				case typeDButton:
					{
						resp.Type = typeDButton
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
				case typeSendB:
					{
						resp.Type = typeSendB
						resp.Data["status"] = msg.Status
						if msg.Status {
							var uState agspudge.UserCross
							raw, _ := json.Marshal(msg.Data)
							_ = json.Unmarshal(raw, &uState)
							resp.Data["state"] = uState.State
							resp.Data["user"] = msg.User
						}
						if msg.Status {
							//если есть поле отправить всем кто слушает
							for client := range h.clients {
								if client.crossInfo.Pos == msg.Pos {
									client.send <- resp
								}
							}
							mainCross.ChangeState <- msg
						} else {
							// если нету поля отправить ошибку только пользователю
							for client := range h.clients {
								if client.crossInfo.Login == msg.User && client.crossInfo.Pos == msg.Pos {
									client.send <- resp
								}
							}
						}
					}
				case typeCreateB:
					{
						resp.Type = typeCreateB
						resp.Data["status"] = msg.Status
						for client := range h.clients {
							if client.crossInfo.Login == msg.User && client.crossInfo.Pos == msg.Pos {
								client.send <- resp
							}
						}

					}
				case typeDeleteB:
					{
						resp.Type = typeDeleteB
						resp.Data["status"] = msg.Status
						if msg.Status {
							//если есть поле отправить всем кто слушает
							for client := range h.clients {
								if client.crossInfo.Pos == msg.Pos {
									client.send <- resp
								}
							}
							mainCross.ArmDeleted <- msg
						} else {
							// если нету поля отправить ошибку только пользователю
							for client := range h.clients {
								if client.crossInfo.Login == msg.User && client.crossInfo.Pos == msg.Pos {
									client.send <- resp
								}
							}
						}
					}
				}
			}
		}
	}
}

func (h *HubControlCross) usersList() []crossSock.CrossInfo {
	var temp = make([]crossSock.CrossInfo, 0)
	for client := range h.clients {
		if client.crossInfo.Edit {
			temp = append(temp, client.crossInfo)
		}
	}
	return temp
}
