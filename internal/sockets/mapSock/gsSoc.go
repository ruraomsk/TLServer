package mapSock

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/routeGS"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/comm"
	"time"
)

var connectOnGS map[*websocket.Conn]string //пулл соединени
var writeGS chan GSSokResponse             //канал для отправки сообщений
//var GSRepaint chan bool
var userLogout chan string

//GSReader обработчик открытия сокета для режима зеленой улицы
func GSReader(conn *websocket.Conn, mapContx map[string]string, db *sqlx.DB) {
	login := mapContx["login"]
	connectOnGS[conn] = login
	//начальная информация
	{
		resp := newGSMess(typeMapInfo, conn, mapOpenInfo(db))
		resp.Data["routes"] = getAllModes(db)
		data.CacheArea.Mux.Lock()
		resp.Data["areaZone"] = data.CacheArea.Areas
		data.CacheArea.Mux.Unlock()
		resp.send()
	}
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//закрытие коннекта
			resp := newGSMess(typeClose, conn, nil)
			resp.send()
			return
		}

		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			resp := newGSMess(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeCreateRout:
			{
				temp := routeGS.Route{}
				_ = json.Unmarshal(p, &temp)
				resp := newGSMess(typeCreateRout, conn, nil)
				err := temp.Create(db)
				if err != nil {
					resp.Data[typeError] = errCantWriteInBD
				} else {
					resp.Data["route"] = temp
				}
				resp.send()
			}
		case typeUpdateRout:
			{
				temp := routeGS.Route{}
				_ = json.Unmarshal(p, &temp)
				resp := newGSMess(typeUpdateRout, conn, nil)
				err := temp.Update(db)
				if err != nil {
					resp.Data[typeError] = errCantWriteInBD
				} else {
					resp.Data["route"] = temp
				}
				resp.send()
			}
		case typeDeleteRout:
			{
				temp := routeGS.Route{}
				_ = json.Unmarshal(p, &temp)
				resp := newGSMess(typeDeleteRout, conn, nil)
				err := temp.Delete(db)
				if err != nil {
					resp.Data[typeError] = errCantWriteInBD
				} else {
					resp.Data["route"] = temp
				}
				resp.send()
			}
		case typeJump: //отправка default
			{
				location := &data.Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := newGSMess(typeJump, conn, nil)
				resp.Data["boxPoint"] = box
				resp.send()
			}
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = login
				var mess = tcpConnect.TCPMessage{
					User:        arm.User,
					TCPType:     tcpConnect.TypeDispatch,
					Idevice:     arm.ID,
					Data:        arm,
					From:        tcpConnect.FromGsSoc,
					CommandType: typeDButton,
					Pos:         sockets.PosInfo{},
				}
				mess.SendToTCPServer()
			}
		}
	}
}

//GSBroadcast передатчик для ЗУ
func GSBroadcast(db *sqlx.DB) {
	connectOnGS = make(map[*websocket.Conn]string)
	writeGS = make(chan GSSokResponse, 50)

	//GSRepaint = make(chan bool)
	userLogout = make(chan string)
	crossReadTick := time.NewTicker(time.Second * 5)
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		crossReadTick.Stop()
	}()
	oldTFs := selectTL(db)
	for {
		select {
		case <-crossReadTick.C:
			{
				if len(connectOnGS) > 0 {
					newTFs := selectTL(db)
					if len(newTFs) != len(oldTFs) {
						resp := newMapMess(typeRepaint, nil, nil)
						resp.Data["tflight"] = newTFs
						data.CacheArea.Mux.Lock()
						resp.Data["areaZone"] = data.CacheArea.Areas
						data.CacheArea.Mux.Unlock()
						for conn := range connectOnGS {
							_ = conn.WriteJSON(resp)
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
							resp := newGSMess(typeTFlight, nil, nil)
							if flagFill {
								data.FillMapAreaZone()
								data.CacheArea.Mux.Lock()
								resp.Data["areaZone"] = data.CacheArea.Areas
								data.CacheArea.Mux.Unlock()
							}
							resp.Data["tflight"] = tempTF
							for conn := range connectOnGS {
								_ = conn.WriteJSON(resp)
							}
						}
					}
					oldTFs = newTFs
				}
			}
		case <-pingTicker.C:
			{
				for conn := range connectOnGS {
					_ = conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		case login := <-userLogout:
			{
				for conn, user := range connectOnGS {
					if user == login {
						msg := newGSMess(typeClose, nil, nil)
						msg.Data["message"] = "пользователь вышел из системы"
						_ = conn.WriteJSON(msg)
						//_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "пользователь вышел из системы"))
					}
				}
			}
		case msg := <-tcpConnect.TCPRespGS:
			{
				resp := newGSMess(typeDButton, nil, nil)
				resp.Data["status"] = msg.Status
				if msg.Status {
					resp.Data["command"] = msg.Data
					var message = sockets.DBMessage{Data: resp, Idevice: msg.Idevice}
					sockets.DispatchMessageFromAnotherPlace <- message
				}
				for conn, user := range connectOnGS {
					if user == msg.User {
						_ = conn.WriteJSON(resp)
					}
				}
			}
		case msg := <-writeGS:
			switch msg.Type {
			case typeClose:
				{
					delete(connectOnGS, msg.conn)
				}
			case typeCreateRout:
				{
					sendCDU(msg)
				}
			case typeDeleteRout:
				{
					sendCDU(msg)
				}
			case typeUpdateRout:
				{
					sendCDU(msg)
				}
			default:
				{
					_ = msg.conn.WriteJSON(msg)
				}
			}
		}
	}
}

//sendCDU ответ на создание, удаление и обновление
func sendCDU(msg GSSokResponse) {
	if _, ok := msg.Data[typeError]; !ok {
		for conn := range connectOnGS {
			_ = conn.WriteJSON(msg)
		}
	} else {
		_ = msg.conn.WriteJSON(msg)
	}
}
