package whandlers

import (
	"../data"
	u "../utils"
	"net/http"
)

//TestHello
var TestHello = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	account := &data.Account{}
	//account.Password = "3123123"
	//account.Email = "dasd12"
	//a := &data.Point{}
	//a.X = 58.5465465651231
	//a.Y = 36.4564654654321
	// account.TakePoint()
	data.GetDB().Table("accounts").Where("email = ?","super@super").First(&account)
	account.ParserPoints()
	resp := u.Message(true, "Hello")
	resp["account"] = account
	u.Respond(w, resp)
})
