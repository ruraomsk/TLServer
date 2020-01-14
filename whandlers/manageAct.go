package whandlers

import (
	"../data"
	"../logger"
	u "../utils"
	"encoding/json"
	"fmt"
	"net/http"
)

var ActParser = func(w http.ResponseWriter, r *http.Request) {
	mapContx := data.ParserInterface(r.Context().Value("info"))

	flag, resp := FuncAccessCheak(w, r, "ManageAccount")
	if flag {
		switch mapContx["act"] {
		case "update":
			fmt.Println("update")
		case "delete":
			fmt.Println("delete")
		case "add":
			{
				var shortAcc = &data.ShortAccount{}
				err := json.NewDecoder(r.Body).Decode(shortAcc)
				if err != nil {
					logger.Info.Println("ActParser: Incorrectly filled data ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, "Incorrectly filled data"))
					return
				}
				if mapContx["login"] == "RegAdmin" {
					if shortAcc.Role == "Admin" {
						u.Respond(w, r, u.Message(false, "No authority to create admins"))
						return
					}
				}
				account, privilege := shortAcc.ConvertShortToAcc()
				resp = account.Create(privilege)
			}
		case "changepw":
			{

			}
		default:
			{
				logger.Info.Println("ManageAct: Unregistered action", r.RemoteAddr)
				w.WriteHeader(http.StatusBadRequest)
				u.Respond(w, r, u.Message(false, "Unregistered action"))
				return
			}
		}
	}
	u.Respond(w, r, resp)
}
