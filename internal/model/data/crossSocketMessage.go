package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/ag-server/comm"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"strings"

	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
)

//CrossSokResponse структура для отправки сообщений (cross)
type CrossSokResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
	info crossInfo              `json:"-"`
}

//crossInfo информация о перекрестке для которого открыт сокет
type crossInfo struct {
	login   string
	edit    bool
	idevice int
	pos     PosInfo
}

//PosInfo положение перекрестка
type PosInfo struct {
	Region string //регион
	Area   string //район
	Id     int    //ID
}

//newCrossMess создание нового сообщения
func newCrossMess(mType string, conn *websocket.Conn, data map[string]interface{}, info crossInfo) CrossSokResponse {
	var resp = CrossSokResponse{Type: mType, conn: conn, info: info}
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//send отправка с обработкой ошибки
func (m *CrossSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v",
				m.conn.RemoteAddr(),
				m.info.login,
				fmt.Sprintf("/cross?Region=%v&Area=%v&ID=%v", m.info.pos.Region, m.info.pos.Area, m.info.pos.Id),
				m.Data["message"])
		}()
	}
	writeCrossMessage <- *m
}

//phaseInfo инофрмация о фазах
type phaseInfo struct {
	idevice int  `json:"-"`
	Fdk     int  `json:"fdk"`
	Tdk     int  `json:"tdk"`
	Pdk     bool `json:"pdk"`
}

//get запрос фазы из базы
func (p *phaseInfo) get() error {
	err := GetDB().QueryRow(`SELECT Fdk, tdk, pdk FROM public.devices WHERE id = $1`, p.idevice).Scan(&p.Fdk, &p.Tdk, &p.Pdk)
	if err != nil {
		return err
	}
	return nil
}

//getNewState получение обновленного state
func getNewState(pos PosInfo) (agS_pudge.Cross, error) {
	var stateStr string
	rowsTL := GetDB().QueryRow(`SELECT state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	_ = rowsTL.Scan(&stateStr)
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		return agS_pudge.Cross{}, err
	}
	return rState, nil
}

//takeCrossInfo формарование необходимой информации о перекрестке
func takeCrossInfo(pos PosInfo) (CrossSokResponse, int) {
	var (
		dgis     string
		stateStr string
		phase    phaseInfo
	)
	TLignt := TrafficLights{Area: AreaInfo{Num: pos.Area}, Region: RegionInfo{Num: pos.Region}, ID: pos.Id}
	rowsTL := GetDB().QueryRow(`SELECT area, subarea, idevice, dgis, describ, state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &stateStr)
	if err != nil {
		resp := newCrossMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "No result at these points, table cross"
		return resp, 0
	}
	TLignt.Points.StrToFloat(dgis)
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := newCrossMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to parse cross information"
		return resp, 0
	}

	resp := newCrossMess(typeCrossBuild, nil, nil, crossInfo{})
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
	resp.Data["region"] = TLignt.Region.Num
	return resp, TLignt.Idevice
}

//DispatchControl отправка команды на устройство
func DispatchControl(arm comm.CommandARM) CrossSokResponse {
	var (
		err        error
		armMessage tcpConnect.ArmCommandMessage
	)

	armMessage.CommandStr, err = armControlMarshal(arm)
	if err != nil {
		resp := newCrossMess(typeError, nil, nil, crossInfo{})
		resp.Data["message"] = "failed to Marshal ArmControlData information"
		return resp
	}
	armMessage.User = arm.User
	tcpConnect.ArmCommandChan <- armMessage
	for {
		chanRespond := <-tcpConnect.ArmCommandChan
		if strings.Contains(armMessage.User, arm.User) {
			if chanRespond.Message == "ok" {
				resp := newCrossMess(typeDButton, nil, nil, crossInfo{})
				resp.Data["message"] = fmt.Sprintf("command %v send to server", armMessage.CommandStr)
				resp.Data["user"] = arm.User
				return resp
			} else {
				resp := newCrossMess(typeDButton, nil, nil, crossInfo{})
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

var (
	typeClose       = "close"
	typeDButton     = "dispatch"
	typeChangeEdit  = "changeEdit"
	typeCrossBuild  = "crossBuild"
	typePhase       = "phase"
	typeCrossUpdate = "crossUpdate"
	typeStateChange = "stateChange"

	errDoubleOpeningDevice      = "double opening device"
	errThereIsnSuchIntersection = "there isn such intersection"
)
