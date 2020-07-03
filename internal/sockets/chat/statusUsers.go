package chat

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
)

//AllUsersStatus список всех пользователей
type AllUsersStatus struct {
	Users []userInfo `json:"users"`
}

//userInfo информация о статусе юзера
type userInfo struct {
	User   string `json:"user"`
	Status string `json:"status"`
}

//checkOnline проверка есть ли еще подключенный сокет
func checkOnline(login string) bool {
	for _, info := range chatConnUsers {
		if info.User == login {
			return true
		}
	}
	return false
}

//setStatus установить статус пользователя
func (a *AllUsersStatus) setStatus() {
	for i, user := range a.Users {
		for _, name := range chatConnUsers {
			if name.User == user.User {
				a.Users[i].Status = statusOnline
				break
			}
		}
	}
}

//getAllUsers запросить пользователей из БД
func (a *AllUsersStatus) getAllUsers(db *sqlx.DB) error {
	var (
		tempUser userInfo
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
