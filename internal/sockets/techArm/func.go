package techArm

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/device"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
)

//getCross запроса состояния перекрестков
func getCross(reg int, db *sqlx.DB) []CrossInfo {
	var (
		temp    CrossInfo
		crosses []CrossInfo
		sqlStr  = `SELECT region,
 					area, 
 					id,
  					idevice, 
  					describ, 
  					subarea, 
  					state->'arrays'->'type',
  					state->'phone' 
  					FROM public.cross`
	)
	if reg != -1 {
		sqlStr += fmt.Sprintf(` WHERE region = %v`, reg)
	}
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
			&temp.Phone)
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
