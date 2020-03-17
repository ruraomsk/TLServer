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
		//разбираю context
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

		access := false
		for _, perm := range permission {
			if perm == rout.Permission {
				access = true
				break
			}
		}

		if access {
			next.ServeHTTP(w, r)
		} else {
			resp := u.Message(false, "Access denied")
			resp["logLogin"] = mapContx["login"]
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

	})
}
