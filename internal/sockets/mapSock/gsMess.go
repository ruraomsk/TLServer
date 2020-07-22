package mapSock

import (
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
)

var (
	typeCreateRout   = "createRoute"
	typeUpdateRout   = "updateRoute"
	typeDeleteRout   = "deleteRoute"
	typeDButton      = "dispatch"
	errCantWriteInBD = "Запись в БД не удалась"
)

//GSSokResponse структура для отправки сообщений (GS)
type GSSokResponse struct {
	Type string                 `json:"type"` //тип сообщения
	Data map[string]interface{} `json:"data"` //данные
	conn *websocket.Conn        `json:"-"`    //соединение
}

//newGSMess создание нового сообщения
func newGSMess(mType string, conn *websocket.Conn, data map[string]interface{}) GSSokResponse {
	var resp GSSokResponse
	resp.Type = mType
	resp.conn = conn
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//send отправка сообщения с обработкой ошибки
func (m *GSSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", m.conn.RemoteAddr(), "map socket", "/map", m.Data["message"])
		}()
	}
	writeGS <- *m
}

//closeMessage структура для закрытия
type closeMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
