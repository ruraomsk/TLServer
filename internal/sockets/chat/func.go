package chat

import "github.com/jmoiron/sqlx"

type clientInfo struct {
	login  string
	status string
	ip     string
}

//userInfo информация о статусе юзера
type userInfo struct {
	User   string `json:"user"`   //пользователь
	Status string `json:"status"` //статус
}

//setStatus установить статус пользователя
func (u *userInfo) setStatus(h *HubChat) {
	for client := range h.clients {
		if client.clientInfo.login == u.User {
			u.Status = statusOnline
			break
		}
	}
}

//getAllUsers запросить пользователей из БД
func getAllUsers(h *HubChat, db *sqlx.DB) ([]userInfo, error) {
	var users []userInfo
	rows, err := db.Query(`SELECT login FROM public.accounts`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tempUser = userInfo{Status: statusOffline}
		_ = rows.Scan(&tempUser.User)
		tempUser.setStatus(h)
		users = append(users, tempUser)
	}
	return users, nil
}
