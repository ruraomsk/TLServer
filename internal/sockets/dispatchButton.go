package sockets

import (
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/ag-server/comm"
)

var DispatchMessageFromTechArm chan DBMessage

type DBMessage struct {
	Idevice int
	Data    interface{}
}

//DispatchControl отправка команды на устройство
func DispatchControl(arm comm.CommandARM) map[string]interface{} {
	var mess = tcpConnect.TCPMessage{User: arm.User, Type: tcpConnect.TypeDispatch, Id: arm.ID}
	mess.Data = arm
	tcpConnect.SendMessageToTCPServer <- mess
	for {
		tcpResp := <-tcpConnect.SendToUserResp
		if tcpResp.User == arm.User && tcpResp.Id == arm.ID {
			resp := make(map[string]interface{})
			resp["status"] = tcpResp.Status
			resp["command"] = arm
			return resp
		} else {
			tcpConnect.SendRespTCPMess <- tcpResp
		}
	}
}
