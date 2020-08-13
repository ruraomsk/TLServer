package xctrl

var (
	typeXctrlChange = "xctrlChange"
	typeXctrlInfo   = "xctrlInfo"
	typeXctrlUpdate = "xctrlUpdate"
	typeError       = "error"
	typeClose       = "close"
	typeGetSubArea  = "getSubArea"

	errParseType   = "Сервер не смог обработать запрос"
	errGetXctrl    = "Ошибка запроса данных из БД"
	errChangeXctrl = "Ошибка записи данных в БД"
)

//MessXctrl структура пакета сообщения для xctrl
type MessXctrl struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

//newXctrlMess создание
func newXctrlMess(mType string, data map[string]interface{}) MessXctrl {
	var resp MessXctrl
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
