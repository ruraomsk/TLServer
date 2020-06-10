package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
)

//MapSokResponse структура для отправки сообщений (map)
type MapSokResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
}

//newMapMess создание нового сообщения
func newMapMess(mType string, conn *websocket.Conn, data map[string]interface{}) MapSokResponse {
	var resp MapSokResponse
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
func (m *MapSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", m.conn.RemoteAddr(), "map socket", "/map", m.Data["message"])
		}()
	}
	writeMap <- *m
}

//setTypeMessage определение типа сообщения
func setTypeMessage(raw []byte) (string, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(raw, &temp); err != nil {
		return "", err
	}
	return fmt.Sprint(temp["type"]), nil
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

var (
	typeError                  = "error"
	typeJump                   = "jump"
	typeMapInfo                = "mapInfo"
	typeTFlight                = "tflight"
	typeLogin                  = "login"
	typeLogOut                 = "logOut"
	errNoAccessWithDatabase    = "no access with database"
	errCantConvertJSON         = "cant convert JSON"
	errUnregisteredMessageType = "unregistered message type"
)
