package chat

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

var ConnectedUsers map[string][]*websocket.Conn
var WriteSendMessage chan SendMessage

func delConn(login string, conn *websocket.Conn) {
	for index, userConn := range ConnectedUsers[login] {
		if userConn == conn {
			ConnectedUsers[login][index] = ConnectedUsers[login][len(ConnectedUsers[login])-1]
			ConnectedUsers[login][len(ConnectedUsers[login])-1] = nil
			ConnectedUsers[login] = ConnectedUsers[login][:len(ConnectedUsers[login])-1]
			break
		}
	}
}

func Broadcast() {
	ConnectedUsers = make(map[string][]*websocket.Conn)
	WriteSendMessage = make(chan SendMessage)
	for {
		select {
		case msg := <-WriteSendMessage:
			{
				switch {
				case msg.from == msg.to:
					{
						if err := msg.conn.WriteJSON(msg); err != nil {
							msg.conn.Close()
							delConn(msg.from, msg.conn)
							break
						}
					}
				case msg.to == globalMessage:
					{
						for _, userConn := range ConnectedUsers {
							for _, conn := range userConn {
								if err := conn.WriteJSON(msg); err != nil {
									conn.Close()
									delConn(msg.from, conn)
									break
								}
							}
						}
					}
				case msg.from != msg.to:
					{
						for _, conn := range ConnectedUsers[msg.from] {
							if err := conn.WriteJSON(msg); err != nil {
								conn.Close()
								delConn(msg.from, conn)
								break
							}
						}
						for _, conn := range ConnectedUsers[msg.to] {
							if err := conn.WriteJSON(msg); err != nil {
								conn.Close()
								delConn(msg.from, conn)
								break
							}
						}
					}
				default:
					continue
				}
			}
		}
	}
}

func Reader(conn *websocket.Conn, login string, db *sqlx.DB) {
	ConnectedUsers[login] = append(ConnectedUsers[login], conn)
	var message SendMessage
	message.conn = conn
	//выгрузить список доступных пользователей
	{
		var users AllUsersStatus
		err := users.getAllUsers(db)
		if err != nil {
			var myError = ErrorMessage{Error: errNoAccessWithDatabase}
			message.send(myError.toString(), typeError, login, login)
		}
		message.send(users.toString(), typeAllUsers, login, login)
	}

	//сообщить пользователям что мы появились в сети
	uStatus := newStatus(login, statusOnline)
	if !checkAnother(login) {
		message.send(uStatus.toString(), typeStatus, login, globalMessage)
	}

	//выгрузить архив сообщений за последний день
	{
		var arc = ArchiveMessages{TimeStart: time.Now(), TimeEnd: time.Now().AddDate(0, 0, -1)}
		err := arc.takeArchive(db)
		if err != nil {
			var myError = ErrorMessage{Error: errNoAccessWithDatabase}
			message.send(myError.toString(), typeError, login, login)
		}
		message.send(arc.toString(), typeArchive, login, login)
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			delConn(login, conn)
			if !checkAnother(login) {
				uStatus.Status = statusOffline
				message.send(uStatus.toString(), typeStatus, login, globalMessage)
			}
			return
		}
		fmt.Println(ConnectedUsers)

		typeMess, err := setTypeMessage(p)
		if err != nil {
			var myError = ErrorMessage{Error: errUnregisteredMessageType}
			message.send(myError.toString(), typeError, login, login)
		}

		switch typeMess {
		case typeMessage:
			{
				var messageFrom Message
				err = messageFrom.toStruct(p)
				if err != nil {
					var myError = ErrorMessage{Error: errCantConvertJSON}
					message.send(myError.toString(), typeError, login, login)
				}
				if err := saveMessage(messageFrom, db); err != nil {
					var myError = ErrorMessage{Error: errNoAccessWithDatabase}
					message.send(myError.toString(), typeError, login, login)
				}
				message.send(messageFrom.toString(), typeMessage, messageFrom.From, messageFrom.To)
			}
		case typeArchive:
			{
				var arc ArchiveMessages
				err = arc.toStruct(p)
				if err != nil {
					var myError = ErrorMessage{Error: errCantConvertJSON}
					message.send(myError.toString(), typeError, login, login)
				}
				err = arc.takeArchive(db)
				if err != nil {
					var myError = ErrorMessage{Error: errNoAccessWithDatabase}
					message.send(myError.toString(), typeError, login, login)
				}
				message.send(arc.toString(), typeArchive, login, login)
				continue
			}
		}

	}
}
