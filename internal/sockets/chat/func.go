package chat

import (
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/model/data"
)

type clientInfo struct {
	status  string
	accInfo *accToken.Token
}

//userInfo информация о статусе юзера
type userInfo struct {
	User   string `json:"user"`   //пользователь
	Status string `json:"status"` //статус
}

//setStatus установить статус пользователя
func (u *userInfo) setStatus(h *HubChat) {
	for client := range h.clients {
		if client.clientInfo.accInfo.Login == u.User {
			u.Status = statusOnline
			break
		}
	}
}

//getAllUsers запросить пользователей из БД
func getAllUsers(h *HubChat) ([]userInfo, error) {
	var users []userInfo
	db, id := data.GetDB()
	defer data.FreeDB(id)
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
