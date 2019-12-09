package routAuth

import (
	"../data"
	u "../utils"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"strings"
)

var JwtAuth = func(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//заполняем срез путями которые не нужно проверять токеном
		//нужно бы перейти на саброутер чтобы не писать это!
		notAuth := []string{
			"/login",
			"/test",
			"/static/",
			"/hello",
			"/create",
		}
		requestPath := r.URL.Path
		//если введенный путь совпал вываливаемся на исполнение
		for _, val := range notAuth {
			if val == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		tokenHeader := r.Header.Get("Authorization")
		//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
		if tokenHeader == "" {
			response := u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, response)
			return
		}

		//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			response := u.Message(false, "Invalid token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, response)
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
			u.Respond(w, response)
			return
		}

		//токен не действителен, возможно не подписан на этом сервере
		if !token.Valid {
			response := u.Message(false, "Invalid auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, response)
			return
		}

		fmt.Println("User ", tk.UserID)

		ctx := context.WithValue(r.Context(), "user", tk.UserID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}
