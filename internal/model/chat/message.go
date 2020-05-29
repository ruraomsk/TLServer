package chat

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	messageInfo   = "message"
	errorMessage  = "error"
	statusInfo    = "status"
	statusOnline  = "online"
	statusOffline = "offline"
	allUsers      = "users"

	errNoAccessWithDatabase = "no access with database"
)

var Names UsersInfo

type UsersInfo struct {
	Type  string          `json:"type"`
	Users map[string]bool `json:"users"`
}

type Message struct {
	Type    string    `json:"type"`
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

func (m *Message) toString() ([]byte, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return raw, err
}

type ErrorMessage struct {
	Type  string `json:"type"`
	User  string `json:"user"`
	Error string `json:"error"`
}

func newErrorMessage(login, errorStr string) *ErrorMessage {
	return &ErrorMessage{User: login, Type: errorMessage, Error: errorStr}
}

func saveMessage(mess Message, db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO public.chat (time, fromu, tou, message) VALUES ($1, $2, $3, $4)`, mess.Time, mess.From, mess.To, mess.Message)
	if err != nil {
		return err
	}
	return nil
}
