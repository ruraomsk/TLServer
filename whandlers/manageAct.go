package whandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
)

//ActUpdateAccount обработчик запроса обновления (работа с пользователями)
var ActUpdateAccount = func(w http.ResponseWriter, r *http.Request) {
	flag, resp := FuncAccessCheck(w, r, 1)
	if flag {
		mapContx := u.ParserInterface(r.Context().Value("info"))
		var shortAcc = &data.ShortAccount{}
		err := shortAcc.DecodeRequest(w, r)
		if err != nil {
			return
		}
		err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
		if err != nil {
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
	flag, resp := FuncAccessCheck(w, r, 1)
	if flag {
		mapContx := u.ParserInterface(r.Context().Value("info"))
		var shortAcc = &data.ShortAccount{}
		err := shortAcc.DecodeRequest(w, r)
		if err != nil {
			return
		}
		account, err := shortAcc.ValidDelete(mapContx["role"], mapContx["region"])
		if err != nil {
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
	flag, resp := FuncAccessCheck(w, r, 1)
	if flag {
		mapContx := u.ParserInterface(r.Context().Value("info"))
		var shortAcc = &data.ShortAccount{}
		err := shortAcc.DecodeRequest(w, r)
		if err != nil {
			return
		}
		err = shortAcc.ValidCreate(mapContx["role"], mapContx["region"])
		if err != nil {
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
	flag, resp := FuncAccessCheck(w, r, 2)
	if flag {
		mapContx := u.ParserInterface(r.Context().Value("info"))
		var passChange = &data.PassChange{}
		err := json.NewDecoder(r.Body).Decode(passChange)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, fmt.Sprintf("Incorrectly filled data: %s", err.Error())))
			return
		}
		account, err := passChange.ValidOldNewPW(mapContx["login"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, r, u.Message(false, err.Error()))
			return
		}
		resp = account.ChangePW()
	}
	u.Respond(w, r, resp)
}
