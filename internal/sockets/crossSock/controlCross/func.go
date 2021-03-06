package controlCross

import (
	"database/sql"
	"fmt"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/crossCreator"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/internal/model/stateVerified"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/crossSock"
	"github.com/ruraomsk/TLServer/logger"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"strings"
)

//takeControlInfo формарование необходимой информации о арме перекрестка
func takeControlInfo(pos sockets.PosInfo) (resp ControlSokResponse, idev int, desc string) {
	var (
		stateStr string
	)
	db, id := data.GetDB()
	defer data.FreeDB(id)
	rowsTL := db.QueryRow(`SELECT state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil)
		resp.Data["message"] = "No result at these points, table cross"
		return resp, 0, ""
	}
	//Состояние светофора!
	rState, err := crossSock.ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := newControlMess(typeError, nil)
		resp.Data["message"] = "failed to parse cross information"
		return resp, 0, ""
	}

	resp = newControlMess(typeControlBuild, nil)
	resp.Data["state"] = rState

	device.GlobalDevices.Mux.Lock()
	for _, c := range device.GlobalDevices.MapDevices {
		if c.Controller.ID == rState.IDevice {
			resp.Data["deviceIP"] = c.Controller.IPHost
			break
		} else {
			resp.Data["deviceIP"] = ""
		}
	}
	device.GlobalDevices.Mux.Unlock()

	return resp, rState.IDevice, rState.Name
}

//checkCrossData проверка полученных данных
func checkCrossData(state agspudge.Cross) ControlSokResponse {
	var verif stateVerified.StateResult
	crossSock.VerifiedState(&state, &verif)
	resp := newControlMess(typeCheckB, nil)
	if verif.Err != nil {
		resp.Data["status"] = false
	} else {
		resp.Data["status"] = true
	}
	resp.Data["result"] = verif.SumResult
	return resp
}

//createCrossData добавление нового перекрестка
func createCrossData(state agspudge.Cross, pos sockets.PosInfo, login string, z int) map[string]interface{} {
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
	db, id := data.GetDB()
	defer data.FreeDB(id)
	sqlStr := fmt.Sprintf(`SELECT state FROM public.cross WHERE state::jsonb @> '{"idevice":%v}'::jsonb OR (region = %v and area = %v and id = %v)`, state.IDevice, state.Region, state.Area, state.ID)
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
		if strings.Contains(stateSql, fmt.Sprintf(`"idevice": %v`, state.IDevice)) {
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
		resp["message"] = "перекресток создан"
	} else {
		resp["message"] = "перекресток создан без расположения - свяжитесь с Администратором"
	}
	return resp
}

func sendCrossData(state agspudge.Cross, cIDev int, pos sockets.PosInfo, login string) map[string]interface{} {
	var (
		userCross = agspudge.UserCross{User: login, State: state}
		mess      = tcpConnect.TCPMessage{
			User:        login,
			TCPType:     tcpConnect.TypeState,
			Idevice:     state.IDevice,
			Data:        userCross,
			From:        tcpConnect.FromCrControlSoc,
			CommandType: typeSendB,
			Pos:         pos,
		}
		resp = make(map[string]interface{})
	)
	db, id := data.GetDB()
	defer data.FreeDB(id)
	if cIDev != state.IDevice {
		var strRow string
		sqlStr := fmt.Sprintf(`SELECT state FROM public.cross WHERE state::jsonb @> '{"idevice":%v}'::jsonb`, state.IDevice)
		err := db.QueryRow(sqlStr).Scan(&strRow)
		if err != nil && err != sql.ErrNoRows {
			logger.Error.Println("|Message: control socket (send Button), DB not respond : ", err.Error())
			resp["status"] = false
			resp["message"] = "сервер баз данных не отвечает"
			return resp
		}
		if strings.Contains(strRow, fmt.Sprintf(`"idevice": %v`, state.IDevice)) {
			resp["status"] = false
			resp["message"] = fmt.Sprintf("№ %v модема уже используется в системе", state.IDevice)
			return resp
		}
	}
	mess.SendToTCPServer()

	return resp
}
