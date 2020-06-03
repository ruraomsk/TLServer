package data

import (
	"encoding/json"
	"fmt"
)

type mapMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func (m *mapMessage) send(mType string, mData map[string]interface{}, ch chan mapMessage) {
	m.Type = mType
	m.Data = mData
	ch <- *m
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

func (e *ErrorMessage) toString() string {
	raw, _ := json.Marshal(e)
	return string(raw)
}

var (
	typeError   = "error"
	typeUpdate  = "update"
	typeStatus  = "status"
	typeJump    = "jump"
	typeMapInfo = "mapInfo"
	typeNewBox  = "newBox"
	typeTFlight = "tflight"
	closeSocket = "closeSocket"

	errNoAccessWithDatabase    = "no access with database"
	errCantConvertJSON         = "cant convert JSON"
	errUnregisteredMessageType = "unregistered message type"
)
