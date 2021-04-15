package techArm

import (
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/logger"
)

//getCross запроса состояния перекрестков
func getCross(reg int) []CrossInfo {
	var (
		temp    CrossInfo
		crosses []CrossInfo
		model   []byte
		sqlStr  = `SELECT region,
 					area, 
 					id,
  					idevice, 
  					describ, 
  					subarea, 
  					state->'arrays'->'type',
  					state->'phone',
       				state ->'status',
  					state->'Model' 
  					FROM public.cross`
	)
	if reg != -1 {
		sqlStr += fmt.Sprintf(` WHERE region = %v`, reg)
	}
	db, id := data.GetDB()
	defer data.FreeDB(id)
	rows, err := db.Query(sqlStr)
	if err != nil {
		logger.Error.Println("|IP: server |Login: server |Resource: /techArm |Message: Error get Cross from BD ", err.Error())
		return make([]CrossInfo, 0)
	}
	for rows.Next() {
		_ = rows.Scan(&temp.Region,
			&temp.Area,
			&temp.ID,
			&temp.Idevice,
			&temp.Describe,
			&temp.Subarea,
			&temp.ArrayType,
			&temp.Phone,
			&temp.StatusCode,
			&model)
		_ = json.Unmarshal(model, &temp.Model)
		temp.Status = data.CacheInfo.MapTLSost[temp.StatusCode].Description
		crosses = append(crosses, temp)
	}
	return crosses
}

//getCross запрос состояния устройств
func getDevice() []DevInfo {
	var (
		devices []DevInfo
		copyDev = make(map[int]device.DevInfo)
	)

	device.GlobalDevices.Mux.Lock()
	for key, c := range device.GlobalDevices.MapDevices {
		copyDev[key] = c
	}
	device.GlobalDevices.Mux.Unlock()

	for _, dev := range copyDev {
		var temp = DevInfo{Area: dev.Area, Region: dev.Region, Idevice: dev.Controller.ID, Device: dev.Controller}
		temp.ModeRdk = modeRDK[temp.Device.DK.RDK]
		temp.TechMode = texMode[temp.Device.TechMode]
		devices = append(devices, temp)
	}

	return devices
}
