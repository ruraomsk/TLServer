package utils

import (
	"../logger"
	"fmt"
	"strings"
)

func WriteLogMessage(status interface{}, ip string, data map[string]interface{}, info interface{}) {
	mapContx := ParserInterface(info)

	login := mapContx["login"]
	if login == "" {
		login = fmt.Sprint(data["login"])
		if login == "" {
			login = "-"
		}
	}
	if !strings.Contains(fmt.Sprint(data["message"]), "Update box data") {
		if status == false {
			logger.Warning.Printf("IP: %s Login: %s Message: %v", ip, login, data["message"])
		} else {
			logger.Info.Printf("IP: %s Login: %s Message: %v", ip, login, data["message"])
		}
	}
}
