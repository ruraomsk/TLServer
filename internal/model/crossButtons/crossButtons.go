package crossButtons

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JanFant/newTLServer/internal/app/tcpConnect"
	"github.com/JanFant/newTLServer/internal/model/logger"
	u "github.com/JanFant/newTLServer/internal/utils"
	"github.com/ruraomsk/ag-server/comm"
)

//DispatchControl отправка команды на устройство
func DispatchControl(arm comm.CommandARM, mapContx map[string]string) u.Response {
	var (
		err        error
		armMessage tcpConnect.ArmCommandMessage
	)
	arm.User = mapContx["login"]
	armMessage.CommandStr, err = armControlMarshal(arm)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal ArmControlData information: ", err.Error())
		return u.Message(http.StatusBadRequest, "failed to Marshal ArmControlData information")
	}
	armMessage.User = mapContx["login"]
	tcpConnect.ArmCommandChan <- armMessage
	for {
		chanRespond := <-tcpConnect.ArmCommandChan
		if strings.Contains(armMessage.User, mapContx["login"]) {
			if chanRespond.Message == "ok" {
				return u.Message(http.StatusOK, fmt.Sprintf("command %v send to server", armMessage.CommandStr))
			} else {
				return u.Message(http.StatusInternalServerError, "TCP Server not responding")
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
