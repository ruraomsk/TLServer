package whandlers

import (
	"fmt"
	"github.com/JanFant/TLServer/data"
	u "github.com/JanFant/TLServer/utils"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var Chat = func(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		delete(data.Connections, conn)
		fmt.Println("start")
		w.WriteHeader(http.StatusInternalServerError)
		u.Respond(w, r, u.Message(false, "Bad socket connect"))
		return
	}
	defer conn.Close()
	mapContx := u.ParserInterface(r.Context().Value("info"))
	data.ChatReader(conn, mapContx)
}
