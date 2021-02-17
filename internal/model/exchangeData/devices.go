package exchangeData

import (
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/internal/model/license"
	u "github.com/ruraomsk/TLServer/internal/utils"
	"github.com/ruraomsk/ag-server/pudge"
	"net/http"
)

func GetDevices(iDevice []int) u.Response {
	var (
		devicesList     = make([]pudge.Controller, 0)
		mapActivDevices = make(map[int]device.DevInfo)
	)

	device.GlobalDevices.Mux.Lock()
	mapActivDevices = device.GlobalDevices.MapDevices
	device.GlobalDevices.Mux.Unlock()

	for _, numDev := range iDevice {
		if dev, ok := mapActivDevices[numDev]; ok {
			devicesList = append(devicesList, dev.Controller)
		}
	}

	//обережим количество устройств по количеству доступному в лицензии
	numDev := license.LicenseFields.NumDev
	if len(devicesList) > numDev {
		devicesList = devicesList[:numDev]
	}

	resp := u.Response{Code: http.StatusOK, Obj: map[string]interface{}{}}
	resp.Obj["data"] = devicesList
	return resp
}
