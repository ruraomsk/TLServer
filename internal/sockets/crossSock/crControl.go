package crossSock

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/crossCreator"
	"github.com/JanFant/TLServer/internal/model/deviceLog"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"
	"strings"

	"github.com/JanFant/TLServer/internal/model/stateVerified"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

//TestCrossStateData проверить все стрейты на наличие ошибок
func TestCrossStateData(mapContx map[string]string, db *sqlx.DB) u.Response {
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
	rows, err := db.Query(sqlStr)
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
func takeControlInfo(pos sockets.PosInfo, db *sqlx.DB) (resp ControlSokResponse, idev int) {
	var (
		stateStr string
	)
	rowsTL := db.QueryRow(`SELECT state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil, nil, CrossInfo{})
		resp.Data["message"] = "No result at these points, table cross"
		return resp, 0
	}
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil, nil, CrossInfo{})
		resp.Data["message"] = "failed to parse cross information"
		return resp, 0
	}
	resp = newControlMess(typeControlBuild, nil, nil, CrossInfo{})
	resp.Data["state"] = rState
	return resp, rState.IDevice
}

//checkCrossData проверка полученных данных
func checkCrossData(state agspudge.Cross) ControlSokResponse {
	var verif stateVerified.StateResult
	verifiedState(&state, &verif)
	resp := newControlMess(typeCheckB, nil, nil, CrossInfo{})
	if verif.Err != nil {
		resp.Data["status"] = false
	} else {
		resp.Data["status"] = true
	}
	resp.Data["result"] = verif.SumResult
	return resp
}

//createCrossData добавление нового перекрестка
func createCrossData(state agspudge.Cross, pos sockets.PosInfo, login string, z int, db *sqlx.DB) map[string]interface{} {
	var (
		userCross = agspudge.UserCross{User: login, State: state}
		mess      = tcpConnect.TCPMessage{
			User:        login,
			TCPType:     tcpConnect.TypeState,
			Idevice:     userCross.State.IDevice,
			Data:        userCross,
			From:        tcpConnect.FromCrControlSoc,
			CommandType: typeCreateB,
			Pos:         pos,
		}
		verRes   []string
		stateSql string
	)
	sqlStr := fmt.Sprintf(`SELECT state FROM public.cross WHERE state::jsonb @> '{"Idevice":%v}'::jsonb OR (region = %v and area = %v and id = %v)`, state.IDevice, state.Region, state.Area, state.ID)
	rows, err := db.Query(sqlStr)
	if err != nil {
		logger.Error.Println("|Message: control socket (create Button), DB not respond : ", err.Error())
		resp := make(map[string]interface{})
		resp["status"] = false
		resp["message"] = "DB not respond"
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
		resp := make(map[string]interface{})
		resp["status"] = false
		resp["result"] = verRes
		return resp
	}

	mess.SendToTCPServer()

	resp := make(map[string]interface{})
	if crossCreator.ShortCreateDirPng(state.Region, state.Area, state.ID, z, state.Dgis) {
		resp["message"] = "cross created"
	} else {
		resp["message"] = "cross created without Map.png - contact admin"
	}
	return resp
}

//verifiedState набор проверкок для стейта
func verifiedState(cross *agspudge.Cross, result *stateVerified.StateResult) {
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
