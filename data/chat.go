package data

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Message struct {
	Type    string    `json:"type"`
	From    string    `json:"from"`
	To      string    `json:"to"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

type ErrorMessage struct {
	Type  string `json:"type"`
	User  string `json:"user"`
	Error string `json:"error"`
}

type UsersInfo struct {
	Type  string          `json:"type"`
	Users map[string]bool `json:"users"`
}

type StatusUser struct {
	Type   string `json:"type"`
	User   string `json:"user"`
	Status string `json:"status"`
}

type AllUsers struct {
	Type  string   `json:"type"`
	Users []string `json:"users"`
}

var Connections map[*websocket.Conn]string
var Names UsersInfo

var (
	messageInfo   = "message"
	errorMessage  = "error"
	statusInfo    = "status"
	statusOnline  = "online"
	statusOffline = "offline"
	allUsers      = "users"

	errNoAccessWithDatabase = "no access with database"
)

func (message *Message) toStruct(str []byte) (err error) {
	err = json.Unmarshal(str, message)
	if err != nil {
		return err
	}
	return nil
}

func (status *StatusUser) send() {
	for conn := range Connections {
		_ = conn.WriteJSON(status)
	}
}

func (myUser *UsersInfo) add(newUser string) {
	myUser.Users[newUser] = true
}

func (myUser *UsersInfo) del(outUser string) {
	delete(myUser.Users, outUser)
}

func (myUser *UsersInfo) send() {
	for conn := range Connections {
		_ = conn.WriteJSON(myUser)
	}
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

	//все пользователи
	users, err := getAllUsers()
	if err != nil {
		var mess = ErrorMessage{User: mapContx["login"], Error: errNoAccessWithDatabase, Type: errorMessage}
		_ = conn.WriteJSON(mess)
	}
	_ = conn.WriteJSON(users)

	//пользователи онлайн
	//Names.add(mapContx["login"])
	//Names.Type = online
	//Names.send()

	var userStatus = StatusUser{Status: statusOnline, Type: statusInfo, User: mapContx["login"]}
	userStatus.send()

	fmt.Println(Connections)

	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			delete(Connections, conn)
			//Names.del(mapContx["login"])
			//Names.send()
			userStatus.Status = statusOffline
			userStatus.send()
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
				if err := saveMessage(messageFrom); err != nil {
					var mess = ErrorMessage{User: messageFrom.From, Error: "Сообщение не доставленно попробуйте еще раз", Type: errorMessage}
					err = conn.WriteJSON(mess)
				}
				for connect := range Connections {
					if err := connect.WriteJSON(messageFrom); err != nil {
						delete(Connections, connect)
						userStatus.Status = statusOffline
						userStatus.send()
						//Names.del(mapContx["login"])
						//Names.send()
						fmt.Println(Connections)
						return
					}
				}
			}
		case messageFrom.To != "Global":
			{
				if err := saveMessage(messageFrom); err != nil {
					var mess = ErrorMessage{User: messageFrom.From, Error: "Сообщение не доставленно попробуйте еще раз", Type: errorMessage}
					err = conn.WriteJSON(mess)
				}
				for connect, state := range Connections {
					if messageFrom.To == state || messageFrom.From == mapContx["login"] {
						if err := connect.WriteJSON(messageFrom); err != nil {
							delete(Connections, connect)
							userStatus.Status = statusOffline
							userStatus.send()
							return
						}
					}
				}
			}
		}
	}
}

func saveMessage(mess Message) error {
	_, err := db.DB().Exec(`INSERT INTO public.chat (time, fromu, tou, message) VALUES ($1, $2, $3, $4)`, mess.Time, mess.From, mess.To, mess.Message)

	if err != nil {
		return err
	}
	return nil
}

func getAllUsers() (AllUsers, error) {
	var (
		tempUser string
		users    AllUsers
	)
	rows, err := db.DB().Query(`SELECT login FROM public.accounts`)
	if err != nil {
		return users, err
	}
	for rows.Next() {
		_ = rows.Scan(&tempUser)
		users.Users = append(users.Users, tempUser)
	}
	users.Type = allUsers
	return users, nil
}
