package crossSock

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/sockets"

	"github.com/jmoiron/sqlx"

	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
)

var (
	typeError          = "error"
	typeClose          = "close"
	typeDButton        = "dispatch"
	typeChangeEdit     = "changeEdit"
	typeCrossBuild     = "crossBuild"
	typePhase          = "phase"
	typeCrossUpdate    = "crossUpdate"
	typeStateChange    = "stateChange"
	typeEditCrossUsers = "editCrossUsers"

	errDoubleOpeningDevice     = "запрашиваемый перекресток уже открыт"
	errCrossDoesntExist        = "запрашиваемый перекресток не существует"
	errUnregisteredMessageType = "unregistered message type"
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
	Login       string          `json:"login"`       //пользователь
	Role        string          `json:"-"`           //роль
	Edit        bool            `json:"edit"`        //признак редактирования
	Idevice     int             `json:"idevice"`     //идентификатор утройства
	Description string          `json:"description"` //описание
	Pos         sockets.PosInfo `json:"pos"`         //расположение перекрестка
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

//phaseInfo инофрмация о фазах
type phaseInfo struct {
	idevice int  `json:"-"`   //идентификатор утройства
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

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

//closeMessage структура для закрытия
type closeMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
