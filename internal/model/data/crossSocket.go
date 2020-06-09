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

//var cConnMapUsers map[string][]CrossConn
var WriteCrossMessage chan CrossSokResponse
var CrossMapConnect map[*websocket.Conn]newInfo

//delConn удаление подключения из массива подключений пользователя
//func delConn(login string, conn *websocket.Conn) {
//	for index, userConn := range cConnMapUsers[login] {
//		if userConn.Conn == conn {
//			cConnMapUsers[login][index] = cConnMapUsers[login][len(cConnMapUsers[login])-1]
//			cConnMapUsers[login][len(cConnMapUsers[login])-1] = CrossConn{}
//			cConnMapUsers[login] = cConnMapUsers[login][:len(cConnMapUsers[login])-1]
//			break
//		}
//	}
//}

//var CrossMapConnect map[*websocket.Conn]newInfo

func CrossReader(conn *websocket.Conn, pos CrossEditInfo, mapContx map[string]string) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	var crossCI = newInfo{login: mapContx["login"], pos: pos, edit: false}

	for _, info := range CrossMapConnect {
		if info.pos == pos && info.login == crossCI.login {
			resp := crossSokMessage(typeError, conn, nil, crossCI)
			resp.Data["message"] = ErrorMessage{Error: errDoubleOpeningDevice}
			resp.send()
			return
		}
	}

	//флаг редактирования перекрестка
	flagEdit := false
	for _, info := range CrossMapConnect {
		if crossCI.pos == info.pos && info.edit {
			flagEdit = true
			break
		}
	}
	if !flagEdit {
		crossCI.edit = true
	}

	//добавление в пул перекрестка
	CrossMapConnect[conn] = crossCI

	//сборка начальной информации для отображения перекрестка
	{
		resp := crossInfo(crossCI.pos)
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

	fmt.Println(CrossMapConnect)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//проверка редактирования
			if CrossMapConnect[conn].edit {
				resp := crossSokMessage(typeChangeEdit, conn, nil, crossCI)
				resp.send()
			} else {
				resp := crossSokMessage(typeClose, conn, nil, crossCI)
				resp.send()
			}
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := crossSokMessage(typeError, conn, nil, crossCI)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		// case
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
		resp := crossSokMessage(typeError, nil, nil, newInfo{})
		resp.Data["message"] = "failed to Marshal ArmControlData information"
		return resp
	}
	armMessage.User = arm.User
	tcpConnect.ArmCommandChan <- armMessage
	for {
		chanRespond := <-tcpConnect.ArmCommandChan
		if strings.Contains(armMessage.User, arm.User) {
			if chanRespond.Message == "ok" {
				resp := crossSokMessage(typeDButton, nil, nil, newInfo{})
				resp.Data["message"] = fmt.Sprintf("command %v send to server", armMessage.CommandStr)
				resp.Data["user"] = arm.User
				return resp
			} else {
				resp := crossSokMessage(typeDButton, nil, nil, newInfo{})
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

//CrossEditInfo вещатель для кросса
func CrossBroadcast() {
	WriteCrossMessage = make(chan CrossSokResponse)
	CrossMapConnect = make(map[*websocket.Conn]newInfo)

	readTick := time.Tick(time.Second * 1)
	for {
		select {
		case <-readTick:
			{

			}
		case msg := <-WriteCrossMessage:
			{
				switch msg.Type {
				case typeDButton:
					{
						for conn, info := range CrossMapConnect {
							if info.pos == msg.info.pos {
								if err := conn.WriteJSON(msg); err != nil {
									delete(CrossMapConnect, conn)
									_ = msg.conn.Close()
								}

							}
						}
					}
				case typeChangeEdit:
					{
						delC := CrossMapConnect[msg.conn]
						delete(CrossMapConnect, msg.conn)
						for cc, coI := range CrossMapConnect {
							if coI.pos == delC.pos {
								coI.edit = true
								CrossMapConnect[cc] = coI
								msg.Data["edit"] = true
								if err := cc.WriteJSON(msg); err != nil {
									delete(CrossMapConnect, cc)
									_ = cc.Close()
								}
								break
							}

						}
					}
				case typeClose:
					{
						delete(CrossMapConnect, msg.conn)
						_ = msg.conn.Close()
					}
				default:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							delete(CrossMapConnect, msg.conn)
							_ = msg.conn.Close()
						}
					}
				}
			}
		}
	}
}
