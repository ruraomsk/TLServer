package chat

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
)

var Connections map[*websocket.Conn]string
var WriteAll chan []byte
var WriteTo chan Message

func Broadcast() {
	Connections = make(map[*websocket.Conn]string)
	WriteAll = make(chan []byte)
	WriteTo = make(chan Message)
	for {
		select {
		case msg := <-WriteAll:
			{
				for connect := range Connections {
					if err := connect.WriteMessage(websocket.TextMessage, msg); err != nil {
						connect.Close()
						delete(Connections, connect)
						fmt.Println(Connections)
						return
					}
				}
			}
		case msg := <-WriteTo:
			{
				for connect, state := range Connections {
					if msg.To == state || msg.From == state {
						if err := connect.WriteJSON(msg); err != nil {
							connect.Close()
							delete(Connections, connect)
							return
						}
					}
				}
			}
		}
	}
}

func Reader(conn *websocket.Conn, mapContx map[string]string, db *sqlx.DB) {
	Connections[conn] = mapContx["login"]

	var users AllUsersStatus
	////все пользователи
	err := users.getAllUsers(db)
	if err != nil {
		_ = conn.WriteJSON(newErrorMessage(mapContx["login"], errNoAccessWithDatabase))
	}
	_ = conn.WriteJSON(users)

	uStatus := newStatus(mapContx["login"])
	uStatus.send(statusOnline)

	fmt.Println(Connections)

	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			delete(Connections, conn)
			uStatus.send(statusOffline)
			fmt.Println(Connections)
			return
		}
		var messageFrom Message
		err = messageFrom.toStruct(p)
		if err != nil {
			fmt.Println(err.Error())
		}
		// print out that message for clarity
		switch {
		case messageFrom.To == "Global":
			{
				if err := saveMessage(messageFrom, db); err != nil {
					var mess = ErrorMessage{User: messageFrom.From, Error: "Сообщение не доставленно попробуйте еще раз", Type: errorMessage}
					err = conn.WriteJSON(mess)
				}
				raw, _ := messageFrom.toString()

				WriteAll <- raw
			}
		case messageFrom.To != "Global":
			{
				if err := saveMessage(messageFrom, db); err != nil {
					var mess = ErrorMessage{User: messageFrom.From, Error: "Сообщение не доставленно попробуйте еще раз", Type: errorMessage}
					err = conn.WriteJSON(mess)
				}
				WriteTo <- messageFrom
			}
		}
	}
}
