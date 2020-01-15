package whandlers

import (
	"../data"
	"../logger"
	u "../utils"
	"encoding/json"
	"net/http"
)

var ActParser = func(w http.ResponseWriter, r *http.Request) {
	mapContx := data.ParserInterface(r.Context().Value("info"))

	flag, resp := FuncAccessCheak(w, r, "ManageAccount")
	if flag {
		switch mapContx["act"] {
		case "update":
			{
				var shortAcc = &data.ShortAccount{}
				err := json.NewDecoder(r.Body).Decode(shortAcc)
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, "Incorrectly filled data"))
					return
				}
				err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, err.Error()))
					return
				}
				account, privilege := shortAcc.ConvertShortToAcc()
				resp = account.Update(privilege)
			}
		case "delete":
			{
				var shortAcc = &data.ShortAccount{}
				err := json.NewDecoder(r.Body).Decode(shortAcc)
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, "Incorrectly filled data"))
					return
				}
				account, err := shortAcc.ValidDelete(mapContx["role"], mapContx["region"])
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, err.Error()))
					return
				}
				resp = account.Delete()
			}
		case "add":
			{
				var shortAcc = &data.ShortAccount{}
				err := json.NewDecoder(r.Body).Decode(shortAcc)
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, "Incorrectly filled data"))
					return
				}
				err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, err.Error()))
					return
				}

				account, privilege := shortAcc.ConvertShortToAcc()
				resp = account.Create(privilege)
			}
		case "changepw":
			{
				var shortAcc = &data.ShortAccount{}
				err := json.NewDecoder(r.Body).Decode(shortAcc)
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, "Incorrectly filled data"))
					return
				}
				account, err := shortAcc.ValidChangePW(mapContx["role"], mapContx["region"])
				if err != nil {
					logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
					w.WriteHeader(http.StatusBadRequest)
					u.Respond(w, r, u.Message(false, err.Error()))
					return
				}
				resp = account.ChangePW()
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