package whandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"../data"
	u "../utils"
)

//ActUpdateAccount обработчик запроса обновления (работа с пользователями)
var ActUpdateAccount = func(w http.ResponseWriter, r *http.Request) {
	mapContx := u.ParserInterface(r.Context().Value("info"))
	flag, resp := FuncAccessCheck(w, r, "ManageAccount")
	if flag {
		var shortAcc = &data.ShortAccount{}
		err := shortAcc.DecodeRequest(w, r)
		if err != nil {
			return
		}
		err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
		if err != nil {
			//logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
			return
		}
		account, privilege := shortAcc.ConvertShortToAcc()
		resp = account.Update(privilege)
	}
	u.Respond(w, r, resp)
}

//ActDeleteAccount обработчик запроса удаления (работа с пользователями)
var ActDeleteAccount = func(w http.ResponseWriter, r *http.Request) {
	mapContx := u.ParserInterface(r.Context().Value("info"))
	flag, resp := FuncAccessCheck(w, r, "ManageAccount")
	if flag {
		var shortAcc = &data.ShortAccount{}
		err := shortAcc.DecodeRequest(w, r)
		if err != nil {
			return
		}
		account, err := shortAcc.ValidDelete(mapContx["role"], mapContx["region"])
		if err != nil {
			//logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
			return
		}
		resp = account.Delete()
	}
	u.Respond(w, r, resp)
}

//ActAddAccount обработчик запроса добавления (работа с пользователями)
var ActAddAccount = func(w http.ResponseWriter, r *http.Request) {
	mapContx := u.ParserInterface(r.Context().Value("info"))
	flag, resp := FuncAccessCheck(w, r, "ManageAccount")
	if flag {
		var shortAcc = &data.ShortAccount{}
		err := shortAcc.DecodeRequest(w, r)
		if err != nil {
			return
		}
		err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
		if err != nil {
			//logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
			return
		}

		account, privilege := shortAcc.ConvertShortToAcc()
		resp = account.Create(privilege)
	}
	u.Respond(w, r, resp)
}

//ActChangePw обработчик запроса смены пароля для пользователя
var ActChangePw = func(w http.ResponseWriter, r *http.Request) {
	mapContx := u.ParserInterface(r.Context().Value("info"))
	flag, resp := FuncAccessCheck(w, r, "ActChangePw")
	if flag {
		var passChange = &data.PassChange{}
		err := json.NewDecoder(r.Body).Decode(passChange)
		if err != nil {
			//logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
			return
		}
		account, err := passChange.ValidOldNewPW(mapContx["login"])
		if err != nil {
			// logger.Info.Println("|Message: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, err.Error()))
			return
		}
		resp = account.ChangePW()
	}
	u.Respond(w, r, resp)
}

//ActParser обработчик запроса {act} update, delete, add работа с пользователем
//var ActParser = func(w http.ResponseWriter, r *http.Request) {
//	mapContx := u.ParserInterface(r.Context().Value("info"))
//	flag, resp := FuncAccessCheck(w, r, "ManageAccount")
//	if flag {
//		switch mapContx["act"] {
//		case "update":
//			{
//				var shortAcc = &data.ShortAccount{}
//				err := shortAcc.DecodeRequest(w, r)
//				if err != nil {
//					return
//				}
//				err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
//				if err != nil {
//					//logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
//					w.WriteHeader(http.StatusBadRequest)
//					u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
//					return
//				}
//				account, privilege := shortAcc.ConvertShortToAcc()
//				resp = account.Update(privilege)
//			}
//		case "delete":
//			{
//				var shortAcc = &data.ShortAccount{}
//				err := shortAcc.DecodeRequest(w, r)
//				if err != nil {
//					return
//				}
//				account, err := shortAcc.ValidDelete(mapContx["role"], mapContx["region"])
//				if err != nil {
//					//logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
//					w.WriteHeader(http.StatusBadRequest)
//					u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
//					return
//				}
//				resp = account.Delete()
//			}
//		case "add":
//			{
//				var shortAcc = &data.ShortAccount{}
//				err := shortAcc.DecodeRequest(w, r)
//				if err != nil {
//					return
//				}
//				err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
//				if err != nil {
//					//logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
//					w.WriteHeader(http.StatusBadRequest)
//					u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
//					return
//				}
//
//				account, privilege := shortAcc.ConvertShortToAcc()
//				resp = account.Create(privilege)
//			}
//		//case "changepw":
//		//	{
//		//		var shortAcc = &data.ShortAccount{}
//		//		err := shortAcc.DecodeRequest(w, r)
//		//		if err != nil {
//		//			return
//		//		}
//		//		account, err := shortAcc.ValidChangePW(mapContx["role"], mapContx["region"])
//		//		if err != nil {
//		//			logger.Info.Println("ActParser, Add: Incorrectly filled data: ", err.Error(), "  ", r.RemoteAddr)
//		//			w.WriteHeader(http.StatusBadRequest)
//		//			u.Respond(w, r, u.Message(false, err.Error()))
//		//			return
//		//		}
//		//		resp = account.ChangePW()
//		//	}
//		default:
//			{
//				//logger.Info.Println("ManageAct: Unregistered action", r.RemoteAddr)
//				w.WriteHeader(http.StatusBadRequest)
//				u.Respond(w, r, u.Message(false, "Unregistered action"))
//				return
//			}
//		}
//	}
//	u.Respond(w, r, resp)
//}
