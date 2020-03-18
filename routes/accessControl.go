package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
)

//AccessControl проверка разрешен ли пользователя доступ к запрашиваемому ресурсу
var AccessControl = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//достаем разрешенные группы маршрутов из контекста
		mapContx := u.ParserInterface(r.Context().Value("info"))
		var permission []int
		permStr := mapContx["perm"]
		permStr = strings.TrimPrefix(permStr, "[")
		permStr = strings.TrimSuffix(permStr, "]")
		for _, value := range strings.Split(permStr, " ") {
			intVal, _ := strconv.Atoi(value)
			permission = append(permission, intVal)
		}

		//убираем из url лишнее
		url := r.URL.Path
		url = strings.TrimPrefix(url, "/user/")
		url = url[strings.Index(url, "/"):]

		//если маршрут не найдет отправляем дальше, там разберутся (404)
		rout, ok := data.RoleInfo.MapRoutes[url]
		if !ok {
			http.ServeFile(w, r, data.GlobalConfig.ResourcePath+"/notFound.html")
		}

		access := false
		if mapContx["role"] == "Admin" {
			access = true
		}
		//смотрим если ли доступ у пользователя к этому машруту
		for _, perm := range permission {
			if perm == rout.Permission {
				access = true
				break
			}
		}

		//если все нормально отправляем дальше, или запрещаем доступ
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
