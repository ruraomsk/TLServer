package data

import (
	"../logger"
	"../tcpConnect"
	u "../utils"
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/ag-server/comm"
	"strings"
)

func DispatchControl(arm comm.CommandARM, mapContx map[string]string) map[string]interface{} {
	var (
		err        error
		armMessage tcpConnect.ArmCommandMessage
	)
	armMessage.CommandStr, err = armControlMarshal(arm)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal ArmControlData information: ", err.Error())
		return u.Message(false, "Failed to Marshal ArmControlData information")
	}
	armMessage.User = mapContx["login"]
	fmt.Println(armMessage)
	tcpConnect.ArmCommandChan <- armMessage
	for {
		chanRespond := <-tcpConnect.ArmCommandChan
		if strings.Contains(armMessage.User, mapContx["login"]) {
			if chanRespond.Message == "ok" {
				return u.Message(true, "ArmCommand send to server")
			} else {
				return u.Message(false, "TCP Server not responding")
			}
		}
	}
}

//armControlMarshal преобразовать структуру в строку
func armControlMarshal(arm comm.CommandARM) (str string, err error) {
	newByte, err := json.Marshal(arm)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}
