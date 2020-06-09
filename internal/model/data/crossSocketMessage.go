package data

import (
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
)

type CrossSokResponse struct {
	Type   string                 `json:"type"`
	Data   map[string]interface{} `json:"data"`
	ccInfo CrossConn              `json:"-"`
}

//CrossConn соединение
type CrossConn struct {
	Pos   CrossEditInfo   //какой перекресток редактирует
	Conn  *websocket.Conn //подключение
	Edit  bool            //флаг редактирование
	Login string          //пользователь
}

//CrossEditInfo положение перекрестка
type CrossEditInfo struct {
	Region string //регион
	Area   string //район
	Id     int    //ID
}

func crossSokMessage(mType string, crConn CrossConn, data map[string]interface{}) CrossSokResponse {
	var resp CrossSokResponse
	resp.Type = mType
	resp.ccInfo = crConn
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

func (m *CrossSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v",
				m.ccInfo.Conn.RemoteAddr(),
				m.ccInfo.Login,
				fmt.Sprintf("/cross?Region=%v&Area=%v&ID=%v", m.ccInfo.Pos.Region, m.ccInfo.Pos.Area, m.ccInfo.Pos.Id),
				m.Data["message"])
		}()
	}
	WriteCrossMessage <- *m
}

type phaseInfo struct {
	idevice int  `json:"-"`
	Fdk     int  `json:"fdk"`
	Tdk     int  `json:"tdk"`
	Pdk     bool `json:"pdk"`
}

func (p *phaseInfo) get() error {
	err := GetDB().QueryRow(`SELECT Fdk, tdk, pdk FROM public.devices WHERE id = $1`, p.idevice).Scan(&p.Fdk, &p.Tdk, &p.Pdk)
	if err != nil {
		return err
	}
	return nil
}

func crossInfo(pos CrossEditInfo) CrossSokResponse {
	var (
		dgis     string
		stateStr string
		phase    phaseInfo
	)
	TLignt := TrafficLights{Area: AreaInfo{Num: pos.Area}, Region: RegionInfo{Num: pos.Region}, ID: pos.Id}
	rowsTL := GetDB().QueryRow(`SELECT area, subarea, idevice, dgis, describ, state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &stateStr)
	if err != nil {
		resp := crossSokMessage(typeError, CrossConn{}, nil)
		resp.Data["message"] = "No result at these points, table cross"
		return resp
	}
	TLignt.Points.StrToFloat(dgis)
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := crossSokMessage(typeError, CrossConn{}, nil)
		resp.Data["message"] = "failed to parse cross information"
		return resp
	}

	resp := crossSokMessage(typeCrossBuild, CrossConn{}, nil)
	CacheInfo.Mux.Lock()
	TLignt.Region.NameRegion = CacheInfo.MapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.MapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	TLignt.Sost.Num = rState.StatusDevice
	TLignt.Sost.Description = CacheInfo.MapTLSost[TLignt.Sost.Num]
	CacheInfo.Mux.Unlock()
	phase.idevice = TLignt.Idevice
	err = phase.get()
	if err != nil {
		resp.Data["phase"] = phaseInfo{}
	} else {
		resp.Data["phase"] = phase
	}
	resp.Data["cross"] = TLignt
	resp.Data["state"] = rState

	return resp
}

var (
	typeDButton            = "dispatch"
	typeChangeEdit         = "changeEdit"
	typeCrossBuild         = "crossBuild"
	errDoubleOpeningDevice = "double opening device"
)
