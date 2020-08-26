package mainMap

import (
	"github.com/JanFant/TLServer/internal/model/accToken"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"net/http"
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

	accInfo := new(accToken.Token)
	tokenInfo := new(jwt.Token)

	cookie, err := c.Cookie("Authorization")
	//Проверка куков получили ли их вообще
	if err != nil {
		cookie = ""
	}
	accInfo.IP = c.ClientIP()
	client := &ClientMainMap{hub: hub, conn: conn, send: make(chan mapResponse, 256), cInfo: accInfo, rawToken: tokenInfo.Raw, cookie: cookie}
	client.hub.register <- client

	go client.writePump()
	go client.readPump(db)
}
