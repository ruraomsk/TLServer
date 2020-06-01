package techSupport

import (
	"fmt"
	"net/http"
	"net/smtp"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/jordan-wright/email"
)

type EmailJS struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

func SendEmail(emailInfo EmailJS, login, companyName string) u.Response {
	e := email.NewEmail()

	e.From = fmt.Sprintf("%s <%s>", login, "AsudServ@gmail.com")
	e.To = []string{emailInfo.To}
	e.Subject = fmt.Sprintf("Tech Support from server: %s", companyName)
	e.Text = []byte(emailInfo.Text)

	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", "AsudServ@gmail.com", "H49qFgUDTzSUQFYmoVwf", "smtp.gmail.com"))
	if err != nil {
		return u.Message(http.StatusInternalServerError, fmt.Sprint("Failed send email: ", err.Error()))
	}

	resp := u.Message(http.StatusOK, "email sent")
	return resp
}
