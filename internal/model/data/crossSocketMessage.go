package data

import (
	"fmt"

	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
)

var (
	typeError       = "error"
	typeClose       = "close"
	typeDButton     = "dispatch"
	typeChangeEdit  = "changeEdit"
	typeCrossBuild  = "crossBuild"
	typePhase       = "phase"
	typeCrossUpdate = "crossUpdate"
	typeStateChange = "stateChange"

	errDoubleOpeningDevice = "запрашиваемый перекресток уже открыт"
	errCrossDoesntExist    = "запрашиваемый перекресток не существует"
)

//CrossSokResponse структура для отправки сообщений (cross)
type CrossSokResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
	info CrossInfo              `json:"-"`
}

//crossInfo информация о перекрестке для которого открыт сокет
type CrossInfo struct {
	Login       string  `json:"login"`
	Role        string  `json:"-"`
	Edit        bool    `json:"edit"`
	Idevice     int     `json:"idevice"`
	Description string  `json:"description"` //описание
	Pos         PosInfo `json:"pos"`
}

//PosInfo положение перекрестка
type PosInfo struct {
	Region string `json:"region"` //регион
	Area   string `json:"area"`   //район
	Id     int    `json:"id"`     //ID
}

//newCrossMess создание нового сообщения
func newCrossMess(mType string, conn *websocket.Conn, data map[string]interface{}, info CrossInfo) CrossSokResponse {
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
				m.info.Login,
				fmt.Sprintf("/cross?Region=%v&Area=%v&ID=%v", m.info.Pos.Region, m.info.Pos.Area, m.info.Pos.Id),
				m.Data["message"])
		}()
	}
	writeCrossMessage <- *m
}

//phaseInfo инофрмация о фазах3123123123
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

//formCrossUser сформировать пользователей которые редактируеют кросы
func formCrossUser() []CrossInfo {
	var temp = make([]CrossInfo, 0)
	for _, info := range crossConnect {
		if info.Edit {
			temp = append(temp, info)
		}
	}
	return temp
}
