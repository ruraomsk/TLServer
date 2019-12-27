package whandlers

import (
	"../data"
	u "../utils"
	"encoding/json"
	"net/http"
)

var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	_ = json.NewDecoder(r.Body).Decode(account)
	//str:= account.Privilege.ToSqlStrUpdate("accounts","MMM")
	//data.GetDB().Exec(str)
	//fmt.Println(str)
	u.Respond(w, r, u.Message(true, "Chil its ok"))
})

var TestToken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//resp := make(map[string]interface{})
	//num, err := strconv.Atoi(r.URL.Query().Get("Num"))
	//var flag bool
	//if num == 1 {
	//	flag, err = data.RoleCheck(data.ParserInterface(r.Context().Value("info")), "MakeTest")
	//
	//}
	//if num == 2 {
	//	flag, err = data.RoleCheck(data.ParserInterface(r.Context().Value("info")), "akeTest")
	//
	//}
	//if err != nil {
	//	resp["Test"] = err.Error()
	//	u.Respond(w, r, resp)
	//	return
	//}

	flag, resp := FuncAccessCheak(w, r, "MakeTest")
	if flag {
		resp["Test"] = "OK!"
	}

	u.Respond(w, r, resp)
})
