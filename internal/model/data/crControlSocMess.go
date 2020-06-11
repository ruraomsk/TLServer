package data

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/stateVerified"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"strings"
)

//ControlSokResponse структура для отправки сообщений (cross control)
type ControlSokResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
	info crossInfo              `json:"-"`
}

//newControlMess создание нового сообщения
func newControlMess(mType string, conn *websocket.Conn, data map[string]interface{}, info crossInfo) ControlSokResponse {
	var resp = ControlSokResponse{Type: mType, conn: conn, info: info}
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//send отправка с обработкой ошибки
func (m *ControlSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v",
				m.conn.RemoteAddr(),
				m.info.login,
				fmt.Sprintf("/cross/control?Region=%v&Area=%v&ID=%v", m.info.pos.Region, m.info.pos.Area, m.info.pos.Id),
				m.Data["message"])
		}()
	}
	writeControlMessage <- *m
}

//takeControlInfo формарование необходимой информации о арме перекрестка
func takeControlInfo(pos PosInfo) (ControlSokResponse, int) {
	var (
		stateStr string
	)
	rowsTL := GetDB().QueryRow(`SELECT state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "No result at these points, table cross"
		return resp, 0
	}
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to parse cross information"
		return resp, 0
	}

	resp := newControlMess(typeControlBuild, nil, nil, crossInfo{})
	resp.Data["state"] = rState
	return resp, rState.IDevice
}

var (
	typeSendB     = "sendB"
	typeCheckB    = "checkB"
	typeCreateB   = "createB"
	typeDeleteB   = "deleteB"
	typeUpdateB   = "updateB"
	typeEditInfoB = "editInfoB"

	typeControlBuild = "controlInfo"
	typeNotEdit      = "no authority to edit"
)

type StateHandler struct {
	Type  string          `json:"type"`
	State agS_pudge.Cross `json:"state"`
}

//checkCrossData проверка полученных данных
func checkCrossData(state agS_pudge.Cross) ControlSokResponse {
	var verif stateVerified.StateResult
	verifiedState(&state, &verif)
	resp := newControlMess(typeCheckB, nil, nil, crossInfo{})
	if verif.Err != nil {
		resp.Data["status"] = false
	} else {
		resp.Data["status"] = true
	}
	resp.Data["result"] = verif.SumResult
	return resp
}

//sendCrossData получение данных от пользователя проверка и отправка серверу(устройств)
func sendCrossData(state agS_pudge.Cross, login string) ControlSokResponse {
	var (
		stateMessage tcpConnect.StateMessage
		err          error
		userCross    agS_pudge.UserCross
	)

	userCross.State = state
	userCross.User = login
	stateMessage.StateStr, err = stateMarshal(userCross)
	if err != nil {
		logger.Error.Println("|Message: control socket, failed to Marshal state information: ", err.Error())
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to Marshal state information"
		return resp
	}
	stateMessage.User = login
	resp := newControlMess(typeSendB, nil, nil, crossInfo{})
	resp.Data["user"] = login
	if sendToUDPServer(stateMessage) {
		resp.Data["message"] = "cross send to server"
		resp.Data["state"] = state
		return resp
	} else {
		resp.Data["message"] = "TCP Server not responding"
		return resp
	}

}

//deleteCrossData удаление перекрестка на сервере
func deleteCrossData(state agS_pudge.Cross, login string) ControlSokResponse {
	var (
		stateMessage tcpConnect.StateMessage
		userCross    agS_pudge.UserCross
		err          error
	)
	stateMessage.Info = fmt.Sprintf("idevice: %v, position : %v//%v//%v", state.IDevice, state.Region, state.Area, state.ID)
	state.IDevice = -1
	userCross.State = state
	userCross.User = login
	stateMessage.StateStr, err = stateMarshal(userCross)
	if err != nil {
		logger.Error.Println("|Message: control socket, failed to Marshal state information: ", err.Error())
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to Marshal state information"
		return resp
	}
	stateMessage.User = login
	resp := newControlMess(typeDeleteB, nil, nil, crossInfo{})
	resp.Data["user"] = login
	if sendToUDPServer(stateMessage) {
		resp.Data["message"] = fmt.Sprintf("cross data deleted. Info (%v)", stateMessage.Info)
		return resp
	} else {
		resp.Data["message"] = "TCP Server not responding"
		resp.Data["ok"] = false
		return resp
	}
}

//createCrossData добавление нового перекрестка
func createCrossData(state agS_pudge.Cross, login string) ControlSokResponse {
	var (
		stateMessage tcpConnect.StateMessage
		userCross    agS_pudge.UserCross
		verRes       []string
		stateSql     string
	)
	sqlStr := fmt.Sprintf(`SELECT state FROM public.cross WHERE state::jsonb @> '{"idevice":%v}'::jsonb OR (region = %v and area = %v and id = %v)`, state.IDevice, state.Region, state.Area, state.ID)
	rows, err := GetDB().Query(sqlStr)
	if err != nil {
		logger.Error.Println("|Message: control socket (create Button), DB not respond : ", err.Error())
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "DB not respond"
		return resp
	}

	for rows.Next() {
		_ = rows.Scan(&stateSql)
		if strings.Contains(stateSql, fmt.Sprintf(`"idevice": %v`, state.IDevice)) {
			verRes = append(verRes, fmt.Sprintf("№ %v модема уже используется в системе", state.IDevice))
		}
		if strings.Contains(stateSql, fmt.Sprintf(`"id": %v`, state.ID)) {
			verRes = append(verRes, fmt.Sprintf("Данный ID = %v уже занят в регионе: %v районе: %v", state.ID, state.Region, state.Area))
		}
	}
	if len(verRes) > 0 {
		resp := newControlMess(typeCreateB, nil, nil, crossInfo{})
		resp.Data["result"] = verRes
		return resp
	}

	userCross.State = state
	userCross.User = login
	stateMessage.StateStr, err = stateMarshal(userCross)
	if err != nil {
		logger.Error.Println("|Message: control socket, failed to Marshal state information: ", err.Error())
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to Marshal state information"
		return resp
	}
	stateMessage.User = login
	resp := newControlMess(typeCreateB, nil, nil, crossInfo{})
	resp.Data["user"] = login
	if sendToUDPServer(stateMessage) {
		if ShortCreateDirPng(state.Region, state.Area, state.ID, state.Dgis) {
			resp.Data["message"] = "cross created"
			return resp
		} else {
			resp.Data["message"] = "cross created without Map.png - contact admin"
			return resp
		}
	} else {
		resp.Data["message"] = "TCP Server not responding"
		return resp
	}
}
