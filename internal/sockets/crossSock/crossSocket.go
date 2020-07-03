package crossSock

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/comm"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

var writeCrossMessage chan CrossSokResponse
var crossConnect map[*websocket.Conn]CrossInfo
var changeState chan PosInfo
var crossUsersForDisplay chan []CrossInfo
var CrossUsersForMap chan []CrossInfo
var discCrossUsers chan []CrossInfo
var getCrossUsersForDisplay chan bool
var armDeleted chan CrossInfo
var GetCrossUserForMap chan bool

const pingPeriod = time.Second * 30

//CrossReader обработчик открытия сокета для перекрестка
func CrossReader(conn *websocket.Conn, pos PosInfo, mapContx map[string]string, db *sqlx.DB) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var crossCI = CrossInfo{Login: mapContx["login"], Role: mapContx["role"], Pos: pos, Edit: false}

	//проверка не существование такого перекрестка (сбос если нету)
	_, err := getNewState(pos, db)
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
	//если роль пришедшего Viewer то влаг ему не ставим
	flagEdit := false
	if crossCI.Role != "Viewer" {
		for _, info := range crossConnect {
			if crossCI.Pos == info.Pos && info.Edit {
				flagEdit = true
				break
			}
		}
		if !flagEdit {
			crossCI.Edit = true
		}
	}

	//сборка начальной информации для отображения перекрестка
	{
		resp := newCrossMess(typeCrossBuild, conn, nil, crossCI)
		resp, crossCI.Idevice, crossCI.Description = takeCrossInfo(crossCI.Pos, db)
		resp.conn = conn
		resp.Data["edit"] = crossCI.Edit
		resp.Data["controlCrossFlag"] = false
		controlCrossFlag, _ := data.AccessCheck(crossCI.Login, mapContx["role"], 4)
		if (fmt.Sprint(resp.Data["region"]) == mapContx["region"]) || (mapContx["region"] == "*") {
			resp.Data["controlCrossFlag"] = controlCrossFlag
		}
		delete(resp.Data, "region")
		resp.send()
	}

	//добавление в пул перекрестка
	crossConnect[conn] = crossCI
	if crossCI.Edit {
		resp := newCrossMess(typeEditCrossUsers, conn, nil, crossCI)
		resp.send()
	}
	fmt.Println("cross: ", crossConnect)
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

		typeSelect, err := sockets.ChoseTypeMessage(p)
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
func CrossBroadcast(db *sqlx.DB) {
	writeCrossMessage = make(chan CrossSokResponse)
	crossConnect = make(map[*websocket.Conn]CrossInfo)
	changeState = make(chan PosInfo)
	crossUsersForDisplay = make(chan []CrossInfo)
	CrossUsersForMap = make(chan []CrossInfo)
	discCrossUsers = make(chan []CrossInfo)
	getCrossUsersForDisplay = make(chan bool)
	armDeleted = make(chan CrossInfo)
	GetCrossUserForMap = make(chan bool)

	type crossUpdateInfo struct {
		Idevice  int             `json:"idevice"`
		Status   data.TLSostInfo `json:"status"`
		State    agspudge.Cross  `json:"state"`
		stateStr string
	}

	globArrCross := make(map[int]crossUpdateInfo)
	globArrPhase := make(map[int]phaseInfo)
	readTick := time.NewTicker(time.Second * 1)
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		readTick.Stop()
		pingTicker.Stop()
	}()
	for {
		select {
		case <-readTick.C: //ok
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
						query = db.Rebind(query)
						rows, err := db.Queryx(query, args...)
						if err != nil {
							logger.Error.Println("|Message: db not respond", err.Error())
							continue
						}
						for rows.Next() {
							var tempCR crossUpdateInfo
							_ = rows.Scan(&tempCR.Idevice, &tempCR.Status.Num, &tempCR.stateStr)
							tempCR.Status.Description = data.CacheInfo.MapTLSost[tempCR.Status.Num]
							tempCR.State, _ = convertStateStrToStruct(tempCR.stateStr)
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
						query = db.Rebind(query)
						rows, err = db.Queryx(query, args...)
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
		case <-GetCrossUserForMap:
			{
				//отправить на мапу подключенные устройства которые редактируют
				CrossUsersForMap <- formCrossUser()
			}
		case pos := <-changeState: //ok
			{
				resp := newCrossMess(typeStateChange, nil, nil, CrossInfo{})
				state, _ := getNewState(pos, db)
				resp.Data["state"] = state
				for conn, info := range crossConnect {
					if info.Pos == pos {
						_ = conn.WriteJSON(resp)
					}
				}
			}
		case msg := <-writeCrossMessage: //ok
			{
				switch msg.Type {
				case typeDButton:
					{
						for conn, info := range crossConnect {
							if info.Pos == msg.info.Pos {
								_ = conn.WriteJSON(msg)
							}
						}
					}
				case typeChangeEdit:
					{
						delC := crossConnect[msg.conn]
						delete(crossConnect, msg.conn)
						for cc, coI := range crossConnect {
							if (coI.Pos == delC.Pos) && (coI.Role != "Viewer") {
								coI.Edit = true
								crossConnect[cc] = coI
								msg.Data["edit"] = true
								_ = cc.WriteJSON(msg)
								break
							}
						}
						//отправить на мапу подключенные устройства которые редактируют
						CrossUsersForMap <- formCrossUser()
					}
				case typeEditCrossUsers:
					{
						CrossUsersForMap <- formCrossUser()
					}
				case typeClose:
					{
						delete(crossConnect, msg.conn)
					}
				default:
					{
						_ = msg.conn.WriteJSON(msg)
					}
				}
			}
		case <-getCrossUsersForDisplay: // собрать всех кто есть на перекрестке
			{
				var temp = make([]CrossInfo, 0)
				for _, info := range crossConnect {
					temp = append(temp, info)
				}
				crossUsersForDisplay <- temp
			}
		case dCrInfo := <-discCrossUsers: //ok
			{
				for _, dCr := range dCrInfo {
					for conn, cross := range crossConnect {
						if cross.Pos == dCr.Pos && cross.Login == dCr.Login {
							_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "закрытие администратором"))
						}
					}
				}
			}
		case <-pingTicker.C: //ok
			{
				for conn := range crossConnect {
					_ = conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		case armInfo := <-armDeleted: //ok
			{
				for conn, info := range crossConnect {
					if info.Pos == armInfo.Pos {
						_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "перекресток удален"))
					}
				}
			}
		}

	}
}
