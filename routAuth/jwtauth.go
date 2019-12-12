package routAuth

import (
	"context"
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

		tokenHeader := r.Header.Get("Authorization")
		ip := strings.Split(r.RemoteAddr, ":")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenHeader == "" {
			response := u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			response := u.Message(false, "Invalid token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}
		//берем часть где хранится токен
		tokenPath := splitted[1]
		tk := &data.Token{}

		token, err := jwt.ParseWithClaims(tokenPath, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		//не правильный токен возвращаем ошибку с кодом 403
		if err != nil {
			response := u.Message(false, "Wrong auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//проверка с какого ip пришел токен
		if tk.IP != ip[0] {
			response := u.Message(false, "invalid token, log in again")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			response := u.Message(false, "Invalid auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, r, response)
			return
		}

		//проверка токен пришел от правльного URL
		vars := mux.Vars(r)
		slug := vars["slug"]
		if slug != tk.Login {
			u.Respond(w, r, u.Message(false, "token isn't registered for this user"))
			return
		}

		ctx := context.WithValue(r.Context(), "user", tk.Login)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}
