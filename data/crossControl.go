package data

import (
	"../logger"
	"../stateVerified"
	"../tcpConnect"
	u "../utils"
	"encoding/json"
	"fmt"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"os"
	"strings"
)

//ControlGetCrossInfo сбор информации для пользователя в расширенном варианте
func ControlGetCrossInfo(TLignt TrafficLights) map[string]interface{} {
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
	TLignt.Points.StrToFloat(dgis)
	TLignt.Region.NameRegion = CacheInfo.mapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.mapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(StateStr)
	if err != nil {
		logger.Error.Println("|Message: Failed to parse cross information: ", err.Error())
		return u.Message(false, "Failed to parse cross information")
	}
	TLignt.Sost.Num = rState.StatusDevice
	TLignt.Sost.Description = CacheInfo.mapTLSost[TLignt.Sost.Num]
	resp := u.Message(true, "Cross information")
	resp["cross"] = TLignt
	resp["state"] = rState

	resp["areaMap"] = CacheInfo.mapArea[TLignt.Region.NameRegion]

	return resp
}

//SendCrossData получение данных от пользователя проверка и отправка серверу(устройств)
func SendCrossData(state agS_pudge.Cross, mapContx map[string]string) map[string]interface{} {
	var (
		stateMessage tcpConnect.StateMessage
		err          error
	)
	verif := verifiedState(&state)
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
	tcpConnect.StateChan <- stateMessage
	for {
		chanRespond := <-tcpConnect.StateChan
		if strings.Contains(chanRespond.User, mapContx["login"]) {
			if chanRespond.Message == "ok" {
				return u.Message(true, fmt.Sprintf("Cross data send to server. IDevice: %v", state.IDevice))
			} else {
				return u.Message(false, "TCP Server not responding")
			}
		}
	}
}

//CheckCrossData проверка полученных данных
func CheckCrossData(state agS_pudge.Cross) map[string]interface{} {
	verif := verifiedState(&state)
	if verif.Err != nil {
		resp := u.Message(false, fmt.Sprintf("Data didn't pass verification. IDevice: %v", state.IDevice))
		resp["result"] = verif.SumResult
		return resp
	}
	resp := u.Message(true, "Data is correct")
	resp["result"] = verif.SumResult
	return resp
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
	tcpConnect.StateChan <- stateMessage
	for {
		chanRespond := <-tcpConnect.StateChan
		if strings.Contains(chanRespond.User, mapContx["login"]) {
			if chanRespond.Message == "ok" {
				return u.Message(true, fmt.Sprintf("Cross data deleted. Info (%v)", chanRespond.Info))
			} else {
				return u.Message(false, "TCP Server not responding")
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
func verifiedState(cross *agS_pudge.Cross) (result stateVerified.StateResult) {
	resultDay := stateVerified.DaySetsVerified(&cross.Arrays.DaySets)
	appendResult(&result, resultDay)
	resultWeek, empty := stateVerified.WeekSetsVerified(cross)
	appendResult(&result, resultWeek)
	resultMouth := stateVerified.MouthSetsVerified(cross, empty)
	appendResult(&result, resultMouth)
	resultTimeUse := stateVerified.TimeUseVerified(cross)
	appendResult(&result, resultTimeUse)
	return
}

//appendResult накапливание результатов верификации
func appendResult(mainRes *stateVerified.StateResult, addResult stateVerified.StateResult) {
	mainRes.SumResult = append(mainRes.SumResult, addResult.SumResult...)
	if addResult.Err != nil {
		mainRes.Err = addResult.Err
	}
}
