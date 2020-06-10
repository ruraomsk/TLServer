package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/comm"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"strings"
	"time"
)

var writeCrossMessage chan CrossSokResponse
var crossMapConnect map[*websocket.Conn]crossInfo

//CrossReader обработчик открытия сокета для перекрестка
func CrossReader(conn *websocket.Conn, pos PosInfo, mapContx map[string]string) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var crossCI = crossInfo{login: mapContx["login"], pos: pos, edit: false}

	for _, info := range crossMapConnect {
		if info.pos == pos && info.login == crossCI.login {
			resp := newCrossMess(typeError, conn, nil, crossCI)
			resp.Data["message"] = ErrorMessage{Error: errDoubleOpeningDevice}
			resp.send()
			return
		}
	}

	//флаг редактирования перекрестка
	flagEdit := false
	for _, info := range crossMapConnect {
		if crossCI.pos == info.pos && info.edit {
			flagEdit = true
			break
		}
	}
	if !flagEdit {
		crossCI.edit = true
	}

	//сборка начальной информации для отображения перекрестка
	{
		resp := newCrossMess(typeCrossBuild, conn, nil, crossCI)
		resp, crossCI.idevice = takeCrossInfo(crossCI.pos)
		resp.conn = conn
		resp.Data["edit"] = crossCI.edit
		resp.Data["controlCrossFlag"] = false
		controlCrossFlag, _ := AccessCheck(crossCI.login, mapContx["role"], 5)
		if (fmt.Sprint(resp.Data["region"]) == mapContx["region"]) || (mapContx["region"] == "*") {
			resp.Data["controlCrossFlag"] = controlCrossFlag
		}
		delete(resp.Data, "region")
		resp.send()
	}

	//добавление в пул перекрестка
	crossMapConnect[conn] = crossCI

	fmt.Println(crossMapConnect)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//проверка редактирования
			if crossMapConnect[conn].edit {
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
				arm.User = crossCI.login
				resp := DispatchControl(arm)
				resp.info = crossCI
				resp.conn = conn
				resp.send()
			}
		}
	}
}

//DispatchControl отправка команды на устройство
func DispatchControl(arm comm.CommandARM) CrossSokResponse {
	var (
		err        error
		armMessage tcpConnect.ArmCommandMessage
	)

	armMessage.CommandStr, err = armControlMarshal(arm)
	if err != nil {
		resp := newCrossMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to Marshal ArmControlData information"
		return resp
	}
	armMessage.User = arm.User
	tcpConnect.ArmCommandChan <- armMessage
	for {
		chanRespond := <-tcpConnect.ArmCommandChan
		if strings.Contains(armMessage.User, arm.User) {
			if chanRespond.Message == "ok" {
				resp := newCrossMess(typeDButton, nil, nil, crossInfo{})
				resp.Data["message"] = fmt.Sprintf("command %v send to server", armMessage.CommandStr)
				resp.Data["user"] = arm.User
				return resp
			} else {
				resp := newCrossMess(typeDButton, nil, nil, crossInfo{})
				resp.Data["message"] = "TCP Server not responding"
				resp.Data["user"] = arm.User
				return resp
			}
		}
	}
}

//armControlMarshal преобразовать структуру в строку
func armControlMarshal(arm comm.CommandARM) (str string, err error) {
	newByte, err := json.Marshal(arm)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}

//PosInfo передатчик для перекрестка (cross)
func CrossBroadcast() {
	writeCrossMessage = make(chan CrossSokResponse)
	crossMapConnect = make(map[*websocket.Conn]crossInfo)

	type crossUpdateInfo struct {
		Idevice  int             `json:"idevice"`
		Status   TLSostInfo      `json:"status"`
		State    agS_pudge.Cross `json:"state"`
		stateStr string
	}

	globArrCross := make(map[int]crossUpdateInfo)
	globArrPhase := make(map[int]phaseInfo)
	readTick := time.Tick(time.Second * 1)
	for {
		select {
		case <-readTick:
			{
				var aPos []int
				arrayCross := make(map[int]crossUpdateInfo)
				arrayPhase := make(map[int]phaseInfo)
				for _, crInfo := range crossMapConnect {
					aPos = append(aPos, crInfo.idevice)
					//arrayPhase[crInfo] = phaseInfo{}
				}
				//выполняем если хоть что-то есть
				if len(aPos) > 0 {
					//запрос статуса и state
					query, args, err := sqlx.In("SELECT idevice, status, state FROM public.cross WHERE idevice IN (?)", aPos)
					if err != nil {
						//todo пока не знаю че с этим делать
						fmt.Println(err.Error())
						continue
					}
					query = GetDB().Rebind(query)
					rows, err := GetDB().Queryx(query, args...)
					if err != nil {
						//todo пока не знаю че с этим делать
						fmt.Println(err.Error())
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
								for conn, info := range crossMapConnect {
									if info.idevice == newData.Idevice {
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
							for conn, info := range crossMapConnect {
								if info.idevice == newData.Idevice {
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
						//todo пока не знаю че с этим делать
						fmt.Println(err.Error())
						continue
					}
					query = GetDB().Rebind(query)
					rows, err = GetDB().Queryx(query, args...)
					if err != nil {
						//todo пока не знаю че с этим делать
						fmt.Println(err.Error())
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
								for conn, info := range crossMapConnect {
									if info.idevice == newData.idevice {
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
							for conn, info := range crossMapConnect {
								if info.idevice == newData.idevice {
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
		case msg := <-writeCrossMessage:
			{
				switch msg.Type {
				case typeDButton:
					{
						for conn, info := range crossMapConnect {
							if info.pos == msg.info.pos {
								if err := conn.WriteJSON(msg); err != nil {
									delete(crossMapConnect, conn)
									_ = msg.conn.Close()
								}

							}
						}
					}
				case typeChangeEdit:
					{
						delC := crossMapConnect[msg.conn]
						delete(crossMapConnect, msg.conn)
						for cc, coI := range crossMapConnect {
							if coI.pos == delC.pos {
								coI.edit = true
								crossMapConnect[cc] = coI
								msg.Data["edit"] = true
								if err := cc.WriteJSON(msg); err != nil {
									delete(crossMapConnect, cc)
									_ = cc.Close()
								}
								break
							}

						}
					}
				case typeClose:
					{
						delete(crossMapConnect, msg.conn)
						_ = msg.conn.Close()
					}
				default:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(crossMapConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				}
			}
		}
	}
}
