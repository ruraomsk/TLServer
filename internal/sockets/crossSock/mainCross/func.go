package mainCross

import (
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/jmoiron/sqlx"
)

//phaseInfo инофрмация о фазах
type phaseInfo struct {
	idevice int  //идентификатор утройства
	Fdk     int  `json:"fdk"` //фаза
	Tdk     int  `json:"tdk"` //время обработки
	Pdk     bool `json:"pdk"` //переходный период
}

//get запрос фазы из базы
func (p *phaseInfo) get(db *sqlx.DB) error {
	err := db.QueryRow(`SELECT fdk, tdk, pdk FROM public.devices WHERE id = $1`, p.idevice).Scan(&p.Fdk, &p.Tdk, &p.Pdk)
	if err != nil {
		return err
	}
	return nil
}

//takeCrossInfo формарование необходимой информации о перекрестке
func takeCrossInfo(pos sockets.PosInfo, db *sqlx.DB) (resp crossResponse, idev int, description string) {
	var (
		dgis     string
		stateStr string
		phase    phaseInfo
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
