package chat

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

var (
	messageInfo   = "message"
	messages      = "messages"
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

type PeriodMessage struct {
	Type      string    `json:"type"`
	Messages  []Message `json:"messages"`
	TimeStart time.Time `json:"timeStart"` //время начала отсчета
	TimeEnd   time.Time `json:"timeEnd"`   //время конца отсчета
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
	Error string `json:"error"`
}

func newErrorMessage(errorStr string) *ErrorMessage {
	return &ErrorMessage{Type: errorMessage, Error: errorStr}
}

func saveMessage(mess Message, db *sqlx.DB) error {
	_, err := db.Exec(`INSERT INTO public.chat (time, fromu, tou, message) VALUES ($1, $2, $3, $4)`, mess.Time, mess.From, mess.To, mess.Message)
	if err != nil {
		return err
	}
	return nil
}

func (p *PeriodMessage) takeMessages(db *sqlx.DB) error {
	p.Type = messages
	rows, err := db.Query(`SELECT time, fromu, tou, message FROM public.chat WHERE time < $1 AND time > $2`, p.TimeStart.Format("2006-01-02 15:04:05"), p.TimeEnd.Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}
	for rows.Next() {
		var tempMess Message
		tempMess.Type = messageInfo
		_ = rows.Scan(&tempMess.Time, &tempMess.From, &tempMess.To, &tempMess.Message)
		p.Messages = append(p.Messages, tempMess)
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
