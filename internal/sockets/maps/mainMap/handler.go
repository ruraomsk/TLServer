package mainMap

import (
	"github.com/JanFant/TLServer/internal/model/accToken"
	"github.com/JanFant/TLServer/internal/model/license"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

//HMainMap обработчик открытия сокета
func HMainMap(c *gin.Context, hub *HubMainMap, db *sqlx.DB) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "bad socket connect"))
		return
	}

	tokenStr, _ := c.Cookie("Authorization")
	ip := strings.Split(c.Request.RemoteAddr, ":")

	var cInfo = clientInfo{login: "", ip: ip, tokenStr: tokenStr, token: getToken(tokenStr)}
	client := &ClientMainMap{hub: hub, conn: conn, send: make(chan mapResponse, 256), cInfo: &cInfo}
	client.hub.register <- client

	go client.writePump()
	go client.readPump(db, c)
}

func getToken(tokenStr string) (token *jwt.Token) {
	token = new(jwt.Token)
	if tokenStr != "" {
		splitted := strings.Split(tokenStr, " ")
		if len(splitted) == 2 {
			//берем часть где хранится токен
			tokenSTR := splitted[1]
			tk := &accToken.Token{}
			token, _ = jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
				return []byte(license.LicenseFields.TokenPass), nil
			})
		}
	}
	return token
}
