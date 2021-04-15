package chat

import (
	"github.com/ruraomsk/TLServer/internal/model/data"
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
	From    string    `json:"from"`    //от какого пользователя
	To      string    `json:"to"`      //какому пользователю
	Message string    `json:"message"` //сообщение
	Time    time.Time `json:"time"`    //время отправки сообщения
}

//SaveMessage сохранение сообщения в БД
func (m *Message) SaveMessage() error {
	db, id := data.GetDB()
	defer data.FreeDB(id)
	_, err := db.Exec(`INSERT INTO public.chat (time, fromu, tou, message) VALUES ($1, $2, $3, $4)`, m.Time, m.From, m.To, m.Message)
	if err != nil {
		return err
	}
	return nil
}
