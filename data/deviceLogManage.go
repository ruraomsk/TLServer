package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/JanFant/TLServer/logger"
	u "github.com/JanFant/TLServer/utils"
)

//DeviceLog описание таблицы храняшей лог от устройств
type DeviceLog struct {
	Time    time.Time `json:"time"`    //время записи
	ID      int       `json:"id"`      //id устройства которое прислало информацию
	Text    string    `json:"text"`    //информация о событие
	Devices BusyArm   `json:"devices"` //информация о девайсе
}

//DeviceLogInfo струтура запроса пользователя за данными в бд
type DeviceLogInfo struct {
	Devices   []BusyArm `json:"devices"`   //информация о девайсах
	TimeStart string    `json:"timeStart"` //время начала отсчета
	TimeEnd   string    `json:"timeEnd"`   //время конца отсчета
	structStr string    //строка для запроса в бд
}

//DisplayDeviceLog формирование начальной информации отображения логов устройства
func DisplayDeviceLog(mapContx map[string]string) map[string]interface{} {
	var (
		devices  []BusyArm
		fillInfo FillingInfo
	)
	fillInfo.User = mapContx["login"]
	FillingDeviceChan <- fillInfo
	for {
		chanRespond := <-FillingDeviceChan
		if strings.Contains(chanRespond.User, fillInfo.User) {
			if chanRespond.Status {
				break
			} else {
				return u.Message(false, "Incorrect data in logDevice table. Please report it to Admin")
			}
		}
	}
	var sqlStr string
	if mapContx["region"] == "*" {
		sqlStr = fmt.Sprintf("SELECT distinct crossinfo FROM %v", GlobalConfig.DBConfig.LogDeviceTable)
	} else {
		sqlStr = fmt.Sprintf(`SELECT distinct crossinfo FROM %v where crossinfo::jsonb @> '{"region": "%v"}'::jsonb`, GlobalConfig.DBConfig.LogDeviceTable, mapContx["region"])
	}
	rowsDevice, err := GetDB().Raw(sqlStr).Rows()
	if err != nil {
		return u.Message(false, "Connection to DB error. Please try again")
	}
	for rowsDevice.Next() {
		var (
			tempDev BusyArm
			infoStr string
		)
		err := rowsDevice.Scan(&infoStr)
		if err != nil {
			logger.Error.Println("|Message: Incorrect data ", err.Error())
			return u.Message(false, "Incorrect data. Please report it to Admin")
		}
		err = tempDev.toStruct(infoStr)
		if err != nil {
			logger.Error.Println("|Message: Data can't convert ", err.Error())
			return u.Message(false, "Data can't convert. Please report it to Admin")
		}
		devices = append(devices, tempDev)
	}
	resp := u.Message(true, "List of device")
	resp["devices"] = devices
	CacheInfo.mux.Lock()
	resp["regionInfo"] = CacheInfo.mapRegion
	resp["areaInfo"] = CacheInfo.mapArea
	CacheInfo.mux.Unlock()
	return resp
}

//DisplayDeviceLogInfo обработчик запроса пользователя, выгрузка логов за запрошенный период
func DisplayDeviceLogInfo(arms DeviceLogInfo, mapContx map[string]string) map[string]interface{} {
	var (
		deviceLogs []DeviceLog
		fillInfo   FillingInfo
	)
	fillInfo.User = mapContx["login"]
	FillingDeviceChan <- fillInfo
	for {
		chanRespond := <-FillingDeviceChan
		if chanRespond.User == fillInfo.User {
			if chanRespond.Status {
				break
			} else {
				return u.Message(false, "Incorrect data in logDevice table. Please report it to Admin")
			}
		}
	}
	arms.TimeStart = "2010-03-19T11:50:31.697736Z"
	arms.TimeEnd = "2030-03-19T11:50:31.697736Z"
	for _, arm := range arms.Devices {
		sqlStr := fmt.Sprintf(`SELECT tm, id, txt FROM %v where crossinfo::jsonb @> '{"ID": %v, "area": "%v", "region": "%v"}'::jsonb and tm > '%v' and tm < '%v'`, GlobalConfig.DBConfig.LogDeviceTable, arm.ID, arm.Area, arm.Region, arms.TimeStart, arms.TimeEnd)
		rowsDevices, err := GetDB().Raw(sqlStr).Rows()
		if err != nil {
			return u.Message(false, "Connection to DB error. Please try again")
		}
		for rowsDevices.Next() {
			var tempDev DeviceLog
			err := rowsDevices.Scan(&tempDev.Time, &tempDev.ID, &tempDev.Text)
			if err != nil {
				logger.Error.Println("|Message: Incorrect data ", err.Error())
				return u.Message(false, "Incorrect data. Please report it to Admin")
			}
			tempDev.Devices = arm
			deviceLogs = append(deviceLogs, tempDev)
		}
	}
	if deviceLogs == nil {
		deviceLogs = make([]DeviceLog, 0)
	}
	resp := u.Message(true, "Get device Log")
	resp["deviceLogs"] = deviceLogs
	return resp
}
