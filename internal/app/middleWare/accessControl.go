package middleWare

import (
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ruraomsk/TLServer/internal/model/data"
)

//AccessControl проверка разрешен ли пользователя доступ к запрашиваемому ресурсу
var AccessControl = func() gin.HandlerFunc {
	return func(c *gin.Context) {
		//достаем разрешенные группы маршрутов из контекста
		accTK, _ := c.Get("tk")
		accInfo, _ := accTK.(*accToken.Token)

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
		if accInfo.Role == "Admin" {
			access = true
		}
		//смотрим если ли доступ у пользователя к этому машруту
		for _, perm := range accInfo.Permission {
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
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", c.Request.RemoteAddr, accInfo.Login, c.Request.RequestURI, "accessDenied")
			c.Abort()
			return
		}

	}
}
