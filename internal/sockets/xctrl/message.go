package xctrl

var (
	typeCPStart        = "CPStart"
	typeError          = "error"
	typeClose          = "close"
	typeCustInfo       = "custInfo"
	typeCustUpdate     = "custUpdate"
	typeCreateCustomer = "createCustomer"
	typeDeleteCustomer = "deleteCustomer"
	typeUpdateCustomer = "updateCustomer"
	errParseType       = "сервер не смог обработать запрос"
)

type CPMess struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func newCPMess(mType string, data map[string]interface{}) CPMess {
	var resp CPMess
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
