package chatH

import (
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

////ChatEngine обработчик вебсокета для работы с чатом
//var ChatEngine = func(c *gin.Context) {
//	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
//	if err != nil {
//		u.SendRespond(c, u.Message(http.StatusInternalServerError, "Bad socket connect"))
//		return
//	}
//	defer conn.Close()
//	mapContx := u.ParserInterface(c.Value("info"))
//	chat.Reader(conn, mapContx["login"], data.GetDB())
//}
