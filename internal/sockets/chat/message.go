package chat

import (
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	typeError     = "error"
	statusOnline  = "online"
	statusOffline = "offline"

	typeClose = "close"

	typeMessage  = "message"
	typeArchive  = "archive"
	typeStatus   = "status"
	typeAllUsers = "users"
	//typeCheckStatus = "checkStatus"
	globalMessage = "Global"

	//errCantConvertJSON         = "cant convert JSON"
	errNoAccessWithDatabase = "Нет связи с сервором БД"
	errParseType            = "Сервер не смог обработать запрос"
)

//chatResponse структура для отправки сообщений (chat)
type chatResponse struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
	to   string
	from string
}

//newChatMess создание нового сообщения
func newChatMess(mType string, data map[string]interface{}) chatResponse {
	var resp chatResponse
	resp.Type = mType
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
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
