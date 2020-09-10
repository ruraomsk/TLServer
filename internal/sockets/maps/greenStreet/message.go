package greenStreet

var (
	typeError   = "error"
	typeClose   = "close"
	typeMapInfo = "mapInfo"
	typeJump    = "jump"
	typeRepaint = "repaint"
	typeTFlight = "tflight"

	typeRoute      = "route"
	typeCreateRout = "createRoute"
	typeUpdateRout = "updateRoute"
	typeDeleteRout = "deleteRoute"
	typeDButton    = "dispatch"

	errParseType = "Сервер не смог обработать запрос"
)

//gSResponse структура для отправки сообщений (GS)
type gSResponse struct {
	Type string                 `json:"type"` //тип сообщения
	Data map[string]interface{} `json:"data"` //данные
}

//newGSMess создание нового сообщения
func newGSMess(mType string, data map[string]interface{}) gSResponse {
	var resp gSResponse
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
