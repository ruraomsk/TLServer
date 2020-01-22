package utils

import (
	"../logger"
	"fmt"
)

func WriteLogMessage(ip string, url string, data map[string]interface{}, info interface{}) {
	mapContx := ParserInterface(info)
	login := mapContx["login"]
	if login == "" {
		if _, ok := data["logLogin"]; ok {
			login = fmt.Sprintf("%v", data["logLogin"])
		} else {
			if _, ok := data["login"]; ok {
				login = fmt.Sprintf("%v", data["login"])
			} else {
				login = "-"
			}
		}
	}
	if data["status"] == false {
		logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, login, url, data["message"])
	} else {
		logger.Info.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", ip, login, url, data["message"])
	}
}
