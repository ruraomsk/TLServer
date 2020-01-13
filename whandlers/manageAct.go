package whandlers

import (
	"../data"
	"fmt"
	"net/http"
)

var actParser = func(w http.ResponseWriter, r *http.Request) {
	mapContx := data.ParserInterface(r.Context().Value("info"))
	//var actions = []string{
	//	"update",
	//	"delete",
	//	"add",
	//}

	switch mapContx["act"] {
	case "update":
		fmt.Println("update")
	case "delete":
		fmt.Println("delete")
	case "add":
		fmt.Println("add")
	default:
		fmt.Println("bad requst")
	}


	//if mapContx["act"] == actions[0] { //проверка update
	//	flag, resp := FuncAccessCheak(w, r, "BuildMapPage")
	//	if flag {
	//
	//	}
	//} else if mapContx["act"] == actions[1] { //проверка delete
	//	flag, resp := FuncAccessCheak(w, r, "BuildMapPage")
	//	if flag {
	//
	//	}
	//} else if mapContx["act"] == actions[2] { //проверка add
	//	flag, resp := FuncAccessCheak(w, r, "BuildMapPage")
	//	if flag {
	//
	//	}
	//} else {
	//
	//}

}
