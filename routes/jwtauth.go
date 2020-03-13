package routes

import (
	"context"
	"fmt"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

//JwtAuth контроль токена для всех прошедших регистрацию и обрашающихся к ресурсу
var JwtAuth = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		cookie, err := r.Cookie("Authorization")
		//Проверка куков получили ли их вообще
		if err != nil {
			resp := u.Message(false, "Missing cookie")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}
		tokenString = cookie.Value

		ip := strings.Split(r.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			resp := u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			resp := u.Message(false, "Invalid token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(data.GlobalConfig.TokenPassword), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			resp := u.Message(false, "Wrong auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		//Проверка на уникальность токена
		var tokenStrFromBd string
		err = data.GetDB().Table("accounts").Select("token").Where("login = ?", tk.Login).Row().Scan(&tokenStrFromBd)
		if err != nil {
			resp := u.Message(false, fmt.Sprintf("Can't take token from BD: %s", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		if tokenSTR != tokenStrFromBd {
			resp := u.Message(false, "Token is out of date, log in")
			resp["logLogin"] = tk.Login
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			resp := u.Message(false, "Invalid token, log in again")
			resp["logLogin"] = tk.Login
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			resp := u.Message(false, "Invalid auth token")
			resp["logLogin"] = tk.Login
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		//проверка токен пришел от правильного URL
		vars := mux.Vars(r)
		var mapCont = make(map[string]string)
		slug := vars["slug"]
		if strings.Contains(r.RequestURI, "/manage/") {
			mapCont["act"] = vars["act"]
		}
		if slug != tk.Login {
			resp := u.Message(false, fmt.Sprintf("Token isn't registered for user: %s", slug))
			resp["logLogin"] = tk.Login
			u.Respond(w, r, resp)
			return
		}
		mapCont["login"] = tk.Login
		mapCont["region"] = tk.Region
		mapCont["role"] = tk.Role
		mapCont["perm"] = fmt.Sprint(tk.Permission)

		ctx := context.WithValue(r.Context(), "info", mapCont)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)

	})
}

//JwtFile упрошенных контроль токена для получения данных из файлового хранилища
var JwtFile = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		cookie, err := r.Cookie("Authorization")
		//Проверка куков получили ли их вообще
		if err != nil {
			resp := u.Message(false, "Missing cookie")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}
		tokenString = cookie.Value

		ip := strings.Split(r.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			resp := u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			resp := u.Message(false, "Invalid token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(data.GlobalConfig.TokenPassword), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			resp := u.Message(false, "Wrong auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		//Проверка на уникальность токена
		var tokenStrFromBd string
		err = data.GetDB().Table("accounts").Select("token").Where("login = ?", tk.Login).Row().Scan(&tokenStrFromBd)
		if err != nil {
			resp := u.Message(false, fmt.Sprintf("Can't take token from BD: %v", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		if tokenSTR != tokenStrFromBd {
			resp := u.Message(false, "Token is out of date, log in")
			resp["logLogin"] = tk.Login
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			resp := u.Message(false, "Invalid token, log in again")
			resp["logLogin"] = tk.Login
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			resp := u.Message(false, "Invalid auth token")
			resp["logLogin"] = tk.Login
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, resp)
			return
		}
		next.ServeHTTP(w, r)
	})
}
