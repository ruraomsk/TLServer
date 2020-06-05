package middleWare

import (
	"github.com/JanFant/TLServer/logger"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/JanFant/TLServer/internal/model/data"
	u "github.com/JanFant/TLServer/internal/utils"
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
			return
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
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "accessDenied"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", c.Request.RemoteAddr, mapContx["login"], c.Request.RequestURI, "accessDenied")
			c.Abort()
			return
		}

	}
}
