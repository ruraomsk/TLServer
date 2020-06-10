package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/ag-server/comm"
	"strings"
	"time"
)

var writeCrossMessage chan CrossSokResponse
var crossMapConnect map[*websocket.Conn]crossInfo

//CrossReader обработчик открытия сокета для перекрестка
func CrossReader(conn *websocket.Conn, pos CrossEditInfo, mapContx map[string]string) {
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

	//добавление в пул перекрестка
	crossMapConnect[conn] = crossCI

	//сборка начальной информации для отображения перекрестка
	{
		resp := takeCrossInfo(crossCI.pos)
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

//CrossEditInfo передатчик для перекрестка (cross)
func CrossBroadcast() {
	writeCrossMessage = make(chan CrossSokResponse)
	crossMapConnect = make(map[*websocket.Conn]crossInfo)

	readTick := time.Tick(time.Second * 1)
	for {
		select {
		case <-readTick:
			{

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
