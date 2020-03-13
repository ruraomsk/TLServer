package routes

import (
	"fmt"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"net/http"
	"strconv"
	"strings"
)

var AccessControl = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//разбираю конетест
		mapContx := u.ParserInterface(r.Context().Value("info"))
		var permission []int
		permStr := mapContx["perm"]
		permStr = strings.TrimPrefix(permStr, "[")
		permStr = strings.TrimSuffix(permStr, "]")
		for _, value := range strings.Split(permStr, " ") {
			intVal, _ := strconv.Atoi(value)
			permission = append(permission, intVal)
		}
		//контроль к ресурсу
		url := r.URL.Path
		url = strings.TrimPrefix(url, "/user/")
		url = url[strings.Index(url, "/"):]

		a := data.HasPath(url)
		fmt.Println(a)
		fmt.Println("---------------------")
		fmt.Println(permission)
		fmt.Println(r.URL.Path)
		fmt.Println(url)
		fmt.Println("---------------------")

		next.ServeHTTP(w, r)
	})
}
