package techArm

import (
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
	//modeRDK мапа состояний ДК
	modeRDK = map[int]string{
		1: "РУ",
		2: "РУ",
		3: "ЗУ",
		4: "ДУ",
		5: "ДУ",
		6: "ЛУ",
		8: "ЛУ",
		9: "КУ",
	}
	texMode = map[int]string{
		1:  "выбор ПК по времени по суточной карте ВР-СК",
		2:  "выбор ПК по недельной карте ВР-НК",
		3:  "выбор ПК по времени по суточной карте, назначенной оператором ДУ-СК",
		4:  "выбор ПК по недельной карте, назначенной оператором ДУ-НК",
		5:  "план по запросу оператора ДУ-ПК",
		6:  "резервный план (отсутствие точного времени) РП",
		7:  "коррекция привязки с ИП",
		8:  "коррекция привязки с сервера",
		9:  "выбор ПК по годовой карте",
		10: "выбор ПК по ХТ",
		11: "выбор ПК по картограмме",
		12: "противозаторовое управление",
	}
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

//ArmInfo информация о запрашиваемом арме
type ArmInfo struct {
	Login  string   `json:"login"`  //логин
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
	TexMode string           `json:"texMode"`
	ModeRdk string           `json:"modeRdk"`
	Device  pudge.Controller `json:"device"`
}
