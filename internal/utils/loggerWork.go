package utils

import (
	"fmt"

	"github.com/JanFant/TLServer/logger"
)

//writeLogMessage обработчик u.message преобразует сообщение для записи в лог файл
func writeLogMessage(ip string, url string, data map[string]interface{}, info interface{}) {
	mapContx := ParserInterface(info)
	login := mapContx["login"]
	if login == "" {
		if _, ok := data["login"]; ok {
			login = fmt.Sprintf("%v", data["login"])
		} else {
			login = "-"
		}
	}
	logger.Info.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, login, url, data["message"])
}
