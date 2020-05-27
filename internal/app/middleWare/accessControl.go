package middleWare

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"

	"github.com/JanFant/newTLServer/internal/model/data"
	u "github.com/JanFant/newTLServer/internal/utils"
)

//AccessControl проверка разрешен ли пользователя доступ к запрашиваемому ресурсу
var AccessControl = func() gin.HandlerFunc {
	return func(c *gin.Context) {
		//достаем разрешенные группы маршрутов из контекста
		mapContx := u.ParserInterface(c.Value("info"))
		var permission []int
		permStr := mapContx["perm"]
		permStr = strings.TrimPrefix(permStr, "[")
		permStr = strings.TrimSuffix(permStr, "]")
		for _, value := range strings.Split(permStr, " ") {
			intVal, _ := strconv.Atoi(value)
			permission = append(permission, intVal)
		}

		//убираем из url лишнее
		url := c.Request.URL.Path
		url = strings.TrimPrefix(url, "/user/")
		url = url[strings.Index(url, "/"):]

		//если маршрут не найдет отправляем дальше, там разберутся (404)
		rout, ok := data.RoleInfo.MapRoutes[url]

		if !ok {
			c.HTML(http.StatusNotFound, "notFound.html", gin.H{"message": "page not found"})
			c.Abort()
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
			c.Next()
		} else {
			resp := u.Message(http.StatusForbidden, "access denied")
			resp.Obj["logLogin"] = mapContx["login"]
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"message": "accessDenied"})
			u.SendRespond(c, resp)
			c.Abort()
			return
		}

	}
}
