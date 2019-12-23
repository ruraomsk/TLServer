package routes

import (
	"../logger"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strings"

	"../data"
	u "../utils"
	"github.com/dgrijalva/jwt-go"
)

var JwtAuth = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var tokenString string
		cookie, err := r.Cookie("Authorization")
		//Проверка куков получили ли их вообще
		if err != nil {
			logger.Info.Println("jwtauth: Missing cookie ", r.RemoteAddr)
			response := u.Message(false, "Missing cookie")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		tokenString = cookie.Value

		ip := strings.Split(r.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			logger.Info.Println("jwtauth: Missing auth token ", r.RemoteAddr)
			response := u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			logger.Info.Println("jwtauth: Invalid token ", r.RemoteAddr)
			response := u.Message(false, "Invalid token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			logger.Info.Println("jwtauth: Wrong auth token ", r.RemoteAddr)
			response := u.Message(false, "Wrong auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//Проверка на уникальность токена
		var tokenStrFromBd string
		err = data.GetDB().Table("accounts").Select("token").Where("login = ?", tk.Login).Row().Scan(&tokenStrFromBd)
		if err != nil {
			logger.Info.Println("jwtauth: Can't take token from BD ", r.RemoteAddr)
			response := u.Message(false, fmt.Sprintf("Can't take token from BD: %v", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		if tokenSTR != tokenStrFromBd {
			logger.Info.Println("jwtauth: Token is out of date, log in ", r.RemoteAddr)
			response := u.Message(false, "Token is out of date, log in")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			logger.Info.Println("jwtauth: invalid token, log in again ", r.RemoteAddr)
			response := u.Message(false, "invalid token, log in again")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			logger.Info.Println("jwtauth: Invalid auth token ", r.RemoteAddr)
			response := u.Message(false, "Invalid auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//проверка токен пришел от правльного URL
		vars := mux.Vars(r)
		slug := vars["slug"]
		if slug != tk.Login {
			logger.Info.Println("jwtauth: token isn't registered for this user ", r.RemoteAddr, "  ", tk.Login)
			u.Respond(w, r, u.Message(false, "token isn't registered for this user"))
			return
		}

		ctx := context.WithValue(r.Context(), "user", tk.Login)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}

var JwtFile = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		cookie, err := r.Cookie("Authorization")
		//Проверка куков получили ли их вообще
		if err != nil {
			logger.Info.Println("jwtauthFileserv: Missing cookie ", r.RemoteAddr)
			response := u.Message(false, "Missing cookie")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		tokenString = cookie.Value

		ip := strings.Split(r.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenString == "" {
			logger.Info.Println("jwtauthFileserv: Missing auth token ", r.RemoteAddr)
			response := u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenString, " ")
		if len(splitted) != 2 {
			logger.Info.Println("jwtauthFileserv: Invalid token ", r.RemoteAddr)
			response := u.Message(false, "Invalid token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		//берем часть где хранится токен
		tokenSTR := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			logger.Info.Println("jwtauthFileserv: Wrong auth token ", r.RemoteAddr)
			response := u.Message(false, "Wrong auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//Проверка на уникальность токена
		var tokenStrFromBd string
		err = data.GetDB().Table("accounts").Select("token").Where("login = ?", tk.Login).Row().Scan(&tokenStrFromBd)
		if err != nil {
			logger.Info.Println("jwtauthFileserv: Can't take token from BD ", r.RemoteAddr)
			response := u.Message(false, fmt.Sprintf("Can't take token from BD: %v", err.Error()))
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}


		if tokenSTR != tokenStrFromBd {
			logger.Info.Println("jwtauthFileserv: Token is out of date, log in ", r.RemoteAddr)
			response := u.Message(false, "Token is out of date, log in")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			logger.Info.Println("jwtauthFileserv: Invalid token, log in again ", r.RemoteAddr)
			response := u.Message(false, "Invalid token, log in again")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}




		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			logger.Info.Println("jwtauthFileserv: Invalid auth token ", r.RemoteAddr)
			response := u.Message(false, "Invalid auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		next.ServeHTTP(w, r)

	})
}
