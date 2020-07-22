package crossSock

import (
	"fmt"

	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

var (
	typeSendB     = "sendB"
	typeCheckB    = "checkB"
	typeCreateB   = "createB"
	typeDeleteB   = "deleteB"
	typeUpdateB   = "updateB"
	typeEditInfoB = "editInfoB"

	typeControlBuild = "controlInfo"
	typeNotEdit      = "вам не разрешено редактировать данный перекресток"
)

//ControlSokResponse структура для отправки сообщений (cross control)
type ControlSokResponse struct {
	Type string                 `json:"type"` //тип сообщения
	Data map[string]interface{} `json:"data"` //данные
	conn *websocket.Conn        `json:"-"`    //соединение
	info CrossInfo              `json:"-"`    //информация о соединении
}

//StateHandler структура приема / отправки state
type StateHandler struct {
	Type  string         `json:"type"`
	State agspudge.Cross `json:"state"`
}

//send отправка с обработкой ошибки
func (m *ControlSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v",
				m.conn.RemoteAddr(),
				m.info.Login,
				fmt.Sprintf("/cross/control?Region=%v&Area=%v&ID=%v", m.info.Pos.Region, m.info.Pos.Area, m.info.Pos.Id),
				m.Data["message"])
		}()
	}
	writeControlMessage <- *m
}

//newControlMess создание нового сообщения
func newControlMess(mType string, conn *websocket.Conn, data map[string]interface{}, info CrossInfo) ControlSokResponse {
	var resp = ControlSokResponse{Type: mType, conn: conn, info: info}
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}
