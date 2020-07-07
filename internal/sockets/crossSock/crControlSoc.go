package crossSock

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/techArm"
	"github.com/jmoiron/sqlx"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ruraomsk/ag-server/comm"
)

var writeControlMessage chan ControlSokResponse
var controlConnect map[*websocket.Conn]CrossInfo
var crArmUsersForDisplay chan []CrossInfo
var discArmUsers chan []CrossInfo
var getArmUsersForDisplay chan bool
var MapRepaint chan bool
var UserLogoutCrControl chan string

//ControlReader обработчик открытия сокета для арма перекрестка
func ControlReader(conn *websocket.Conn, pos PosInfo, mapContx map[string]string, db *sqlx.DB) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var controlI = CrossInfo{Login: mapContx["login"], Pos: pos, Edit: false}

	//проверка не существование такого перекрестка (сбос если нету)
	_, err := getNewState(pos, db)
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errCrossDoesntExist))
		return
	}

	//проверка на полномочия редактирования
	if !((pos.Region == mapContx["region"]) || (mapContx["region"] == "*")) {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, typeNotEdit))
		return
	}

	//есть ли уже открытый арм у этого пользователя
	for _, info := range controlConnect {
		if info.Pos == pos && info.Login == controlI.Login {
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errDoubleOpeningDevice))
			return
		}
	}

	//флаг редактирования перекрестка
	flagEdit := false
	for _, info := range controlConnect {
		if controlI.Pos == info.Pos && info.Edit {
			flagEdit = true
			break
		}
	}
	if !flagEdit {
		controlI.Edit = true
	}

	//сборка начальной информации для отображения перекрестка
	{
		resp := newControlMess(typeControlBuild, conn, nil, controlI)
		resp, controlI.Idevice, controlI.Description = takeControlInfo(controlI.Pos, db)
		resp.conn = conn
		data.CacheInfo.Mux.Lock()
		resp.Data["areaMap"] = data.CacheInfo.MapArea[data.CacheInfo.MapRegion[pos.Region]]
		data.CacheInfo.Mux.Unlock()
		resp.Data["edit"] = controlI.Edit
		resp.send()
	}

	//добавление в пул перекрестка
	controlConnect[conn] = controlI

	fmt.Println("control cok :", controlConnect)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//проверка редактирования
			if controlConnect[conn].Edit {
				resp := newControlMess(typeChangeEdit, conn, nil, controlI)
				resp.send()
			} else {
				resp := newControlMess(typeClose, conn, nil, controlI)
				resp.send()
			}
			return
		}

		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			resp := newControlMess(typeError, conn, nil, controlI)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}

		switch typeSelect {
		case typeSendB: //отправка state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := sendCrossData(temp.State, controlI.Login)
				resp.conn = conn
				resp.info = controlI
				resp.send()
			}
		case typeCheckB: //проверка state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := checkCrossData(temp.State)
				resp.info = controlI
				resp.conn = conn
				resp.send()
			}
		case typeCreateB: //создание перекрестка
			{
				temp := struct {
					Type  string         `json:"type"`
					State agspudge.Cross `json:"state"`
					Z     int            `json:"z"`
				}{}
				_ = json.Unmarshal(p, &temp)
				resp := createCrossData(temp.State, controlI.Login, temp.Z, db)
				resp.info = controlI
				resp.conn = conn
				resp.send()

			}
		case typeDeleteB: //удаление state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := deleteCrossData(temp.State, controlI.Login)
				resp.info = controlI
				resp.conn = conn
				resp.send()
			}
		case typeUpdateB: //обновление state
			{
				resp := newControlMess(typeUpdateB, conn, nil, controlI)
				resp, _, _ = takeControlInfo(controlI.Pos, db)
				resp.info = controlI
				resp.Data["edit"] = controlI.Edit
				resp.conn = conn
				resp.send()
			}
		case typeEditInfoB: //информация о пользователях что редактируют перекресток
			{
				resp := newControlMess(typeEditInfoB, conn, nil, controlI)

				type usersEdit struct {
					User string `json:"user"`
					Edit bool   `json:"edit"`
				}
				var users []usersEdit

				for _, info := range controlConnect {
					if info.Pos == controlI.Pos {
						temp := usersEdit{User: info.Login, Edit: info.Edit}
						users = append(users, temp)
					}
				}
				resp.Data["users"] = users
				resp.send()
			}
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = controlI.Login
				resp := newControlMess(typeDButton, conn, nil, controlI)
				resp.Data = sockets.DispatchControl(arm)
				resp.send()
			}
		}
	}
}

