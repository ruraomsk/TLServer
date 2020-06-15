package data

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/ag-server/comm"
)

var writeControlMessage chan ControlSokResponse
var controlConnect map[*websocket.Conn]crossInfo

//ControlReader обработчик открытия сокета для арма перекрестка
func ControlReader(conn *websocket.Conn, pos PosInfo, mapContx map[string]string) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var controlI = crossInfo{login: mapContx["login"], pos: pos, edit: false}

	//проверка не существование такого перекрестка (сбос если нету)
	_, err := getNewState(pos)
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errThereIsnSuchIntersection))
		return
	}

	//проверка на полномочия редактирования
	if !((pos.Region == mapContx["region"]) || (mapContx["region"] == "*")) {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, typeNotEdit))
		return
	}

	//есть ли уже открытый арм у этого пользователя
	for _, info := range controlConnect {
		if info.pos == pos && info.login == controlI.login {
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errDoubleOpeningDevice))
			return
		}
	}

	//флаг редактирования перекрестка
	flagEdit := false
	for _, info := range controlConnect {
		if controlI.pos == info.pos && info.edit {
			flagEdit = true
			break
		}
	}
	if !flagEdit {
		controlI.edit = true
	}

	//сборка начальной информации для отображения перекрестка
	{
		resp := newControlMess(typeControlBuild, conn, nil, controlI)
		resp, controlI.idevice = takeControlInfo(controlI.pos)
		resp.conn = conn
		CacheInfo.Mux.Lock()
		resp.Data["areaMap"] = CacheInfo.MapArea[CacheInfo.MapRegion[pos.Region]]
		CacheInfo.Mux.Unlock()
		resp.Data["edit"] = controlI.edit
		resp.send()
	}

	//добавление в пул перекрестка
	controlConnect[conn] = controlI

	fmt.Println("control cok :", controlConnect)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//проверка редактирования
			if controlConnect[conn].edit {
				resp := newControlMess(typeChangeEdit, conn, nil, controlI)
				resp.send()
			} else {
				resp := newControlMess(typeClose, conn, nil, controlI)
				resp.send()
			}
			return
		}

		typeSelect, err := setTypeMessage(p)
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
				resp := sendCrossData(temp.State, controlI.login)
				resp.conn = conn
				resp.info = controlI
				resp.send()
			}
		case typeCheckB: //проверка state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := checkCrossData(temp.State)
				resp.conn = conn
				resp.send()
			}
		case typeCreateB: //создание перекрестка
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := createCrossData(temp.State, controlI.login)
				resp.conn = conn
				resp.send()

			}
		case typeDeleteB: //удаление state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := deleteCrossData(temp.State, controlI.login)
				resp.conn = conn
				resp.send()
			}
		case typeUpdateB: //обновление state
			{
				resp := newControlMess(typeUpdateB, conn, nil, controlI)
				resp, _ = takeControlInfo(controlI.pos)
				resp.conn = conn
				resp.send()
			}
		case typeEditInfoB: //информация о пользователях что редактируют перекресток
			{
				resp := newControlMess(typeEditInfoB, conn, nil, controlI)

				type UsersEdit struct {
					User string `json:"user"`
					Edit bool   `json:"edit"`
				}
				var users []UsersEdit

				for _, info := range controlConnect {
					if info.pos == controlI.pos {
						temp := UsersEdit{User: info.login, Edit: info.edit}
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
				arm.User = controlI.login
				resp := DispatchControl(arm)
				resp.info = controlI
				resp.conn = conn
				resp.send()
			}
		}
	}
}

//ControlBroadcast передатчик для арма перекрестка (crossControl)
func ControlBroadcast() {
	writeControlMessage = make(chan ControlSokResponse)
	controlConnect = make(map[*websocket.Conn]crossInfo)
	for {
		select {
		case msg := <-writeControlMessage:
			{
				switch msg.Type {
				case typeSendB:
					{
						if _, ok := msg.Data["state"]; ok {
							//если есть поле отправить всем кто слушает
							for conn, info := range controlConnect {
								if info.pos == msg.info.pos {
									if err := conn.WriteJSON(msg); err != nil {
										delete(controlConnect, conn)
										_ = conn.Close()
									}
								}
							}
							changeState <- msg.info.pos
						} else {
							// если нету поля отправить ошибку только пользователю
							defaultSend(msg)
						}
					}
				case typeCheckB:
					{
						defaultSend(msg)
					}
				case typeCreateB:
					{
						defaultSend(msg)
						mapRepaint <- true
					}
				case typeDeleteB:
					{
						if _, ok := msg.Data["ok"]; ok {
							//если есть поле отправить всем кто слушает
							for conn, info := range controlConnect {
								if info.pos == msg.info.pos {
									delete(controlConnect, conn)
									_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "cross deleted"))
									_ = conn.Close()
								}
							}
							for conn, info := range crossConnect {
								if info.pos == msg.info.pos {
									delete(crossConnect, conn)
									_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "cross deleted"))
									_ = conn.Close()
								}
							}
						} else {
							// если нету поля отправить ошибку только пользователю
							defaultSend(msg)
						}
						mapRepaint <- true
					}
				case typeUpdateB:
					{
						defaultSend(msg)
					}
				case typeEditInfoB:
					{
						defaultSend(msg)
					}
				case typeDButton:
					{
						for conn, info := range controlConnect {
							if info.pos == msg.info.pos {
								if err := conn.WriteJSON(msg); err != nil {
									delete(controlConnect, conn)
									_ = conn.Close()
								}
							}
						}
					}
				case typeChangeEdit:
					{
						delC := controlConnect[msg.conn]
						delete(controlConnect, msg.conn)
						for cc, coI := range controlConnect {
							if coI.pos == delC.pos {
								coI.edit = true
								controlConnect[cc] = coI
								msg.Data["edit"] = true
								if err := cc.WriteJSON(msg); err != nil {
									delete(controlConnect, cc)
									_ = cc.Close()
								}
								break
							}
						}
					}
				case typeClose:
					{
						delete(controlConnect, msg.conn)
						_ = msg.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, typeClose))
						_ = msg.conn.Close()
					}
				default:
					{
						defaultSend(msg)
					}
				}
			}
		}
	}
}

//defaultSend стандартная отправка для одного пользователя
func defaultSend(msg ControlSokResponse) {
	if err := msg.conn.WriteJSON(msg); err != nil {
		delete(controlConnect, msg.conn)
		_ = msg.conn.Close()
	}
}
