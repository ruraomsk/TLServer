package data

import (
	"fmt"
	"github.com/JanFant/TLServer/license"
	u "github.com/JanFant/TLServer/utils"
	"github.com/jordan-wright/email"
	"net/smtp"
)

type EmailJS struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

func SendEmail(emailInfo EmailJS, mapContx map[string]string) map[string]interface{} {
	e := email.NewEmail()

	e.From = fmt.Sprintf("%s <%s>", mapContx["login"], "AsudServ@gmail.com")
	e.To = []string{emailInfo.To}
	e.Subject = fmt.Sprintf("Tech Support from server: %s", license.LicenseFields.Name)
	e.Text = []byte(emailInfo.Text)

	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", "AsudServ@gmail.com", "H49qFgUDTzSUQFYmoVwf", "smtp.gmail.com"))
	if err != nil {
		return u.Message(false, fmt.Sprint("Failed send email: ", err.Error()))
	}

	resp := u.Message(true, "email sent")
	return resp
}
