package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/comm"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

var writeCrossMessage chan CrossSokResponse
var crossConnect map[*websocket.Conn]crossInfo
var changeState chan PosInfo

//CrossReader обработчик открытия сокета для перекрестка
func CrossReader(conn *websocket.Conn, pos PosInfo, mapContx map[string]string) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var crossCI = crossInfo{Login: mapContx["login"], Pos: pos, Edit: false}

	//проверка не существование такого перекрестка (сбос если нету)
	_, err := getNewState(pos)
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errCrossDoesntExist))
		return
	}

	//проверка открыт ли у этого пользователя такой перекресток
	for _, info := range crossConnect {
		if info.Pos == pos && info.Login == crossCI.Login {
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errDoubleOpeningDevice))
			return
		}
	}

	//флаг редактирования перекрестка
	flagEdit := false
	for _, info := range crossConnect {
		if crossCI.Pos == info.Pos && info.Edit {
			flagEdit = true
			break
		}
	}
	if !flagEdit {
		crossCI.Edit = true
	}

	//сборка начальной информации для отображения перекрестка
	{
		resp := newCrossMess(typeCrossBuild, conn, nil, crossCI)
		resp, crossCI.Idevice, crossCI.Description = takeCrossInfo(crossCI.Pos)
		resp.conn = conn
		resp.Data["edit"] = crossCI.Edit
		resp.Data["controlCrossFlag"] = false
		controlCrossFlag, _ := AccessCheck(crossCI.Login, mapContx["role"], 5)
		if (fmt.Sprint(resp.Data["region"]) == mapContx["region"]) || (mapContx["region"] == "*") {
			resp.Data["controlCrossFlag"] = controlCrossFlag
		}
		delete(resp.Data, "region")
		resp.send()
	}

	//добавление в пул перекрестка
	crossConnect[conn] = crossCI

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//проверка редактирования
			if crossConnect[conn].Edit {
				resp := newCrossMess(typeChangeEdit, conn, nil, crossCI)
				resp.send()
			} else {
				resp := newCrossMess(typeClose, conn, nil, crossCI)
				resp.send()
			}
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := newCrossMess(typeError, conn, nil, crossCI)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeDButton:
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = crossCI.Login
				resp := dispatchControl(arm)
				resp.info = crossCI
				resp.conn = conn
				resp.send()
			}
		}
	}
}

