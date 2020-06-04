package data

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

type mapResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
}

func mapMessage(mType string, conn *websocket.Conn) mapResponse {
	return mapResponse{Type: mType, conn: conn, Data: map[string]interface{}{}}
}

func (m *mapResponse) send() {
	WriteMap <- *m
}

func setTypeMessage(raw []byte) (string, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(raw, &temp); err != nil {
		return "", err
	}
	return fmt.Sprint(temp["type"]), nil
}

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
