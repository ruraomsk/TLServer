package data

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Message struct {
	From    string    `json:"from"`
	To      string    `json:"to"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type ErrorMessage struct {
	User  string `json:"user"`
	Error string `json:"error"`
}

var Connections map[*websocket.Conn]string
var Names UsersInfo

type UsersInfo struct {
	Users map[string]bool `json:"Users"`
}

func (myUser *UsersInfo) add(newUser string) {
	myUser.Users[newUser] = true
}

func (myUser *UsersInfo) delete(delUser string) {
	delete(myUser.Users, delUser)
}

func sendAllUsers(users []int) {
	for connect := range Connections {
		if err := connect.WriteJSON(users); err != nil {
			delete(Connections, connect)
			return
		}
	}
}

func ChatReader(conn *websocket.Conn, mapContx map[string]string) {
	Connections[conn] = mapContx["login"]

	fmt.Println(Connections)

	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			delete(Connections, conn)
			return
		}

		var MessageFrom Message
		MessageFrom.From = Connections[conn]
		MessageFrom.To = "Global"
		MessageFrom.Message = string(p)
		MessageFrom.Time = time.Now()

		// print out that message for clarity
		fmt.Println(string(p))
		fmt.Println(len(Connections))

		switch {
		case MessageFrom.To == "Global":
			{
				if err := saveMessage(MessageFrom); err != nil {
					var mess = ErrorMessage{User: MessageFrom.From, Error: "Сообщение не доставленно попробуйте еще раз"}
					_ = conn.WriteJSON(mess)
				}
				for connect := range Connections {
					if err := connect.WriteJSON(MessageFrom); err != nil {
						delete(Connections, connect)
						return
					}
				}
			}
		case MessageFrom.To != "Global":
			{
				for connect, state := range Connections {
					if MessageFrom.To == state || MessageFrom.From == mapContx["login"] {
						if err := connect.WriteJSON(MessageFrom); err != nil {
							delete(Connections, connect)
							return
						}
					}
				}
			}
		}
	}
}

func saveMessage(mess Message) error {
	_, err := db.DB().Query(`INSERT INTO public.chat (time, fromu, tou, message) VALUES ($1, $2, $3, $4)`, mess.Time, mess.From, mess.To, mess.Message)
	if err != nil {
		return err
	}
	return nil
}
