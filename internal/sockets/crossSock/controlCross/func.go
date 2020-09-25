package controlCross

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/crossCreator"
	"github.com/JanFant/TLServer/internal/model/stateVerified"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"strings"
)

//takeControlInfo формарование необходимой информации о арме перекрестка
func takeControlInfo(pos sockets.PosInfo, db *sqlx.DB) (resp ControlSokResponse, idev int, desc string) {
	var (
		stateStr string
	)
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
	return resp, rState.IDevice, rState.Name
}

//checkCrossData проверка полученных данных
func checkCrossData(state agspudge.Cross, db *sqlx.DB) ControlSokResponse {
	var verif stateVerified.StateResult
	crossSock.VerifiedState(&state, &verif, db)
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
		resp["message"] = "перекресток создан"
	} else {
		resp["message"] = "перекресток создан без красположения - свяжитесь с Администратором"
	}
	return resp
}