//CrossBroadcast передатчик для перекрестка (cross)
func CrossBroadcast() {
	writeCrossMessage = make(chan CrossSokResponse)
	crossConnect = make(map[*websocket.Conn]crossInfo)
	changeState = make(chan PosInfo)

	type crossUpdateInfo struct {
		Idevice  int            `json:"idevice"`
		Status   TLSostInfo     `json:"status"`
		State    agspudge.Cross `json:"state"`
		stateStr string
	}

	globArrCross := make(map[int]crossUpdateInfo)
	globArrPhase := make(map[int]phaseInfo)
	readTick := time.Tick(time.Second * 1)
	for {
		select {
		case <-readTick:
			{
				if len(crossConnect) > 0 {
					aPos := make([]int, 0)
					arrayCross := make(map[int]crossUpdateInfo)
					arrayPhase := make(map[int]phaseInfo)
					for _, crInfo := range crossConnect {
						if len(aPos) == 0 {
							aPos = append(aPos, crInfo.Idevice)
							continue
						}
						for _, a := range aPos {
							if a == crInfo.Idevice {
								break
							}
							aPos = append(aPos, crInfo.Idevice)
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
						query = GetDB().Rebind(query)
						rows, err := GetDB().Queryx(query, args...)
						if err != nil {
							logger.Error.Println("|Message: db not respond", err.Error())
							continue
						}
						for rows.Next() {
							var tempCR crossUpdateInfo
							_ = rows.Scan(&tempCR.Idevice, &tempCR.Status.Num, &tempCR.stateStr)
							tempCR.Status.Description = CacheInfo.MapTLSost[tempCR.Status.Num]
							tempCR.State, _ = ConvertStateStrToStruct(tempCR.stateStr)
							arrayCross[tempCR.Idevice] = tempCR
						}
						for idevice, newData := range arrayCross {
							if oldData, ok := globArrCross[idevice]; ok {
								//если запись есть нужно сравнить и если есть разница отправить изменения
								if oldData.State.PK != newData.State.PK || oldData.State.NK != newData.State.NK || oldData.State.CK != newData.State.CK || oldData.Status.Num != newData.Status.Num {
									for conn, info := range crossConnect {
										if info.Idevice == newData.Idevice {
											msg := newCrossMess(typeCrossUpdate, conn, nil, info)
											msg.Data["idevice"] = newData.Idevice
											msg.Data["status"] = newData.Status
											msg.Data["state"] = newData.State
											_ = conn.WriteJSON(msg)
										}
									}
								}
							} else {
								//если не существует старой записи ее нужно отправить
								for conn, info := range crossConnect {
									if info.Idevice == newData.Idevice {
										msg := newCrossMess(typeCrossUpdate, conn, nil, info)
										msg.Data["idevice"] = newData.Idevice
										msg.Data["status"] = newData.Status
										msg.Data["state"] = newData.State
										_ = conn.WriteJSON(msg)
									}
								}
							}
						}
						globArrCross = arrayCross

						//запрос phase
						query, args, err = sqlx.In("SELECT id, fdk, tdk, pdk FROM public.devices WHERE id IN (?)", aPos)
						if err != nil {
							logger.Error.Println("|Message: cross socket cant make IN ", err.Error())
							continue
						}
						query = GetDB().Rebind(query)
						rows, err = GetDB().Queryx(query, args...)
						if err != nil {
							logger.Error.Println("|Message: db not respond", err.Error())
							continue
						}
						for rows.Next() {
							var tempPhase phaseInfo
							_ = rows.Scan(&tempPhase.idevice, &tempPhase.Fdk, &tempPhase.Tdk, &tempPhase.Pdk)
							arrayPhase[tempPhase.idevice] = tempPhase
						}
						for idevice, newData := range arrayPhase {
							if oldData, ok := globArrPhase[idevice]; ok {
								//если запись есть нужно сравнить и если есть разница отправить изменения
								if oldData.Pdk != newData.Pdk || oldData.Tdk != newData.Tdk || oldData.Fdk != newData.Fdk {
									for conn, info := range crossConnect {
										if info.Idevice == newData.idevice {
											msg := newCrossMess(typePhase, conn, nil, info)
											msg.Data["idevice"] = newData.idevice
											msg.Data["fdk"] = newData.Fdk
											msg.Data["tdk"] = newData.Tdk
											msg.Data["pdk"] = newData.Pdk
											_ = conn.WriteJSON(msg)
										}
									}
								}
							} else {
								//если не существует старой записи ее нужно отправить
								for conn, info := range crossConnect {
									if info.Idevice == newData.idevice {
										msg := newCrossMess(typePhase, conn, nil, info)
										msg.Data["idevice"] = newData.idevice
										msg.Data["fdk"] = newData.Fdk
										msg.Data["tdk"] = newData.Tdk
										msg.Data["pdk"] = newData.Pdk
										_ = conn.WriteJSON(msg)
									}
								}
							}
						}
						globArrPhase = arrayPhase
					}
				}
			}
		case pos := <-changeState:
			{
				resp := newControlMess(typeStateChange, nil, nil, crossInfo{})
				state, _ := getNewState(pos)
				resp.Data["state"] = state
				for conn, info := range crossConnect {
					if info.Pos == pos {
						if err := conn.WriteJSON(resp); err != nil {
							delete(crossConnect, conn)
							_ = conn.Close()
						}
					}
				}
			}
		case msg := <-writeCrossMessage:
			{
				switch msg.Type {
				case typeDButton:
					{
						for conn, info := range crossConnect {
							if info.Pos == msg.info.Pos {
								if err := conn.WriteJSON(msg); err != nil {
									delete(crossConnect, conn)
									_ = conn.Close()
								}

							}
						}
					}
				case typeChangeEdit:
					{
						delC := crossConnect[msg.conn]
						delete(crossConnect, msg.conn)
						for cc, coI := range crossConnect {
							if coI.Pos == delC.Pos {
								coI.Edit = true
								crossConnect[cc] = coI
								msg.Data["edit"] = true
								if err := cc.WriteJSON(msg); err != nil {
									delete(crossConnect, cc)
									_ = cc.Close()
								}
								break
							}

						}
					}
				case typeClose:
					{
						delete(crossConnect, msg.conn)
						_ = msg.conn.Close()
					}
				default:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(crossConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				}
			}
		}
	}
}
