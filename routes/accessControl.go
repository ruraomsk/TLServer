package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
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

		rout, ok := data.RoleInfo.MapRoutes[url]
		if !ok {
			next.ServeHTTP(w, r)
		}

		var permId int
		for id, perm := range data.RoleInfo.MapPermisson {
			for _, command := range perm.Commands {
				if command == rout.ID {
					permId = id
				}
			}
		}

		access := false
		for _, perm := range permission {
			if perm == permId {
				access = true
			}
		}

		if access {
			flag, err := data.NewRoleCheck(u.ParserInterface(r.Context().Value("info")), permId)
			if err != nil || !flag {
				resp = u.Message(false, err.Error())
				if err != nil {
					resp = u.Message(false, "Access denied")
				}
				w.WriteHeader(http.StatusForbidden)
			}
		}else{

		}
		next.ServeHTTP(w, r)
	})
}
