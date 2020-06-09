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

var cConnMapUsers map[string][]CrossConn
var editMapCross map[CrossEditInfo]string
var WriteCrossMessage chan CrossSokResponse

//delConn удаление подключения из массива подключений пользователя
func delConn(login string, conn *websocket.Conn) {
	for index, userConn := range cConnMapUsers[login] {
		if userConn.Conn == conn {
			cConnMapUsers[login][index] = cConnMapUsers[login][len(cConnMapUsers[login])-1]
			cConnMapUsers[login][len(cConnMapUsers[login])-1] = CrossConn{}
			cConnMapUsers[login] = cConnMapUsers[login][:len(cConnMapUsers[login])-1]
			break
		}
	}
}

var newCrossMapConnect map[*websocket.Conn]newInfo

type newInfo struct {
	login string
	edit  bool
	pos   CrossEditInfo
}

func CrossReader(crossConn CrossConn, mapContx map[string]string) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	for _, cc := range cConnMapUsers[crossConn.Login] {
		if cc.Pos == crossConn.Pos {
			resp := crossSokMessage(typeError, crossConn, nil)
			resp.Data["message"] = ErrorMessage{Error: errDoubleOpeningDevice}
			_ = crossConn.Conn.WriteJSON(resp)
			_ = crossConn.Conn.Close()
			return
		}
	}

	//===================================
	var (
		conn     *websocket.Conn
		pos      CrossEditInfo
		login    string
		newCInfo newInfo
		//var newCrossMapConnect map[*websocket.Conn]newInfo
	)

	//===================================
	//===================================
	newCInfo.login = login
	newCInfo.pos = pos
	newCInfo.edit = false

	for _, info := range newCrossMapConnect {
		if info.pos == pos && info.login == login {
			resp := crossSokMessage(typeError, crossConn, nil)
			resp.Data["message"] = ErrorMessage{Error: errDoubleOpeningDevice}
			_ = conn.WriteJSON(resp)
			_ = conn.Close()
			return
		}
	}
	flagEdit := false
	for _, info := range newCrossMapConnect {
		if newCInfo.pos == info.pos && info.edit {
			flagEdit = true
		}
	}
	if !flagEdit {
		newCInfo.edit = true
	}
	newCrossMapConnect[conn] = newCInfo
	//===================================
	//===================================
	//проверка редактирования
	delC := newCrossMapConnect[conn]
	delete(newCrossMapConnect, conn)
	for cc, coI := range newCrossMapConnect {
		if coI.pos == delC.pos {
			coI.edit = true
			newCrossMapConnect[cc] = coI

		}
	}
	//===================================
	//===================================
	//===================================

	//проверка редактируется ли данный перекресток
	crossConn.Edit = false
	if _, ok := editMapCross[crossConn.Pos]; !ok {
		editMapCross[crossConn.Pos] = crossConn.Login
		crossConn.Edit = true
		//resp := crossSokMessage(typeChangeEdit, cc, nil)
		//resp.Data["edit"] = true
		//resp.send()
	}

	//если все ОК идем дальше
	cConnMapUsers[crossConn.Login] = append(cConnMapUsers[crossConn.Login], crossConn)

	//сборка начальной информации для отображения перекрестка
	{
		resp := crossInfo(crossConn.Pos)
		resp.ccInfo.Conn = crossConn.Conn
		resp.Data["edit"] = crossConn.Edit
		resp.Data["controlCrossFlag"] = false
		controlCrossFlag, _ := AccessCheck(crossConn.Login, mapContx["role"], 5)
		if (crossConn.Pos.Region == mapContx["region"]) || (mapContx["region"] == "*") {
			resp.Data["controlCrossFlag"] = controlCrossFlag
		}
		resp.send()
	}

	fmt.Println(cConnMapUsers)
	fmt.Println(editMapCross)

	for {
		_, p, err := crossConn.Conn.ReadMessage()
		if err != nil {
			delConn(crossConn.Login, crossConn.Conn)

			//проверка был ли пользователь редактором перекрестка
			//TODO мне не нравится этот кусок
			if editMapCross[crossConn.Pos] == crossConn.Login {
				delete(editMapCross, crossConn.Pos)
				for user, ccs := range cConnMapUsers {
					for num, cc := range ccs {
						if cc.Pos == crossConn.Pos {
							editMapCross[crossConn.Pos] = user
							cConnMapUsers[user][num].Edit = true
							resp := crossSokMessage(typeChangeEdit, cc, nil)
							resp.Data["edit"] = true
							resp.send()
						}
					}
				}
			}
			//--------------------------------------

			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := crossSokMessage(typeError, crossConn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		// case
		case typeDButton:
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = crossConn.Login
				resp := DispatchControl(arm)
				resp.ccInfo = crossConn
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
		resp := crossSokMessage(typeError, CrossConn{}, nil)
		resp.Data["message"] = "failed to Marshal ArmControlData information"
		return resp
	}
	armMessage.User = arm.User
	tcpConnect.ArmCommandChan <- armMessage
	for {
		chanRespond := <-tcpConnect.ArmCommandChan
		if strings.Contains(armMessage.User, arm.User) {
			if chanRespond.Message == "ok" {
				resp := crossSokMessage(typeDButton, CrossConn{}, nil)
				resp.Data["message"] = fmt.Sprintf("command %v send to server", armMessage.CommandStr)
				resp.Data["user"] = arm.User
				return resp
			} else {
				resp := crossSokMessage(typeDButton, CrossConn{}, nil)
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
	cConnMapUsers = make(map[string][]CrossConn)
	WriteCrossMessage = make(chan CrossSokResponse)
	editMapCross = make(map[CrossEditInfo]string)
	newCrossMapConnect = make(map[*websocket.Conn]newInfo)

	readTick := time.Tick(time.Second * 1)
	for {
		select {
		case <-readTick:
			{

			}
		case msg := <-WriteCrossMessage:
			{
				if msg.Type == typeDButton {
					for _, ccs := range cConnMapUsers {
						for _, cc := range ccs {
							if cc.Pos == msg.ccInfo.Pos {
								if err := cc.Conn.WriteJSON(msg); err != nil {
									delConn(cc.Login, cc.Conn)
									_ = cc.Conn.Close()
								}
							}
						}
					}
				} else {
					if err := msg.ccInfo.Conn.WriteJSON(msg); err != nil {
						delConn(msg.ccInfo.Login, msg.ccInfo.Conn)
						_ = msg.ccInfo.Conn.Close()
					}
				}
			}
		}
	}
}
