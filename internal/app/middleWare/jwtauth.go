package middleWare

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/gin-gonic/gin"

	"github.com/JanFant/TLServer/internal/model/data"
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
			u.SendRespond(c, u.Message(http.StatusForbidden, "missing cookie"))
			return
		}
		tokenString = cookie

		ip := strings.Split(c.Request.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			u.SendRespond(c, u.Message(http.StatusForbidden, "missing auth token"))
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			u.SendRespond(c, u.Message(http.StatusForbidden, "invalid token"))
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GlobalConfig.TokenPassword), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			resp := u.Message(http.StatusForbidden, "wrong auth token")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
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
			return
		}
		for rows.Next() {
			_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
		}

		if tokenSTR != tokenStrFromBd {
			resp := u.Message(http.StatusForbidden, "token is out of date, log in")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			resp := u.Message(http.StatusForbidden, "Invalid token, log in again")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			resp := u.Message(http.StatusForbidden, "Invalid auth token")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			return
		}

		//проверка токен пришел от правильного URL
		var mapCont = make(map[string]string)
		slug := c.Param("slug")
		if slug != tk.Login {
			resp := u.Message(http.StatusForbidden, fmt.Sprintf("token isn't registered for user: %s", slug))
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			return
		}

		//проверка правильности роли для указанного пользователя
		_ = userPrivilege.ConvertToJson()
		if userPrivilege.Role.Name != tk.Role {
			resp := u.Message(http.StatusForbidden, "Access denied")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
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
			return
		}
		tokenString = cookie

		ip := strings.Split(c.Request.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			u.SendRespond(c, u.Message(http.StatusForbidden, "missing auth token"))
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			u.SendRespond(c, u.Message(http.StatusForbidden, "invalid token"))
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GlobalConfig.TokenPassword), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			resp := u.Message(http.StatusForbidden, "wrong auth token")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
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
			return
		}
		for rows.Next() {
			_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
		}

		if tokenSTR != tokenStrFromBd {
			resp := u.Message(http.StatusForbidden, "token is out of date, log in")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			resp := u.Message(http.StatusForbidden, "Invalid token, log in again")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			resp := u.Message(http.StatusForbidden, "Invalid auth token")
			resp.Obj["logLogin"] = tk.Login
			u.SendRespond(c, resp)
			return
		}

		c.Next()
	}
}
