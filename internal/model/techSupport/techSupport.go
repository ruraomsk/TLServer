package techSupport

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/internal/sockets/chat"
	"github.com/jmoiron/sqlx"
	"net/http"
	"net/smtp"
	"time"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/jordan-wright/email"
)

type EmailJS struct {
	Text string `json:"text"` //сообщение
}

//SendEmail подготовка и отправка сообщения на почту, с сохранением в бд
func SendEmail(emailInfo EmailJS, login, companyName, companyLoc string, db *sqlx.DB) u.Response {
	e := email.NewEmail()

	e.From = fmt.Sprintf("%s <%s>", login, "AsudServ@gmail.com")
	e.To = license.LicenseFields.TechEmail
	e.Subject = fmt.Sprintf("Tech Support from server: %v of %v ", companyName, companyLoc)
	e.Text = []byte(emailInfo.Text)

	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", "AsudServ@gmail.com", "H49qFgUDTzSUQFYmoVwf", "smtp.gmail.com"))
	if err != nil {
		return u.Message(http.StatusInternalServerError, fmt.Sprint("Failed send email: ", err.Error()))
	}

	mess := chat.Message{From: login, Time: time.Now(), To: data.AutomaticLogin}
	mess.Message = fmt.Sprintf("Пользователь %v обратился в техподдержку, время обращения ( %v ), с вопросом: %v", mess.From, mess.Time.Format("2006-01-02 15:04:05"), emailInfo.Text)
	_ = mess.SaveMessage(db)
	resp := u.Message(http.StatusOK, "email sent")
	return resp
}