//ControlBroadcast передатчик для арма перекрестка (crossControl)
func ControlBroadcast() {
	writeControlMessage = make(chan ControlSokResponse)
	controlConnect = make(map[*websocket.Conn]CrossInfo)
	getArmUsersForDisplay = make(chan bool)
	crArmUsersForDisplay = make(chan []CrossInfo)
	discArmUsers = make(chan []CrossInfo)
	MapRepaint = make(chan bool)
	UserLogoutCrControl = make(chan string)

	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
	}()
	for {
		select {
		case msg := <-writeControlMessage: //ok
			{
				switch msg.Type {
				case typeSendB:
					{
						if _, ok := msg.Data["state"]; ok {
							//если есть поле отправить всем кто слушает
							for conn, info := range controlConnect {
								if info.Pos == msg.info.Pos {
									_ = conn.WriteJSON(msg)
								}
							}
							changeState <- msg.info.Pos
							techArm.TArmNewCrossData <- true
						} else {
							// если нету поля отправить ошибку только пользователю
							_ = msg.conn.WriteJSON(msg)
						}
					}
				case typeCreateB:
					{
						_ = msg.conn.WriteJSON(msg)
						if _, ok := msg.Data["ok"]; ok {
							MapRepaint <- true
							techArm.TArmNewCrossData <- true
						}
					}
				case typeDeleteB:
					{
						if _, ok := msg.Data["ok"]; ok {
							//если есть поле отправить всем кто слушает
							for conn, info := range controlConnect {
								if info.Pos == msg.info.Pos {
									_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "перекресток удален"))
								}
							}
							armDeleted <- msg.info
							MapRepaint <- true
							techArm.TArmNewCrossData <- true
						} else {
							// если нету поля отправить ошибку только пользователю
							_ = msg.conn.WriteJSON(msg)
						}

					}

				case typeDButton:
					{
						for conn, info := range controlConnect {
							if info.Pos == msg.info.Pos {
								_ = conn.WriteJSON(msg)
							}
						}
					}
				case typeChangeEdit:
					{
						delC := controlConnect[msg.conn]
						delete(controlConnect, msg.conn)
						for cc, coI := range controlConnect {
							if coI.Pos == delC.Pos {
								coI.Edit = true
								controlConnect[cc] = coI
								msg.Data["edit"] = true
								_ = cc.WriteJSON(msg)
								break
							}
						}
					}
				case typeClose:
					{
						delete(controlConnect, msg.conn)
					}
				default:
					{
						_ = msg.conn.WriteJSON(msg)
					}
				}
			}
		case <-getArmUsersForDisplay: //ok
			{
				var temp []CrossInfo
				for _, info := range controlConnect {
					temp = append(temp, info)
				}
				crArmUsersForDisplay <- temp
			}
		case dArmInfo := <-discArmUsers:
			{
				for _, dArm := range dArmInfo {
					for conn, cross := range controlConnect {
						if cross.Pos == dArm.Pos && cross.Login == dArm.Login {
							_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "закрытие администратором"))
						}
					}
				}
			}
		case <-pingTicker.C: //ok
			{
				for conn := range controlConnect {
					_ = conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		case login := <-UserLogoutCrControl:
			{
				for conn, crossInfo := range controlConnect {
					if crossInfo.Login == login {
						_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "пользователь вышел из системы"))
					}
				}
			}
		}
	}
}
