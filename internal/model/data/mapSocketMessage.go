package data

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

//MapSokResponse структура для отправки сообщений (map)
type MapSokResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	conn *websocket.Conn        `json:"-"`
}

//newMapMess создание нового сообщения
func newMapMess(mType string, conn *websocket.Conn, data map[string]interface{}) MapSokResponse {
	var resp MapSokResponse
	resp.Type = mType
	resp.conn = conn
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//send отправка сообщения с обработкой ошибки
func (m *MapSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", m.conn.RemoteAddr(), "map socket", "/map", m.Data["message"])
		}()
	}
	writeMap <- *m
}

//setTypeMessage определение типа сообщения
func setTypeMessage(raw []byte) (string, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(raw, &temp); err != nil {
		return "", err
	}
	return fmt.Sprint(temp["type"]), nil
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

var (
	typeJump           = "jump"
	typeMapInfo        = "mapInfo"
	typeTFlight        = "tflight"
	typeRepaint        = "repaint"
	typeEditCrossUsers = "editCrossUsers"
	typeLogin          = "login"
	typeLogOut         = "logOut"
	//errNoAccessWithDatabase    = "no access with database"
	//errCantConvertJSON         = "cant convert JSON"
	errUnregisteredMessageType = "unregistered message type"
)

//checkToken проверка токена для вебсокета
func checkToken(c *gin.Context) (flag bool, t *Token) {
	var tokenString string
	cookie, err := c.Cookie("Authorization")
	//Проверка куков получили ли их вообще
	if err != nil {
		return false, nil
	}
	tokenString = cookie

	ip := strings.Split(c.Request.RemoteAddr, ":")
	//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
	if tokenString == "" {
		return false, nil
	}
	//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
	splitted := strings.Split(tokenString, " ")
	if len(splitted) != 2 {
		return false, nil
	}

	//берем часть где хранится токен
	tokenSTR := splitted[1]
	tk := &Token{}

	token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(license.LicenseFields.TokenPass), nil
	})

	//не правильный токен возвращаем ошибку с кодом 403
	if err != nil {
		return false, nil
	}

	//Проверка на уникальность токена
	var (
		userPrivilege  Privilege
		tokenStrFromBd string
	)
	rows, err := GetDB().Query(`SELECT token, privilege FROM public.accounts WHERE login = $1`, tk.Login)
	if err != nil {
		return false, nil
	}
	for rows.Next() {
		_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
	}

	if tokenSTR != tokenStrFromBd || tk.IP != ip[0] || !token.Valid {
		return false, nil
	}

	//проверка токен пришел от правильного URL

	//проверка правильности роли для указанного пользователя
	_ = userPrivilege.ConvertToJson()
	if userPrivilege.Role.Name != tk.Role {
		return false, nil
	}

	return true, tk
}
