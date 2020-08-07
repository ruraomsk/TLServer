package mainMap

var (
	typeJump           = "jump"
	typeMapInfo        = "mapInfo"
	typeTFlight        = "tflight"
	typeRepaint        = "repaint"
	typeEditCrossUsers = "editCrossUsers"
	typeLogin          = "login"
	typeLogOut         = "logOut"
	typeChangeAccount  = "changeAcc"
	typeError          = "error"
	typeClose          = "close"
	typeCheckConn      = "checkConn"
	typeDButton        = "dispatch"

	errParseType = "Сервер не смог обработать запрос"
)

//MapSokResponse структура для отправки сообщений (map)
type mapResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

//newMapMess создание нового сообщения
func newMapMess(mType string, data map[string]interface{}) mapResponse {
	var resp mapResponse
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
