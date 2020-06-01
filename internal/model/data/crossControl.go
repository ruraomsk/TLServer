package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JanFant/TLServer/internal/model/crossEdit"

	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/stateVerified"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
)

//ControlGetCrossInfo сбор информации для пользователя в расширенном варианте
func ControlGetCrossInfo(TLignt TrafficLights) u.Response {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)
	sqlStr = fmt.Sprintf("SELECT area, subarea, idevice, dgis, describ, state FROM public.cross WHERE region = %v AND id = %v AND area = %v", TLignt.Region.Num, TLignt.ID, TLignt.Area.Num)
	rowsTL := GetDB().QueryRow(sqlStr)
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &StateStr)
	if err != nil {
		logger.Error.Println("|Message: No result at these points", err.Error())
		return u.Message(http.StatusInternalServerError, "no result at these points")
	}
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(StateStr)
	if err != nil {
		logger.Error.Println("|Message: Failed to parse cross information: ", err.Error())
		return u.Message(http.StatusInternalServerError, "failed to parse cross information")
	}
	TLignt.Points.StrToFloat(dgis)
	resp := u.Message(http.StatusOK, "cross control information")
	TLignt.Sost.Num = rState.StatusDevice
	CacheInfo.Mux.Lock()
	TLignt.Region.NameRegion = CacheInfo.MapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.MapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	TLignt.Sost.Description = CacheInfo.MapTLSost[TLignt.Sost.Num]
	resp.Obj["areaMap"] = CacheInfo.MapArea[TLignt.Region.NameRegion]
	CacheInfo.Mux.Unlock()
	resp.Obj["cross"] = TLignt
	resp.Obj["state"] = rState
	return resp
}

//ControlEditableCheck проверка редактируется ли данный перекресток
func ControlEditableCheck(arm crossEdit.BusyArm, mapContx map[string]string) u.Response {
	var EditInfo crossEdit.EditCrossInfo
	crossEdit.BusyArmInfo.Mux.Lock()
	defer crossEdit.BusyArmInfo.Mux.Unlock()
	if _, ok := crossEdit.BusyArmInfo.MapBusyArm[arm]; !ok {
		EditInfo.Login = mapContx["login"]
		EditInfo.EditFlag = true
		EditInfo.Kick = false
		EditInfo.Time = time.Now()
		crossEdit.BusyArmInfo.MapBusyArm[arm] = EditInfo
	} else {
		EditInfo = crossEdit.BusyArmInfo.MapBusyArm[arm]
		if EditInfo.Kick {
			delete(crossEdit.BusyArmInfo.MapBusyArm, arm)
		} else {
			if crossEdit.BusyArmInfo.MapBusyArm[arm].Login == mapContx["login"] {
				EditInfo.EditFlag = true
				EditInfo.Time = time.Now()
				crossEdit.BusyArmInfo.MapBusyArm[arm] = EditInfo
			} else {
				EditInfo.EditFlag = false
				EditInfo.Kick = false
			}
			if crossEdit.BusyArmInfo.MapBusyArm[arm].Time.Add(time.Second * 7).Before(time.Now()) {
				EditInfo.Login = mapContx["login"]
				EditInfo.EditFlag = true
				EditInfo.Kick = false
				EditInfo.Time = time.Now()
				crossEdit.BusyArmInfo.MapBusyArm[arm] = EditInfo
			}
		}
	}
	resp := u.Message(http.StatusOK, "editable flag")
	resp.Obj["DontWrite"] = "true"
	resp.Obj["editInfo"] = EditInfo
	return resp
}

