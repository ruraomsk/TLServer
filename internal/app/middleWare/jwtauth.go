package middleWare

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/JanFant/TLServer/logger"

	"github.com/gin-gonic/gin"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/dgrijalva/jwt-go"
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
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(license.LicenseFields.TokenPass), nil
		})

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
		rows, err := data.GetDB().Query(`SELECT token, privilege FROM public.accounts WHERE login = $1`, tk.Login)
		if err != nil {
			c.HTML(http.StatusForbidden, "accessDenied.html", gin.H{"status": http.StatusForbidden, "message": "can't take token from BD"})
			c.Abort()
			return
		}
		for rows.Next() {
			_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
		}

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
		var mapCont = make(map[string]string)
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

		mapCont["login"] = tk.Login
		mapCont["region"] = tk.Region
		mapCont["role"] = tk.Role
		mapCont["perm"] = fmt.Sprint(tk.Permission)

		c.Set("info", mapCont)
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
			u.SendRespond(c, u.Message(http.StatusForbidden, "missing cookie"))
			c.Abort()
			return
		}
		tokenString = cookie

		ip := strings.Split(c.Request.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			u.SendRespond(c, u.Message(http.StatusForbidden, "missing auth token"))
			c.Abort()
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			u.SendRespond(c, u.Message(http.StatusForbidden, "invalid token"))
			c.Abort()
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(license.LicenseFields.TokenPass), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			resp := u.Message(http.StatusForbidden, "wrong auth token")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			c.Abort()
			return
		}

		//Проверка на уникальность токена
		var (
			userPrivilege  data.Privilege
			tokenStrFromBd string
		)
		rows, err := data.GetDB().Query(`SELECT token, privilege FROM public.accounts WHERE login = $1`, tk.Login)
		if err != nil {
			u.SendRespond(c, u.Message(http.StatusForbidden, "can't take token from BD"))
			c.Abort()
			return
		}
		for rows.Next() {
			_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
		}

		if tokenSTR != tokenStrFromBd {
			resp := u.Message(http.StatusForbidden, "token is out of date, log in")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			c.Abort()
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			resp := u.Message(http.StatusForbidden, "Invalid token, log in again")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			c.Abort()
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			resp := u.Message(http.StatusForbidden, "Invalid auth token")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			c.Abort()
			return
		}

		c.Next()
	}
}
