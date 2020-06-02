package chat

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	typeMessage   = "message"
	typeArchive   = "archive"
	typeError     = "error"
	typeStatus    = "status"
	typeAllUsers  = "users"
	statusOnline  = "online"
	statusOffline = "offline"
	globalMessage = "Global"

	errNoAccessWithDatabase    = "no access with database"
	errCantConvertJSON         = "cant convert JSON"
	errUnregisteredMessageType = "unregistered message type"
)

type Message struct {
	From    string    `json:"from"`
	To      string    `json:"to"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func (m *Message) toStruct(str []byte) (err error) {
	err = json.Unmarshal(str, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Message) toString() string {
	raw, _ := json.Marshal(m)
	return string(raw)
}

type SendMessage struct {
	from string
	to   string
	conn *websocket.Conn
	Type string `json:"type"`
	Data string `json:"data"`
}

func (sm *SendMessage) send(data, mType, from, to string) {
	sm.Data = data
	sm.Type = mType
	sm.from = from
	sm.to = to
	WriteSendMessage <- *sm
}

func (sm *SendMessage) toString() (string, error) {
	raw, err := json.Marshal(sm)
	return string(raw), err
}

type ErrorMessage struct {
	Error string `json:"error"`
}

func (e *ErrorMessage) toString() string {
	raw, _ := json.Marshal(e)
	return string(raw)
}

func (m *Message) SaveMessage(db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO public.chat (time, fromu, tou, message) VALUES ($1, $2, $3, $4)`, m.Time, m.From, m.To, m.Message)
	if err != nil {
		return err
	}
	return nil
}

func setTypeMessage(raw []byte) (string, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(raw, &temp); err != nil {
		return "", err
	}
	return fmt.Sprint(temp["type"]), nil
}
