package chat

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"time"
)

type ArchiveMessages struct {
	Messages  []Message `json:"messages"`
	To        string    `json:"to"`
	TimeStart time.Time `json:"timeStart"` //время начала отсчета
	TimeEnd   time.Time `json:"timeEnd"`   //время конца отсчета
}

func (a *ArchiveMessages) takeArchive(db *sqlx.DB) error {
	rows, err := db.Query(`SELECT time, fromu, tou, message FROM public.chat WHERE time < $1 AND time > $2 AND tou = $3`, a.TimeStart.Format("2006-01-02 15:04:05"), a.TimeEnd.Format("2006-01-02 15:04:05"), a.To)
	if err != nil {
		return err
	}
	for rows.Next() {
		var tempMess Message
		_ = rows.Scan(&tempMess.Time, &tempMess.From, &tempMess.To, &tempMess.Message)
		a.Messages = append(a.Messages, tempMess)
	}
	return nil
}

func (a *ArchiveMessages) toString() string {
	raw, _ := json.Marshal(a)
	return string(raw)
}

func (a *ArchiveMessages) toStruct(str []byte) (err error) {
	err = json.Unmarshal(str, a)
	if err != nil {
		return err
	}
	return nil
}
