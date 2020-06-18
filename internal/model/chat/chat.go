package chat

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

var writeChatMess chan chatSokResponse
var chatConnUsers map[*websocket.Conn]userInfo
var UserLogoutChat chan string

const pingPeriod = time.Second * 5

//Reader обработчик соединений (работа с чатом)
func Reader(conn *websocket.Conn, login string, db *sqlx.DB) {
	//-----------------------------------------------------------------------------------------------------------------------------------------------------------
	uInfo := userInfo{User: login, Status: statusOnline}

	//проверяем есть ли еще сокеты этого пользователя, если нет отправляем статус online
	if !checkOnline(login) {
		resp := newChatMess(typeStatus, conn, nil, uInfo)
		resp.Data["user"] = uInfo.User
		resp.Data["status"] = uInfo.Status
		resp.send()
	}

	chatConnUsers[conn] = uInfo
	//-----------------------------------------------------------------------------------------------------------------------------------------------------------

	//выгрузить список доступных пользователей
	{
		var users AllUsersStatus
		err := users.getAllUsers(db)
		if err != nil {
			resp := newChatMess(typeError, conn, nil, uInfo)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		resp := newChatMess(typeAllUsers, conn, nil, uInfo)
		resp.Data[typeAllUsers] = users.Users
		resp.send()
	}

	//выгрузить архив сообщений за последний день
	{
		var arc = ArchiveMessages{TimeStart: time.Now(), TimeEnd: time.Now().AddDate(0, 0, -1), To: globalMessage}
		err := arc.takeArchive(db)
		if err != nil {
			resp := newChatMess(typeError, conn, nil, uInfo)
			resp.Data["message"] = ErrorMessage{Error: errNoAccessWithDatabase}
			resp.send()
		}
		resp := newChatMess(typeArchive, conn, nil, uInfo)
		resp.Data[typeArchive] = arc
		resp.send()
	}
	fmt.Println("chat : ", chatConnUsers)
	for {
		typed, p, err := conn.ReadMessage()
		fmt.Println(typed)
		if err != nil {
			fmt.Println(err.Error())
			resp := newChatMess(typeClose, conn, nil, uInfo)
			resp.send()
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := newChatMess(typeError, conn, nil, uInfo)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}

		switch typeSelect {
		case typeMessage:
			{
				var mF Message
				_ = json.Unmarshal(p, &mF)
				if err := mF.SaveMessage(db); err != nil {
					resp := newChatMess(typeError, conn, nil, uInfo)
					resp.Data["message"] = ErrorMessage{Error: errNoAccessWithDatabase}
					resp.send()
				}
				resp := newChatMess(typeMessage, conn, nil, uInfo)
				resp.Data["message"] = mF.Message
				resp.Data["time"] = mF.Time
				resp.Data["from"] = mF.From
				resp.Data["to"] = mF.To
				resp.to = mF.To
				resp.send()
			}
		case typeArchive:
			{
				var arc ArchiveMessages
				err = arc.toStruct(p)
				if err != nil {
					resp := newChatMess(typeError, conn, nil, uInfo)
					resp.Data["message"] = ErrorMessage{Error: errCantConvertJSON}
					resp.send()
				}
				err = arc.takeArchive(db)
				if err != nil {
					resp := newChatMess(typeError, conn, nil, uInfo)
					resp.Data["message"] = ErrorMessage{Error: errNoAccessWithDatabase}
					resp.send()
				}
				resp := newChatMess(typeArchive, conn, nil, uInfo)
				resp.Data[typeArchive] = arc
				resp.send()
			}
		}

	}
}

//Broadcast обработчик сообщений (работа с чатом)
func CBroadcast() {
	chatConnUsers = make(map[*websocket.Conn]userInfo)
	writeChatMess = make(chan chatSokResponse, 1)
	UserLogoutChat = make(chan string)

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			{
				for conn := range chatConnUsers {
					_ = conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		case msg := <-writeChatMess:
			{
				switch msg.Type {
				case typeStatus:
					{
						for conn, uInfo := range chatConnUsers {
							if uInfo.User != msg.userInfo.User {
								_ = conn.WriteJSON(msg)
								//err != nil {
								//	//delete(chatConnUsers, conn)
								//	deleteUser(msg)
								//	_ = conn.Close()
								//}
							}
						}
					}
				case typeArchive:
					{
						_ = msg.conn.WriteJSON(msg)
						// err != nil {
						//	//delete(chatConnUsers, conn)
						//	deleteUser(msg)
						//	_ = msg.conn.Close()
						//}
					}
				case typeAllUsers:
					{
						for conn := range chatConnUsers {
							_ = conn.WriteJSON(msg)
							//; err != nil {
							//	//delete(chatConnUsers, conn)
							//	deleteUser(msg)
							//	_ = conn.Close()
							//}
						}
					}

				case typeMessage:
					{
						if msg.to == "Global" {
							for conn := range chatConnUsers {
								_ = conn.WriteJSON(msg)
								//; err != nil {
								//	//delete(chatConnUsers, conn)
								//	deleteUser(msg)
								//	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "asd"))
								//	//_ = conn.Close()
								//}
							}
						}
					}

				case typeClose:
					{
						delete(chatConnUsers, msg.conn)
						if !checkOnline(msg.userInfo.User) {
							msg.Type = typeStatus
							msg.Data["user"] = msg.userInfo.User
							msg.Data["status"] = statusOffline
							for conn := range chatConnUsers {
								_ = conn.WriteJSON(msg)
							}
						}
					}
				default:
					{
						_ = msg.conn.WriteJSON(msg)
						//; err != nil {
						//	//delete(chatConnUsers, msg.conn)
						//	deleteUser(msg)
						//	_ = msg.conn.Close()
						//}
					}
				}
			}
		case login := <-UserLogoutChat:
			{
				//disconnected := false
				for conn, infoUser := range chatConnUsers {
					if infoUser.User == login {
						_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "пользователь вышел из системы"))
						//disconnected = true
					}
				}
				//if disconnected {
				//	user := userInfo{User: login, Status: statusOffline}
				//	resp := newChatMess(typeStatus, nil, nil, user)
				//	resp.Data["user"] = user.User
				//	resp.Data["status"] = user.Status
				//	for conn := range chatConnUsers {
				//		_ = conn.WriteJSON(resp)
				//	}
			}
		}
	}
}

//func deleteUser(c chatSokResponse) {
//	delete(chatConnUsers, c.conn)
//	if !checkOnline(c.userInfo.User) {
//		resp := newChatMess(typeStatus, c.conn, nil, c.userInfo)
//		resp.Data["user"] = c.userInfo.User
//		resp.Data["status"] = statusOffline
//		resp.send()
//	}
//}
