package whandlers

import (
	"fmt"
	u "github.com/JanFant/TLServer/utils"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var Connections map[*websocket.Conn]Message
var count = 0

var Chat = func(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		delete(Connections, conn)
		fmt.Println("start")
		w.WriteHeader(http.StatusInternalServerError)
		u.Respond(w, r, u.Message(false, "Bad socket connect"))
		return
	}
	defer conn.Close()
	count++
	Connections[conn] = Message{User: fmt.Sprint(count)}

	reader(conn)
}

type Message struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

func reader(conn *websocket.Conn) {
	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			delete(Connections, conn)
			fmt.Println("1", err)
			return
		}

		// print out that message for clarity
		fmt.Println(string(p), "   ")
		fmt.Println(len(Connections))

		var MessageFrom Message
		MessageFrom.User = Connections[conn].User
		MessageFrom.Message = string(p)
		for connect := range Connections {
			if err := connect.WriteJSON(MessageFrom); err != nil {
				delete(Connections, conn)
				fmt.Println("2", err)
				return
			}
		}

	}
}
