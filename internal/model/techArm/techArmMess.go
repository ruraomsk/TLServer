package techArm

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/ag-server/pudge"
)

var (
	typeError                  = "error"
	typeClose                  = "close"
	typeArmInfo                = "armInfo"
	typeCrosses                = "crosses"
	typeDevices                = "devices"
	errUnregisteredMessageType = "unregistered message type"
)

//armResponse структура для отправки сообщений (map)
type armResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
}

//newMapMess создание нового сообщения
func newArmMess(mType string, conn *websocket.Conn, data map[string]interface{}) armResponse {
	var resp armResponse
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
func (m *armResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", m.conn.RemoteAddr(), "arm socket", "/techArm", m.Data["message"])
		}()
	}
	writeArm <- *m
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

//setTypeMessage определение типа сообщения
func setTypeMessage(raw []byte) (string, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(raw, &temp); err != nil {
		return "", err
	}
	return fmt.Sprint(temp["type"]), nil
}

//ArmInfo информация о запрашиваемом арме
type ArmInfo struct {
	Region int      `json:"region"` //регион
	Area   []string `json:"area"`   //район
}

type CrossInfo struct {
	Region    int    `json:"region"`
	Area      int    `json:"area"`
	ID        int    `json:"id"`
	Idevice   int    `json:"idevice"`
	Subarea   int    `json:"subarea"`
	ArrayType int    `json:"arrayType"`
	Describe  string `json:"describe"`
	Phone     string `json:"phone"`
}

type DevInfo struct {
	Region  int              `json:"region"`
	Area    int              `json:"area"`
	Idevice int              `json:"idevice"`
	Device  pudge.Controller `json:"device"`
}
