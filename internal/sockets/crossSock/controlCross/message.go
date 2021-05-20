package controlCross

import agspudge "github.com/ruraomsk/ag-server/pudge"

var (
	typeError      = "error"
	typeClose      = "close"
	typeDButton    = "dispatch"
	typeChangeEdit = "changeEdit"

	typeSendB        = "sendB"
	typeCheckB       = "checkB"
	typeCreateB      = "createB"
	typeDeleteB      = "deleteB"
	typeUpdateB      = "updateB"
	typeEditInfoB    = "editInfoB"
	typeRepaintCheck = "repaintCheck"
	typeHistory      = "history"
	typeGetHistory   = "getHistory"
	typeSendHistory  = "sendHistory"
	typeDiff         = "diff"

	typeControlBuild = "controlInfo"

	typeNotEdit            = "вам не разрешено редактировать данный перекресток"
	errParseType           = "Сервер не смог обработать запрос"
	errDoubleOpeningDevice = "запрашиваемый перекресток уже открыт"
	errCrossDoesntExist    = "запрашиваемый перекресток не существует"
)

//ControlSokResponse структура для отправки сообщений (cross control)
type ControlSokResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

//StateHandler структура приема / отправки state
type StateHandler struct {
	Type    string         `json:"type"`
	State   agspudge.Cross `json:"state"`
	RePaint bool           `json:"rePaint"`
	Z       int            `json:"z"`
}

//newCrossMess создание нового сообщения
func newControlMess(mType string, data map[string]interface{}) ControlSokResponse {
	var resp ControlSokResponse
	resp.Type = mType
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}
