package mainCross

import (
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/device"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/pudge"
)

//takeCrossInfo формарование необходимой информации о перекрестке
func takeCrossInfo(pos sockets.PosInfo, db *sqlx.DB) (resp crossResponse, idev int, description string) {
	var (
		dgis     string
		stateStr string
	)
	TLignt := data.TrafficLights{Area: data.AreaInfo{Num: pos.Area}, Region: data.RegionInfo{Num: pos.Region}, ID: pos.Id}
	rowsTL := db.QueryRow(`SELECT area, subarea, Idevice, dgis, describ, state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &stateStr)
	if err != nil {
		resp := newCrossMess(typeError, nil)
		resp.Data["message"] = "No result at these points, table cross"
		return resp, 0, ""
	}
	TLignt.Points.StrToFloat(dgis)
	//Состояние светофора!
	rState, err := crossSock.ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := newCrossMess(typeError, nil)
		resp.Data["message"] = "failed to parse cross information"
		return resp, 0, ""
	}

	resp = newCrossMess(typeCrossBuild, nil)
	data.CacheInfo.Mux.Lock()
	TLignt.Region.NameRegion = data.CacheInfo.MapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = data.CacheInfo.MapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	TLignt.Sost.Num = rState.StatusDevice
	TLignt.Sost.Description = data.CacheInfo.MapTLSost[TLignt.Sost.Num].Description
	TLignt.Sost.Control = data.CacheInfo.MapTLSost[TLignt.Sost.Num].Control
	data.CacheInfo.Mux.Unlock()

	device.GlobalDevices.Mux.Lock()
	dev, ok := device.GlobalDevices.MapDevices[TLignt.Idevice]
	device.GlobalDevices.Mux.Unlock()
	if ok {
		resp.Data["dk"] = dev.Controller.DK
	} else {
		resp.Data["dk"] = pudge.DK{}
	}

	resp.Data["cross"] = TLignt
	resp.Data["phases"] = rState.Arrays.SetDK.GetPhases()
	resp.Data["state"] = rState
	resp.Data["region"] = TLignt.Region.Num
	return resp, TLignt.Idevice, TLignt.Description
}