//SendCrossData получение данных от пользователя проверка и отправка серверу(устройств)
func SendCrossData(state agS_pudge.Cross, mapContx map[string]string) u.Response {
	var (
		stateMessage tcpConnect.StateMessage
		err          error
		verif        stateVerified.StateResult
		userCross    agS_pudge.UserCross
	)
	verifiedState(&state, &verif)
	if verif.Err != nil {
		resp := u.Message(http.StatusOK, fmt.Sprintf("data didn't pass verification. IDevice: %v", state.IDevice))
		resp.Obj["result"] = verif.SumResult
		return resp
	}
	userCross.State = state
	userCross.User = mapContx["login"]
	stateMessage.StateStr, err = stateMarshal(userCross)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal state information: ", err.Error())
		return u.Message(http.StatusInternalServerError, "failed to Marshal state information")
	}
	stateMessage.User = mapContx["login"]
	if sendToUDPServer(stateMessage) {
		return u.Message(http.StatusOK, "cross send to server")
	} else {
		return u.Message(http.StatusInternalServerError, "TCP Server not responding")
	}
}

//CheckCrossData проверка полученных данных
func CheckCrossData(state agS_pudge.Cross) u.Response {
	var verif stateVerified.StateResult
	verifiedState(&state, &verif)
	if verif.Err != nil {
		resp := u.Message(http.StatusOK, fmt.Sprintf("data didn't pass verification. IDevice: %v", state.IDevice))
		resp.Obj["status"] = false
		resp.Obj["result"] = verif.SumResult
		return resp
	}
	resp := u.Message(http.StatusOK, "Data is correct")
	resp.Obj["status"] = true
	resp.Obj["result"] = verif.SumResult
	return resp
}

//CreateCrossData добавление нового перекрестка
func CreateCrossData(state agS_pudge.Cross, mapContx map[string]string) u.Response {
	var (
		stateMessage tcpConnect.StateMessage
		verif        stateVerified.StateResult
		userCross    agS_pudge.UserCross
		err          error
		stateSql     string
	)
	sqlStr := fmt.Sprintf(`SELECT state FROM public.cross WHERE state::jsonb @> '{"idevice":%v}'::jsonb OR (region = %v and area = %v and id = %v)`, state.IDevice, state.Region, state.Area, state.ID)
	rows, err := GetDB().Query(sqlStr)
	if err != nil {
		resp := u.Message(http.StatusInternalServerError, "DB not respond")
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
		resp := u.Message(http.StatusOK, fmt.Sprintf("data didn't pass verification. IDevice: %v", state.IDevice))
		resp.Obj["result"] = verif.SumResult
		return resp
	}
	userCross.State = state
	userCross.User = mapContx["login"]
	stateMessage.StateStr, err = stateMarshal(userCross)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal state information: ", err.Error())
		return u.Message(http.StatusInternalServerError, "failed to Marshal state information")
	}
	stateMessage.User = mapContx["login"]
	if sendToUDPServer(stateMessage) {
		if ShortCreateDirPng(state.Region, state.Area, state.ID, state.Dgis) {
			return u.Message(http.StatusOK, "cross created")
		} else {
			return u.Message(http.StatusOK, "cross created without Map.png - contact admin")
		}
	} else {
		return u.Message(http.StatusInternalServerError, "TCP Server not responding")
	}
}

//DeleteCrossData удаление перекрестка на сервере
func DeleteCrossData(state agS_pudge.Cross, mapContx map[string]string) u.Response {
	var (
		stateMessage tcpConnect.StateMessage
		userCross    agS_pudge.UserCross
		err          error
	)
	stateMessage.Info = fmt.Sprintf("idevice: %v, position : %v//%v//%v", state.IDevice, state.Region, state.Area, state.ID)
	state.IDevice = -1
	userCross.State = state
	userCross.User = mapContx["login"]
	stateMessage.StateStr, err = stateMarshal(userCross)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal state information: ", err.Error())
		return u.Message(http.StatusInternalServerError, "failed to Marshal state information")
	}
	stateMessage.User = mapContx["login"]
	if sendToUDPServer(stateMessage) {
		return u.Message(http.StatusOK, fmt.Sprintf("cross data deleted. Info (%v)", stateMessage.Info))
	} else {
		return u.Message(http.StatusInternalServerError, "TCP Server not responding")
	}
}

func TestCrossStateData(mapContx map[string]string) u.Response {
	var (
		stateSql  string
		stateInfo []crossEdit.BusyArm
		state     crossEdit.BusyArm
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
