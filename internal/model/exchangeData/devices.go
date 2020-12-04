package exchangeData

import (
	"github.com/JanFant/TLServer/internal/model/device"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/ruraomsk/ag-server/pudge"
	"net/http"
)

func GetDevices(iDevice []int) u.Response {
	var (
		DevicesList     = make([]pudge.Controller, 0)
		mapActivDevices = make(map[int]device.DevInfo)
	)

	device.GlobalDevices.Mux.Lock()
	mapActivDevices = device.GlobalDevices.MapDevices
	device.GlobalDevices.Mux.Unlock()

	for _, numDev := range iDevice {
		if dev, ok := mapActivDevices[numDev]; ok {
			DevicesList = append(DevicesList, dev.Controller)
		}
	}

	resp := u.Response{Code: http.StatusOK, Obj: map[string]interface{}{}}
	resp.Obj["data"] = DevicesList
	return resp
}
