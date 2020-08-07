package mainCross

import (
	"github.com/JanFant/TLServer/internal/sockets"
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

	errParseType = "Сервер не смог обработать запрос"

	errDoubleOpeningDevice = "запрашиваемый перекресток уже открыт"
	errCrossDoesntExist    = "запрашиваемый перекресток не существует"
)

//crossResponse структура для отправки сообщений (cross)
type crossResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

//crossInfo информация о перекрестке для которого открыт сокет
type CrossInfo struct {
	Login   string          `json:"login"`   //пользователь
	Role    string          `json:"-"`       //роль
	Edit    bool            `json:"edit"`    //признак редактирования
	Idevice int             `json:"idevice"` //идентификатор утройства
	Pos     sockets.PosInfo `json:"pos"`     //расположение перекрестка
	ip      string
	region  string
}

//newCrossMess создание нового сообщения
func newCrossMess(mType string, data map[string]interface{}) crossResponse {
	var resp crossResponse
	resp.Type = mType
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//phaseInfo инофрмация о фазах
type phaseInfo struct {
	idevice int  `json:"-"`   //идентификатор утройства
	Fdk     int  `json:"fdk"` //фаза
	Tdk     int  `json:"tdk"` //время обработки
	Pdk     bool `json:"pdk"` //переходный период
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

type regStatus struct {
	ok      bool
	edit    bool
	idevice int
}
