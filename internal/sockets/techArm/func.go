package techArm

import (
	"encoding/json"
	"fmt"
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
func getDevice(db *sqlx.DB) []DevInfo {
	var (
		temp    DevInfo
		devices []DevInfo
		dStr    string
	)
	rows, err := db.Query(`SELECT c.region, 
									c.area, 
									c.idevice, 
									d.device 
									FROM public.cross as c, public.devices as d WHERE c.idevice IN(d.id);`)
	if err != nil {
		logger.Error.Println("|IP: server |Login: server |Resource: /techArm |Message: Error get Device from BD ", err.Error())
		return make([]DevInfo, 0)
	}
	for rows.Next() {
		_ = rows.Scan(&temp.Region, &temp.Area, &temp.Idevice, &dStr)
		_ = json.Unmarshal([]byte(dStr), &temp.Device)
		temp.ModeRdk = modeRDK[temp.Device.DK.RDK]
		temp.TechMode = texMode[temp.Device.TechMode]
		devices = append(devices, temp)
	}
	return devices
}
