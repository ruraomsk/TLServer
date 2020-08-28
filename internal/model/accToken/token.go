package accToken

import (
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/dgrijalva/jwt-go"
)

//Token (JWT) структура токена доступа
type Token struct {
	Login       string   //Уникальный логин пользователя
	IP          string   //IP пользователя
	Description string   //описание арм
	Role        string   //Роль
	Permission  []int    //Привелегии
	Region      string   //Регион пользователя
	Area        []string //список доступных регионов
	jwt.StandardClaims
}

//Parse разобрать токен
func (t *Token) Parse(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenStr, t, func(token *jwt.Token) (interface{}, error) {
		return []byte(license.LicenseFields.TokenPass), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
