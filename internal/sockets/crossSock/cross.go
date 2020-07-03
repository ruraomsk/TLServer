package crossSock

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/comm"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"strings"
)

//takeCrossInfo формарование необходимой информации о перекрестке
func takeCrossInfo(pos PosInfo, db *sqlx.DB) (resp CrossSokResponse, idev int, desc string) {
	var (
		dgis     string
		stateStr string
		phase    phaseInfo
	)
	TLignt := data.TrafficLights{Area: data.AreaInfo{Num: pos.Area}, Region: data.RegionInfo{Num: pos.Region}, ID: pos.Id}
	rowsTL := db.QueryRow(`SELECT area, subarea, Idevice, dgis, describ, state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &stateStr)
	if err != nil {
		resp := newCrossMess(typeError, nil, nil, CrossInfo{})
		resp.Data["message"] = "No result at these points, table cross"
		return resp, 0, ""
	}
	TLignt.Points.StrToFloat(dgis)
	//Состояние светофора!
	rState, err := convertStateStrToStruct(stateStr)
	if err != nil {
		resp := newCrossMess(typeError, nil, nil, CrossInfo{})
		resp.Data["message"] = "failed to parse cross information"
		return resp, 0, ""
	}

	resp = newCrossMess(typeCrossBuild, nil, nil, CrossInfo{})
	data.CacheInfo.Mux.Lock()
	TLignt.Region.NameRegion = data.CacheInfo.MapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = data.CacheInfo.MapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	TLignt.Sost.Num = rState.StatusDevice
	TLignt.Sost.Description = data.CacheInfo.MapTLSost[TLignt.Sost.Num]
	data.CacheInfo.Mux.Unlock()
	phase.idevice = TLignt.Idevice
	err = phase.get(db)
	if err != nil {
		resp.Data["phase"] = phaseInfo{}
	} else {
		resp.Data["phase"] = phase
	}
	resp.Data["cross"] = TLignt
	resp.Data["state"] = rState
	resp.Data["region"] = TLignt.Region.Num
	return resp, TLignt.Idevice, TLignt.Description
}

//getNewState получение обновленного state
func getNewState(pos PosInfo, db *sqlx.DB) (agspudge.Cross, error) {
	var stateStr string
	rowsTL := db.QueryRow(`SELECT state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	_ = rowsTL.Scan(&stateStr)
	rState, err := convertStateStrToStruct(stateStr)
	if err != nil {
		return agspudge.Cross{}, err
	}
	return rState, nil
}

//dispatchControl отправка команды на устройство
func dispatchControl(arm comm.CommandARM) CrossSokResponse {
	var (
		err        error
		armMessage tcpConnect.ArmCommandMessage
	)

	armMessage.CommandStr, err = armControlMarshal(arm)
	if err != nil {
		resp := newCrossMess(typeError, nil, nil, CrossInfo{})
		resp.Data["message"] = "failed to Marshal ArmControlData information"
		return resp
	}
	armMessage.User = arm.User
	tcpConnect.ArmCommandChan <- armMessage
	for {
		chanRespond := <-tcpConnect.ArmCommandChan
		if strings.Contains(armMessage.User, arm.User) {
			if chanRespond.Message == "ok" {
				resp := newCrossMess(typeDButton, nil, nil, CrossInfo{})
				resp.Data["message"] = fmt.Sprintf("command %v send to server", armMessage.CommandStr)
				resp.Data["user"] = arm.User
				return resp
			} else {
				resp := newCrossMess(typeDButton, nil, nil, CrossInfo{})
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

//ConvertStateStrToStruct разбор данных (Cross) полученных из БД в нужную структуру
func convertStateStrToStruct(str string) (rState agspudge.Cross, err error) {
	if err := json.Unmarshal([]byte(str), &rState); err != nil {
		return rState, err
	}
	return rState, nil
}
