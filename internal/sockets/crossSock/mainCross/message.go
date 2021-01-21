package mainCross

var (
	typeError           = "error"
	typeClose           = "close"
	typeDButton         = "dispatch"
	typeChangeEdit      = "changeEdit"
	typeCrossBuild      = "crossBuild"
	typePhase           = "phase"
	typeCrossUpdate     = "crossUpdate"
	typeCrossConnection = "crossConnection"
	typeStateChange     = "stateChange"
	typeEditCrossUsers  = "editCrossUsers"

	errParseType = "Сервер не смог обработать запрос"

	errDoubleOpeningDevice = "запрашиваемый перекресток уже открыт"
	errCrossDoesntExist    = "запрашиваемый перекресток не существует"
)

//crossResponse структура для отправки сообщений (cross)
type crossResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
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

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}
