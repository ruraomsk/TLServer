package handlers

import (
	"github.com/JanFant/newTLServer/internal/model/chat"
	"github.com/JanFant/newTLServer/internal/model/data"
	u "github.com/JanFant/newTLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ChatEngine = func(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
		return
	}
	defer conn.Close()
	mapContx := u.ParserInterface(c.Value("info"))
	chat.ChatReader(conn, mapContx, data.GetDB())
}
