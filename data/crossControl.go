package data

import (
	"../logger"
	"../stateVerified"
	"../tcpConnect"
	u "../utils"
	"encoding/json"
	"errors"
	"fmt"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"os"
	"strings"
	"time"
)

//ControlGetCrossInfo сбор информации для пользователя в расширенном варианте
func ControlGetCrossInfo(TLignt TrafficLights, mapContx map[string]string) map[string]interface{} {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)
	sqlStr = fmt.Sprintf("select area, subarea, idevice, dgis, describ, state from %v where region = %v and id = %v and area = %v", os.Getenv("gis_table"), TLignt.Region.Num, TLignt.ID, TLignt.Area.Num)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &StateStr)
	if err != nil {
		logger.Error.Println("|Message: No result at these points", err.Error())
		return u.Message(false, "No result at these points")
	}
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(StateStr)
	if err != nil {
		logger.Error.Println("|Message: Failed to parse cross information: ", err.Error())
		return u.Message(false, "Failed to parse cross information")
	}
	TLignt.Points.StrToFloat(dgis)
	resp := u.Message(true, "Cross control information")
	TLignt.Sost.Num = rState.StatusDevice
	CacheInfo.mux.Lock()
	TLignt.Region.NameRegion = CacheInfo.mapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.mapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	TLignt.Sost.Description = CacheInfo.mapTLSost[TLignt.Sost.Num]
	resp["areaMap"] = CacheInfo.mapArea[TLignt.Region.NameRegion]
	CacheInfo.mux.Unlock()
	resp["cross"] = TLignt
	resp["state"] = rState
	return resp
}

//ControlEditableCheck проверка редактируется ли данный перекресток
func ControlEditableCheck(arm BusyArm, mapContx map[string]string) map[string]interface{} {
	var EditInfo EditCrossInfo
	BusyArmInfo.mux.Lock()
	defer BusyArmInfo.mux.Unlock()
	if _, ok := BusyArmInfo.mapBusyArm[arm]; !ok {
		EditInfo.Login = mapContx["login"]
		EditInfo.EditFlag = true
		EditInfo.time = time.Now()
		BusyArmInfo.mapBusyArm[arm] = EditInfo
	} else {
		EditInfo.Login = BusyArmInfo.mapBusyArm[arm].Login
		if BusyArmInfo.mapBusyArm[arm].Login == mapContx["login"] {
			EditInfo.EditFlag = true
			EditInfo.time = time.Now()
			BusyArmInfo.mapBusyArm[arm] = EditInfo
		} else {
			EditInfo.EditFlag = false
		}
		if BusyArmInfo.mapBusyArm[arm].time.Add(time.Second * 7).Before(time.Now()) {
			EditInfo.Login = mapContx["login"]
			EditInfo.EditFlag = true
			EditInfo.time = time.Now()
			BusyArmInfo.mapBusyArm[arm] = EditInfo
		}
	}
	resp := u.Message(true, "Editable flag")
	resp["DontWrite"] = "true"
	resp["editInfo"] = EditInfo
	return resp
}

//SendCrossData получение данных от пользователя проверка и отправка серверу(устройств)
func SendCrossData(state agS_pudge.Cross, mapContx map[string]string) map[string]interface{} {
	var (
		stateMessage tcpConnect.StateMessage
		err          error
		verif        stateVerified.StateResult
	)
	verifiedState(&state, &verif)
	if verif.Err != nil {
		resp := u.Message(false, fmt.Sprintf("Data didn't pass verification. IDevice: %v", state.IDevice))
		resp["result"] = verif.SumResult
		return resp
	}
	stateMessage.StateStr, err = stateMarshal(state)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal state information: ", err.Error())
		return u.Message(false, "Failed to Marshal state information")
	}
	stateMessage.User = mapContx["login"]
	if sendToUDPServer(stateMessage) {
		return u.Message(true, "Cross send to server")
	} else {
		return u.Message(false, "TCP Server not responding")
	}
}

