package data

import (
	"fmt"
	"github.com/JanFant/TLServer/logger"
	u "github.com/JanFant/TLServer/utils"
	"strings"
	"time"
)

//DeviceLog описание БД храняшей лог от устройств
type DeviceLog struct {
	Time time.Time `json:"time"`
	ID   int       `json:"id"`
	Text string    `json:"text"`
}

//DisplayDeviceLog формирование начальной информации
func DisplayDeviceLog(mapContx map[string]string) map[string]interface{} {
	var (
		devices  []tlInfo
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

	sqlStr := fmt.Sprintf("SELECT distinct crossinfo FROM %v", GlobalConfig.DBConfig.LogDeviceTable)
	rowsDevice, err := GetDB().Raw(sqlStr).Rows()
	if err != nil {
		return u.Message(false, "Connection to DB error. Please try again")
	}
	for rowsDevice.Next() {
		var (
			tempDev tlInfo
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
	return resp
}
