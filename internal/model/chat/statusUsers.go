package chat

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
)

//AllUsersStatus список всех пользователей
type AllUsersStatus struct {
	Users []StatusUser `json:"users"`
}

//StatusUser информация о статусе юзера
type StatusUser struct {
	User   string `json:"user"`
	Status string `json:"status"`
}

//toString преобразование в строку
func (s *StatusUser) toString() string {
	raw, _ := json.Marshal(s)
	return string(raw)
}

//checkAnother проверка есть ли еще подключенные сокеты у пользователя
func checkAnother(login string) bool {
	if len(ConnectedUsers[login]) > 0 {
		return true
	}
	return false
}

//newStatus создать статус пользователя
func newStatus(login, status string) *StatusUser {
	return &StatusUser{User: login, Status: status}
}

//setStatus установить статус пользователя
func (a *AllUsersStatus) setStatus() {
	for i, user := range a.Users {
		for name, _ := range ConnectedUsers {
			if name == user.User {
				a.Users[i].Status = statusOnline
				break
			}
		}
	}
}

//getAllUsers запросить пользователей из БД
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
	a.setStatus()
	return nil
}

//toString преобразовать в строку
func (a *AllUsersStatus) toString() string {
	raw, _ := json.Marshal(a)
	return string(raw)
}