//CheckCrossData проверка полученных данных
func CheckCrossData(state agS_pudge.Cross) map[string]interface{} {
	var verif stateVerified.StateResult
	verifiedState(&state, &verif)
	if verif.Err != nil {
		resp := u.Message(false, fmt.Sprintf("Data didn't pass verification. IDevice: %v", state.IDevice))
		resp["result"] = verif.SumResult
		return resp
	}
	resp := u.Message(true, "Data is correct")
	resp["result"] = verif.SumResult
	return resp
}

//CreateCrossData добавление нового перекрестка
func CreateCrossData(state agS_pudge.Cross, mapContx map[string]string) map[string]interface{} {
	var (
		stateMessage tcpConnect.StateMessage
		verif        stateVerified.StateResult
		err          error
		stateSql     string
	)
	sqlStr := fmt.Sprintf(`SELECT state FROM public."cross" where state::jsonb @> '{"idevice":%v}'::jsonb or (region = %v and area = %v and id = %v)`, state.IDevice, state.Region, state.Area, state.ID)
	rows, err := GetDB().Raw(sqlStr).Rows()
	if err != nil {
		resp := u.Message(false, "Server not respond")
		return resp
	}
	for rows.Next() {
		_ = rows.Scan(&stateSql)
		if strings.Contains(stateSql, fmt.Sprintf(`"idevice": %v`, state.IDevice)) {
			verif.SumResult = append(verif.SumResult, fmt.Sprintf("№ %v модема уже используется в системе", state.IDevice))
			verif.Err = errors.New("detected")
		}
		if strings.Contains(stateSql, fmt.Sprintf(`"id": %v`, state.ID)) {
			verif.SumResult = append(verif.SumResult, fmt.Sprintf("Данный ID = %v уже занят в регионе: %v районе: %v", state.ID, state.Region, state.Area))
			verif.Err = errors.New("detected")
		}
	}
	verifiedState(&state, &verif)
	if verif.Err != nil {
		resp := u.Message(false, fmt.Sprintf("Data didn't pass verification. IDevice: %v", state.IDevice))
		resp["result"] = verif.SumResult
		return resp
	}
	stateMessage.StateStr, err = stateMarshal(state)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal state information: ", err.Error())
		return u.Message(false, "Failed to Marshal state information")
	}
	stateMessage.User = mapContx["login"]
	if sendToUDPServer(stateMessage) {
		if ShortCreateDirPng(state.Region, state.Area, state.ID, state.Dgis) {
			return u.Message(true, "Cross created")
		} else {
			return u.Message(true, "Cross created without Map.png - contact admin")
		}
	} else {
		return u.Message(false, "TCP Server not responding")
	}
}

//DeleteCrossData удаление перекрестка на сервере
func DeleteCrossData(state agS_pudge.Cross, mapContx map[string]string) map[string]interface{} {
	var (
		stateMessage tcpConnect.StateMessage
		err          error
	)
	stateMessage.Info = fmt.Sprintf("idevice: %v, position : %v//%v//%v", state.IDevice, state.Region, state.Area, state.ID)
	state.IDevice = -1
	stateMessage.StateStr, err = stateMarshal(state)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal state information: ", err.Error())
		return u.Message(false, "Failed to Marshal state information")
	}
	stateMessage.User = mapContx["login"]
	if sendToUDPServer(stateMessage) {
		return u.Message(true, fmt.Sprintf("Cross data deleted. Info (%v)", stateMessage.Info))
	} else {
		return u.Message(false, "TCP Server not responding")
	}
}

func sendToUDPServer(message tcpConnect.StateMessage) bool {
	tcpConnect.StateChan <- message
	for {
		chanRespond := <-tcpConnect.StateChan
		if strings.Contains(chanRespond.User, message.User) {
			if chanRespond.Message == "ok" {
				return true
			} else {
				return false
			}
		}
	}
}

//stateMarshal преобразовать структуру в строку
func stateMarshal(cross agS_pudge.Cross) (str string, err error) {
	newByte, err := json.Marshal(cross)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}

//verifiedState набор проверкок для стейта
func verifiedState(cross *agS_pudge.Cross, result *stateVerified.StateResult) () {
	resultDay := stateVerified.DaySetsVerified(&cross.Arrays.DaySets)
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
