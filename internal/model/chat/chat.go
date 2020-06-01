package chat

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
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

func Reader(conn *websocket.Conn, login string, db *sqlx.DB) {
	Connections[conn] = login

	var users AllUsersStatus
	////все пользователи
	err := users.getAllUsers(db)
	if err != nil {
		_ = conn.WriteJSON(newErrorMessage(errNoAccessWithDatabase))
	}
	_ = conn.WriteJSON(users)

	uStatus := newStatus(login)
	uStatus.send(statusOnline)

	{
		var messagesss PeriodMessage
		messagesss.TimeStart = time.Now()
		messagesss.TimeEnd = messagesss.TimeStart.AddDate(0, 0, -1)
		_ = messagesss.takeMessages(db)
		_ = conn.WriteJSON(messagesss)
	}
	fmt.Println(Connections)

	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			delete(Connections, conn)
			if !checkAnother(login) {
				uStatus.send(statusOffline)
			}
			fmt.Println(Connections)
			return
		}

		typeMess, err := setTypeMessage(p)
		if err != nil {
			var mess = ErrorMessage{Error: "Не верный тип сообщения", Type: errorMessage}
			err = conn.WriteJSON(mess)
		}

		switch typeMess {
		case messageInfo:
			{
				var messageFrom Message
				err = messageFrom.toStruct(p)
				if err != nil {
					fmt.Println(err.Error())
				}

				switch {
				case messageFrom.To == "Global":
					{
						if err := saveMessage(messageFrom, db); err != nil {
							var mess = ErrorMessage{Error: "Сообщение не доставленно попробуйте еще раз", Type: errorMessage}
							err = conn.WriteJSON(mess)
						}
						raw, _ := messageFrom.toString()

						WriteAll <- raw
					}
				case messageFrom.To != "Global":
					{
						if err := saveMessage(messageFrom, db); err != nil {
							var mess = ErrorMessage{Error: "Сообщение не доставленно попробуйте еще раз", Type: errorMessage}
							err = conn.WriteJSON(mess)
						}
						WriteTo <- messageFrom
					}
				}
			}
		case messages:
			{
				var messagesss PeriodMessage
				messagesss.TimeStart = time.Now()
				messagesss.TimeEnd = messagesss.TimeStart.AddDate(0, 0, -1)
				_ = messagesss.takeMessages(db)
				_ = conn.WriteJSON(messagesss)
				continue
			}
		}

	}
}
