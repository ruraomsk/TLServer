package middleWare

import (
	"fmt"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"net/http"
	"strings"

	"github.com/ruraomsk/TLServer/logger"

	"github.com/gin-gonic/gin"

	"github.com/ruraomsk/TLServer/internal/model/data"
)

//JwtAuth контроль токена для всех прошедших регистрацию и обрашающихся к ресурсу
var JwtAuth = func() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		cookie, err := c.Cookie("Authorization")
		//Проверка куков получили ли их вообще
		if err != nil {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "missing cookie"})
			c.Abort()
			return
		}
		tokenString = cookie

		ip := strings.Split(c.Request.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "missing auth token"})
			c.Abort()
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "invalid token"})
			c.Abort()
			return
		}

		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := new(accToken.Token)

		token, err := tk.Parse(tokenSTR)
		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "wrong auth token"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "wrong auth token")
			c.Abort()
			return
		}

		//Проверка на уникальность токена
		var (
			userPrivilege  data.Privilege
			tokenStrFromBd string
		)
		db, id := data.GetDB()
		rows, err := db.Query(`SELECT token, privilege FROM public.accounts WHERE login = $1`, tk.Login)
		if err != nil {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "can't take token from BD"})
			c.Abort()
			return
		}
		for rows.Next() {
			_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
		}
		data.FreeDB(id)
		if tokenSTR != tokenStrFromBd {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "token is out of date, log in"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "token is out of date, log in")
			c.Abort()
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "Invalid token, log in again"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "Invalid token, log in again")
			c.Abort()
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "Invalid auth token"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "Invalid token, log in again")
			c.Abort()
			return
		}

		//проверка токен пришел от правильного URL
		slug := c.Param("slug")
		if slug != tk.Login {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "token isn't registered for user"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, fmt.Sprintf("token isn't registered for user: %s", slug))
			c.Abort()
			return
		}

		//проверка правильности роли для указанного пользователя
		_ = userPrivilege.ConvertToJson()
		if userPrivilege.Role.Name != tk.Role {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "Access denied"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "Access denied")
			c.Abort()
			return
		}

		c.Set("tk", tk)

		c.Next()
	}
}

//JwtFile контроль токена для всех прошедших регистрацию и обрашающихся к ресурсу
var JwtFile = func() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		cookie, err := c.Cookie("Authorization")
		//Проверка куков получили ли их вообще
		if err != nil {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "missing cookie"})
			c.Abort()
			return
		}
		tokenString = cookie

		ip := strings.Split(c.Request.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "missing auth token"})
			c.Abort()
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "invalid token"})
			c.Abort()
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := new(accToken.Token)

		token, err := tk.Parse(tokenSTR)

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "wrong auth token"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "wrong auth token")
			c.Abort()
			return
		}

		//Проверка на уникальность токена
		var (
			userPrivilege  data.Privilege
			tokenStrFromBd string
		)
		db, id := data.GetDB()

		rows, err := db.Query(`SELECT token, privilege FROM public.accounts WHERE login = $1`, tk.Login)
		if err != nil {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "can't take token from BD"})
			c.Abort()
			return
		}
		for rows.Next() {
			_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
		}
		data.FreeDB(id)

		if tokenSTR != tokenStrFromBd {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "token is out of date, log in"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "token is out of date, log in")
			c.Abort()
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "Invalid token, log in again"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "Invalid token, log in again")
			c.Abort()
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "Invalid auth token"})
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, tk.Login, c.Request.RequestURI, "Invalid token, log in again")
			c.Abort()
			return
		}

		c.Next()
	}
}
