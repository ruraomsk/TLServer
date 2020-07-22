package chat

import (
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	typeMessage  = "message"
	typeArchive  = "archive"
	typeError    = "error"
	typeStatus   = "status"
	typeAllUsers = "users"
	//typeCheckStatus = "checkStatus"
	statusOnline  = "online"
	statusOffline = "offline"
	globalMessage = "Global"
	typeClose     = "close"

	errNoAccessWithDatabase    = "no access with database"
	errCantConvertJSON         = "cant convert JSON"
	errUnregisteredMessageType = "unregistered message type"
)

//chatSokResponse структура для отправки сообщений (chat)
type chatSokResponse struct {
	Type     string                 `json:"type"`
	Data     map[string]interface{} `json:"data"`
	conn     *websocket.Conn        `json:"-"`
	userInfo userInfo               `json:"-"`
	to       string                 `json:"-"`
}

//newChatMess создание нового сообщения
func newChatMess(mType string, conn *websocket.Conn, data map[string]interface{}, info userInfo) chatSokResponse {
	var resp = chatSokResponse{Type: mType, conn: conn, userInfo: info}
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//send отправка с обработкой ошибки
func (m *chatSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v",
				m.conn.RemoteAddr(),
				m.userInfo.User,
				"chat",
				m.Data["message"])
		}()
	}
	writeChatMess <- *m
}

//Message структура для приема сообщений
type Message struct {
	From    string    `json:"from"`
	To      string    `json:"to"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

//SaveMessage сохранение сообщения в БД
func (m *Message) SaveMessage(db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO public.chat (time, fromu, tou, message) VALUES ($1, $2, $3, $4)`, m.Time, m.From, m.To, m.Message)
	if err != nil {
		return err
	}
	return nil
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

//closeMessage структура для закрытия
type closeMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
