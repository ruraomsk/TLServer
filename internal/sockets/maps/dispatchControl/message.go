package dispatchControl

var (
	typeError   = "error"
	typeClose   = "close"
	typeMapInfo = "mapInfo"
	typeJump    = "jump"
	typeRepaint = "repaint"
	typeTFlight = "tflight"
	typePhases  = "phases"
	typeRoute   = "route"

	typeDButton = "dispatch"

	errParseType = "Сервер не смог обработать запрос"
)

//dCResponse структура для отправки сообщений (GS)
type dCResponse struct {
	Type string                 `json:"type"` //тип сообщения
	Data map[string]interface{} `json:"data"` //данные
}

//newDCMess создание нового сообщения
func newDCMess(mType string, data map[string]interface{}) dCResponse {
	var resp dCResponse
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
