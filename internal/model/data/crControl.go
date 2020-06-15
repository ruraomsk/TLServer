package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/deviceLog"
	"net/http"
	"strconv"
	"strings"

	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/stateVerified"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
)

func TestCrossStateData(mapContx map[string]string) u.Response {
	var (
		stateSql  string
		stateInfo []deviceLog.BusyArm
		state     deviceLog.BusyArm
	)
	sqlStr := fmt.Sprintf(`SELECT state FROM public.cross `)
	if mapContx["region"] != "*" {
		sqlStr += fmt.Sprintf(`WHERE region = %v `, mapContx["region"])
	}
	sqlStr += "order by describ"
	rows, err := GetDB().Query(sqlStr)
	if err != nil {
		resp := u.Message(http.StatusInternalServerError, "DB not respond")
		return resp
	}
	for rows.Next() {
		_ = rows.Scan(&stateSql)
		testState, err := ConvertStateStrToStruct(stateSql)
		if err != nil {
			logger.Error.Println("|Message: Failed to parse cross information: ", err.Error())
			return u.Message(http.StatusInternalServerError, "failed to parse cross information")
		}
		var verif stateVerified.StateResult
		verifiedState(&testState, &verif)
		if verif.Err != nil {
			state.ID = testState.ID
			state.Region = strconv.Itoa(testState.Region)
			state.Area = strconv.Itoa(testState.Area)
			state.Description = testState.Name
			stateInfo = append(stateInfo, state)
		}
	}
	resp := u.Message(http.StatusOK, "state data")
	resp.Obj["arms"] = stateInfo
	return resp
}

//takeControlInfo формарование необходимой информации о арме перекрестка
func takeControlInfo(pos PosInfo) (resp ControlSokResponse, idev int, desc string) {
	var (
		stateStr string
	)
	rowsTL := GetDB().QueryRow(`SELECT state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "No result at these points, table cross"
		return resp, 0, ""
	}
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to parse cross information"
		return resp, 0, ""
	}
	resp = newControlMess(typeControlBuild, nil, nil, crossInfo{})
	resp.Data["state"] = rState
	return resp, rState.IDevice, rState.Name
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
	stateMessage.Info = fmt.Sprintf("Idevice: %v, position : %v//%v//%v", state.IDevice, state.Region, state.Area, state.ID)
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
		resp.Data["ok"] = true
		return resp
	} else {
		resp.Data["message"] = "TCP Server not responding"
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
	sqlStr := fmt.Sprintf(`SELECT state FROM public.cross WHERE state::jsonb @> '{"Idevice":%v}'::jsonb OR (region = %v and area = %v and id = %v)`, state.IDevice, state.Region, state.Area, state.ID)
	rows, err := GetDB().Query(sqlStr)
	if err != nil {
		logger.Error.Println("|Message: control socket (create Button), DB not respond : ", err.Error())
		resp := newControlMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "DB not respond"
		return resp
	}

	for rows.Next() {
		_ = rows.Scan(&stateSql)
		if strings.Contains(stateSql, fmt.Sprintf(`"Idevice": %v`, state.IDevice)) {
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
			resp.Data["ok"] = true
			return resp
		} else {
			resp.Data["message"] = "cross created without Map.png - contact admin"
			resp.Data["ok"] = true
			return resp
		}
	} else {
		resp.Data["message"] = "TCP Server not responding"
		return resp
	}
}

//sendToUDPServer отправление данных в канал
func sendToUDPServer(message tcpConnect.StateMessage) bool {
	tcpConnect.StateChan <- message
	for {
		chanRespond := <-tcpConnect.StateChan
		if chanRespond.User == message.User {
			if chanRespond.Message == "ok" {
				return true
			} else {
				return false
			}
		}
	}
}

//stateMarshal преобразовать структуру в строку
func stateMarshal(cross agS_pudge.UserCross) (str string, err error) {
	newByte, err := json.Marshal(cross)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}

//verifiedState набор проверкок для стейта
func verifiedState(cross *agS_pudge.Cross, result *stateVerified.StateResult) {
	resultDay := stateVerified.DaySetsVerified(cross)
	appendResult(result, resultDay)
	resultWeek, empty := stateVerified.WeekSetsVerified(cross)
	appendResult(result, resultWeek)
	resultMouth := stateVerified.MouthSetsVerified(cross, empty)
	appendResult(result, resultMouth)
	resultTimeUse := stateVerified.TimeUseVerified(cross)
	appendResult(result, resultTimeUse)
	resultCtrl := stateVerified.CtrlVerified(cross)
	appendResult(result, resultCtrl)
	return
}

//appendResult накапливание результатов верификации
func appendResult(mainRes *stateVerified.StateResult, addResult stateVerified.StateResult) {
	mainRes.SumResult = append(mainRes.SumResult, addResult.SumResult...)
	if addResult.Err != nil {
		mainRes.Err = addResult.Err
	}
}
