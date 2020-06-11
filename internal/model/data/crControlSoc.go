package data

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/ag-server/comm"
	"time"
)

var writeControlMessage chan ControlSokResponse
var controlMapConnect map[*websocket.Conn]crossInfo

//ControlReader обработчик открытия сокета для арма перекрестка
func ControlReader(conn *websocket.Conn, pos PosInfo, mapContx map[string]string) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var controlI = crossInfo{login: mapContx["login"], pos: pos, edit: false}

	//проверка на полномочия редактирования
	if !((pos.Region == mapContx["region"]) || (mapContx["region"] == "*")) {
		resp := newControlMess(typeError, conn, nil, controlI)
		resp.Data["message"] = ErrorMessage{Error: typeNotEdit}
		_ = conn.WriteMessage(websocket.CloseNormalClosure, []byte(typeNotEdit))
		//resp.send()
		conn.Close()
		return
	}

	//есть ли уже открытый арм у этого пользователя
	for _, info := range controlMapConnect {
		if info.pos == pos && info.login == controlI.login {
			resp := newControlMess(typeError, conn, nil, controlI)
			resp.Data["message"] = ErrorMessage{Error: errDoubleOpeningDevice}
			_ = conn.WriteMessage(websocket.CloseNormalClosure, []byte(errDoubleOpeningDevice))
			conn.Close()
			//resp.send()
			return
		}
	}

	//флаг редактирования перекрестка
	flagEdit := false
	for _, info := range controlMapConnect {
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
	controlMapConnect[conn] = controlI

	fmt.Println("control cok :", controlMapConnect)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//проверка редактирования
			if controlMapConnect[conn].edit {
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
		case typeSendB:
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := sendCrossData(temp.State, controlI.login)
				resp.conn = conn
				resp.info = controlI
				resp.send()
			}
		case typeCheckB:
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := checkCrossData(temp.State)
				resp.conn = conn
				resp.send()
			}
		case typeCreateB:
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := createCrossData(temp.State, controlI.login)
				resp.conn = conn
				resp.send()

			}
		case typeDeleteB:
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := deleteCrossData(temp.State, controlI.login)
				resp.conn = conn
				resp.send()
			}
		case typeUpdateB:
			{
				resp := newControlMess(typeUpdateB, conn, nil, controlI)
				resp, _ = takeControlInfo(controlI.pos)
				resp.conn = conn
				resp.send()
			}
		case typeEditInfoB:
			{
				resp := newControlMess(typeEditInfoB, conn, nil, controlI)

				type UsersEdit struct {
					User string `json:"user"`
					Edit bool   `json:"edit"`
				}
				var users []UsersEdit

				for _, info := range controlMapConnect {
					if info.pos == controlI.pos {
						temp := UsersEdit{User: info.login, Edit: info.edit}
						users = append(users, temp)
					}
				}
				resp.Data["users"] = users
				resp.send()
			}
		case typeDButton:
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
	controlMapConnect = make(map[*websocket.Conn]crossInfo)

	readTick := time.Tick(time.Second * 1)
	for {
		select {
		case <-readTick:
			{

			}
		case msg := <-writeControlMessage:
			{
				switch msg.Type {
				case typeSendB:
					{
						if _, ok := msg.Data["state"]; ok {
							//если есть поле отправить всем кто слушает
							for conn, info := range controlMapConnect {
								if info.pos == msg.info.pos {
									if err := conn.WriteJSON(msg); err != nil {
										delete(controlMapConnect, conn)
										_ = msg.conn.Close()
									}
								}
							}
						} else {
							// если нету поля отправить ошибку только пользователю
							if err := msg.conn.WriteJSON(msg); err != nil {
								delete(controlMapConnect, msg.conn)
								_ = msg.conn.Close()
							}
						}
					}
				case typeCheckB:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(controlMapConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				case typeCreateB:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(controlMapConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				case typeDeleteB:
					{
						if _, ok := msg.Data["ok"]; ok {
							//если есть поле отправить всем кто слушает
							for conn, info := range controlMapConnect {
								if info.pos == msg.info.pos {
									delete(controlMapConnect, conn)
									_ = msg.conn.Close()
								}
							}
						} else {
							// если нету поля отправить ошибку только пользователю
							if err := msg.conn.WriteJSON(msg); err != nil {
								delete(controlMapConnect, msg.conn)
								_ = msg.conn.Close()
							}
						}
					}
				case typeUpdateB:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(controlMapConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				case typeEditInfoB:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(controlMapConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				case typeDButton:
					{
						for conn, info := range controlMapConnect {
							if info.pos == msg.info.pos {
								if err := conn.WriteJSON(msg); err != nil {
									delete(controlMapConnect, conn)
									_ = msg.conn.Close()
								}
							}
						}
					}
				case typeChangeEdit:
					{
						delC := controlMapConnect[msg.conn]
						delete(controlMapConnect, msg.conn)
						for cc, coI := range controlMapConnect {
							if coI.pos == delC.pos {
								coI.edit = true
								controlMapConnect[cc] = coI
								msg.Data["edit"] = true
								if err := cc.WriteJSON(msg); err != nil {
									delete(controlMapConnect, cc)
									_ = cc.Close()
								}
								break
							}
						}
					}
				case typeClose:
					{
						delete(controlMapConnect, msg.conn)
						_ = msg.conn.Close()
					}
				default:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(controlMapConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				}
			}
		}
	}
}
