package data

import (
	"fmt"

	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
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
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
	info CrossInfo              `json:"-"`
}

//StateHandler структура приема / отправки state
type StateHandler struct {
	Type  string          `json:"type"`
	State agS_pudge.Cross `json:"state"`
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
