package deviceLog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
)

//DeviceLog описание таблицы, храняшей лог от устройств
type DeviceLog struct {
	Time    time.Time `json:"time"`    //время записи
	ID      int       `json:"id"`      //id устройства которое прислало информацию
	Text    string    `json:"text"`    //информация о событие
	Devices BusyArm   `json:"devices"` //информация о девайсе
}

//LogDeviceInfo структура запроса пользователя за данными в бд
type LogDeviceInfo struct {
	Devices   []BusyArm `json:"devices"`   //информация о девайсах
	TimeStart time.Time `json:"timeStart"` //время начала отсчета
	TimeEnd   time.Time `json:"timeEnd"`   //время конца отсчета
	structStr string    //строка для запроса в бд
}

//BusyArm информация о занятом перекрестке
type BusyArm struct {
	Region      string `json:"region"`      //регион
	Area        string `json:"area"`        //район
	ID          int    `json:"ID"`          //ID
	Description string `json:"description"` //описание
	Idevice     int    `json:"idevice"`     //уникальный номер устройства
}

//toStr конвертировать в структуру
func (busyArm *BusyArm) toStruct(str string) (err error) {
	err = json.Unmarshal([]byte(str), busyArm)
	if err != nil {
		return err
	}
	return nil
}

//DisplayDeviceLog формирование начальной информации отображения логов устройства
func DisplayDeviceLog(mapContx map[string]string, db *sqlx.DB) u.Response {
	var devices []BusyArm
	var sqlStr = fmt.Sprintf(`SELECT distinct on (crossinfo->'region', crossinfo->'area', crossinfo->'ID') crossinfo, id FROM public.logdevice`)
	if mapContx["region"] != "*" {
		sqlStr += fmt.Sprintf(` WHERE crossinfo::jsonb @> '{"region": "%v"}'::jsonb`, mapContx["region"])
	}
	rowsDevice, err := db.Query(sqlStr)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "connection to DB error. Please try again")
	}
	for rowsDevice.Next() {
		var (
			tempDev BusyArm
			infoStr string
			idevice int
		)
		err := rowsDevice.Scan(&infoStr, &idevice)
		if err != nil {
			logger.Error.Println("|Message: Incorrect data ", err.Error())
			return u.Message(http.StatusInternalServerError, "incorrect data. Please report it to Admin")
		}
		err = tempDev.toStruct(infoStr)
		tempDev.Idevice = idevice
		if err != nil {
			logger.Error.Println("|Message: Data can't convert ", err.Error())
			return u.Message(http.StatusInternalServerError, "data can't convert. Please report it to Admin")
		}
		devices = append(devices, tempDev)
	}
	resp := u.Message(http.StatusOK, "list of device")
	resp.Obj["devices"] = devices

	return resp
}

//DisplayDeviceLogInfo обработчик запроса пользователя, выгрузка логов за запрошенный период
func DisplayDeviceLogInfo(arms LogDeviceInfo, db *sqlx.DB) u.Response {
	var deviceLogs []DeviceLog
	if len(arms.Devices) <= 0 {
		return u.Message(http.StatusBadRequest, "no one devices selected")
	}
	for _, arm := range arms.Devices {
		sqlStr := fmt.Sprintf(`SELECT tm, id, txt FROM public.logdevice WHERE crossinfo::jsonb @> '{"ID": %v, "area": "%v", "region": "%v"}'::jsonb and tm > '%v' and tm < '%v'`, arm.ID, arm.Area, arm.Region, arms.TimeStart.Format("2006-01-02 15:04"), arms.TimeEnd.Format("2006-01-02 15:04"))
		rowsDevices, err := db.Query(sqlStr)
		if err != nil {
			return u.Message(http.StatusInternalServerError, "Connection to DB error. Please try again")
		}
		for rowsDevices.Next() {
			var tempDev DeviceLog
			err := rowsDevices.Scan(&tempDev.Time, &tempDev.ID, &tempDev.Text)
			if err != nil {
				logger.Error.Println("|Message: Incorrect data ", err.Error())
				return u.Message(http.StatusInternalServerError, "incorrect data. Please report it to Admin")
			}
			tempDev.Devices = arm
			deviceLogs = append(deviceLogs, tempDev)
		}
	}
	if deviceLogs == nil {
		deviceLogs = make([]DeviceLog, 0)
	}
	resp := u.Message(http.StatusOK, "get device Log")
	resp.Obj["deviceLogs"] = deviceLogs
	return resp
}
