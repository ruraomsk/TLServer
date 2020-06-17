package data

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/ruraomsk/ag-server/comm"
)

var writeControlMessage chan ControlSokResponse
var controlConnect map[*websocket.Conn]CrossInfo
var crArmUsers chan []CrossInfo
var discArmUsers chan []CrossInfo
var getArmUsers chan bool

//ControlReader обработчик открытия сокета для арма перекрестка
func ControlReader(conn *websocket.Conn, pos PosInfo, mapContx map[string]string) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var controlI = CrossInfo{Login: mapContx["login"], Pos: pos, Edit: false}

	//проверка не существование такого перекрестка (сбос если нету)
	_, err := getNewState(pos)
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
		resp, controlI.Idevice, controlI.Description = takeControlInfo(controlI.Pos)
		resp.conn = conn
		CacheInfo.Mux.Lock()
		resp.Data["areaMap"] = CacheInfo.MapArea[CacheInfo.MapRegion[pos.Region]]
		CacheInfo.Mux.Unlock()
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
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := createCrossData(temp.State, controlI.Login)
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
				resp, _, _ = takeControlInfo(controlI.Pos)
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
				resp := dispatchControl(arm)
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
	controlConnect = make(map[*websocket.Conn]CrossInfo)

	getArmUsers = make(chan bool)
	crArmUsers = make(chan []CrossInfo)
	discArmUsers = make(chan []CrossInfo)

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
								if info.Pos == msg.info.Pos {
									if err := conn.WriteJSON(msg); err != nil {
										delete(controlConnect, conn)
										_ = conn.Close()
									}
								}
							}
							changeState <- msg.info.Pos
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
						if _, ok := msg.Data["ok"]; ok {
							mapRepaint <- true
						}
					}
				case typeDeleteB:
					{
						if _, ok := msg.Data["ok"]; ok {
							//если есть поле отправить всем кто слушает
							for conn, info := range controlConnect {
								if info.Pos == msg.info.Pos {
									delete(controlConnect, conn)
									_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "cross deleted"))
									_ = conn.Close()
								}
							}
							for conn, info := range crossConnect {
								if info.Pos == msg.info.Pos {
									delete(crossConnect, conn)
									_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "cross deleted"))
									_ = conn.Close()
								}
							}
							mapRepaint <- true
						} else {
							// если нету поля отправить ошибку только пользователю
							defaultSend(msg)
						}

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
							if info.Pos == msg.info.Pos {
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
							if coI.Pos == delC.Pos {
								coI.Edit = true
								controlConnect[cc] = coI
								msg.Data["edit"] = true
								if err := cc.WriteJSON(msg); err != nil {
									delete(controlConnect, cc)
									_ = cc.Close()
									continue
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
		case <-getArmUsers:
			{
				var temp []CrossInfo
				for _, info := range controlConnect {
					temp = append(temp, info)
				}
				crArmUsers <- temp
			}
		case dArmInfo := <-discArmUsers:
			{
				for _, dArm := range dArmInfo {
					for conn, cross := range controlConnect {
						if cross.Pos == dArm.Pos && cross.Login == dArm.Login {
							//проверка редактирования
							if cross.Edit {
								delete(controlConnect, conn)
								_ = conn.Close()
								for cc, coI := range controlConnect {
									if coI.Pos == dArm.Pos {
										coI.Edit = true
										controlConnect[cc] = coI
										resp := newCrossMess(typeChangeEdit, nil, nil, coI)
										resp.Data["edit"] = true
										if err := cc.WriteJSON(resp); err != nil {
											delete(controlConnect, cc)
											_ = cc.Close()
											continue
										}
										break
									}
								}
							} else {
								delete(controlConnect, conn)
								_ = conn.Close()
							}
						}
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
