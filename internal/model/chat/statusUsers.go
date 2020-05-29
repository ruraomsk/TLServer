package chat

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
)

type StatusUser struct {
	Type   string `json:"type"`
	User   string `json:"user"`
	Status string `json:"status"`
}

type AllUsersStatus struct {
	Type  string       `json:"type"`
	Users []StatusUser `json:"users"`
}

func (s *StatusUser) toString() ([]byte, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return raw, err
}

func (s *StatusUser) send(status string) {
	s.Status = status
	raw, _ := s.toString()
	WriteAll <- raw
}

func newStatus(login string) *StatusUser {
	return &StatusUser{Type: statusInfo, User: login}
}

func (a *AllUsersStatus) setStatus() {
	for i, user := range a.Users {
		a.Users[i].Type = statusInfo
		for _, name := range Connections {
			if name == user.User {
				a.Users[i].Status = statusOnline
				break
			}
		}
	}
}

func (a *AllUsersStatus) getAllUsers(db *sqlx.DB) error {
	var (
		tempUser StatusUser
	)
	rows, err := db.Query(`SELECT login FROM public.accounts`)
	if err != nil {
		return err
	}
	for rows.Next() {
		_ = rows.Scan(&tempUser.User)
		tempUser.Status = statusOffline
		a.Users = append(a.Users, tempUser)
	}
	a.Type = allUsers
	a.setStatus()
	return nil
}
