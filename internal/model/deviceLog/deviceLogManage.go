package deviceLog

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/accToken"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
)

//DeviceLog описание таблицы, храняшей лог от устройств
type DeviceLog struct {
	Time time.Time `json:"time"` //время записи
	ID   int       `json:"id"`   //id устройства которое прислало информацию
	Text string    `json:"text"` //информация о событие
	Type int       `json:"type"` //тип сообщения
}

//LogDeviceInfo структура запроса пользователя за данными в бд
type LogDeviceInfo struct {
	Devices   []BusyArm `json:"devices"`   //информация о девайсах
	TimeStart time.Time `json:"timeStart"` //время начала отсчета
	TimeEnd   time.Time `json:"timeEnd"`   //время конца отсчета
}

//BusyArm информация о занятом перекрестке
type BusyArm struct {
	Region      string `json:"region"`      //регион
	Area        string `json:"area"`        //район
	ID          int    `json:"ID"`          //ID
	Description string `json:"description"` //описание
	Idevice     int    `json:"idevice"`     //уникальный номер устройства
}

//shortInfo короткое описание
type shortInfo struct {
	Region      string `json:"region"`      //регион
	Area        string `json:"area"`        //район
	ID          int    `json:"ID"`          //ID
	Description string `json:"description"` //описание
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
func DisplayDeviceLog(accInfo *accToken.Token, db *sqlx.DB) u.Response {
	var devices []BusyArm
	var sqlStr = fmt.Sprintf(`SELECT distinct on (crossinfo->'region', crossinfo->'area', crossinfo->'ID', id) crossinfo, id FROM public.logdevice`)
	if accInfo.Region != "*" {
		sqlStr += fmt.Sprintf(` WHERE crossinfo::jsonb @> '{"region": "%v"}'::jsonb`, accInfo.Region)
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
		if tempDev.ID != 0 && tempDev.Area != "0" && tempDev.Region != "0" {
			devices = append(devices, tempDev)
		}
	}
	resp := u.Message(http.StatusOK, "list of device")
	resp.Obj["devices"] = devices

	return resp
}

//DisplayDeviceLogInfo обработчик запроса пользователя, выгрузка логов за запрошенный период
func DisplayDeviceLogInfo(arms LogDeviceInfo, db *sqlx.DB) u.Response {
	if len(arms.Devices) <= 0 {
		return u.Message(http.StatusBadRequest, "no one devices selected")
	}
	var mapDevice = make(map[string][]DeviceLog, 0)
	for _, arm := range arms.Devices {
		var (
			listDevicesLog []DeviceLog
			tempInfo       = shortInfo{ID: arm.ID, Area: arm.Area, Region: arm.Region, Description: arm.Description}
			rawByte, _     = json.Marshal(tempInfo) //перобразование структуру в строку для использования в ключе
		)
		mapDevice[string(rawByte)] = make([]DeviceLog, 0)
		sqlStr := fmt.Sprintf(`SELECT tm, id, txt, crossinfo->'type' FROM public.logdevice WHERE crossinfo::jsonb @> '{"ID": %v, "area": "%v", "region": "%v"}'::jsonb and tm > '%v' and tm < '%v' ORDER BY tm DESC`, arm.ID, arm.Area, arm.Region, arms.TimeStart.Format("2006-01-02 15:04:05"), arms.TimeEnd.Format("2006-01-02 15:04:05"))
		rowsDevices, err := db.Query(sqlStr)
		if err != nil {
			return u.Message(http.StatusInternalServerError, "Connection to DB error. Please try again")
		}
		for rowsDevices.Next() {
			var tempDev DeviceLog
			err := rowsDevices.Scan(&tempDev.Time, &tempDev.ID, &tempDev.Text, &tempDev.Type)
			if err != nil {
				logger.Error.Println("|Message: Incorrect data ", err.Error())
				return u.Message(http.StatusInternalServerError, "incorrect data. Please report it to Admin")
			}
			//tempDev.Devices = arm
			listDevicesLog = append(listDevicesLog, tempDev)
		}
		if len(listDevicesLog) == 0 {
			var tempDev DeviceLog
			sqlStr := fmt.Sprintf(`SELECT tm, id, txt, crossinfo->'type' FROM public.logdevice WHERE crossinfo::jsonb @> '{"ID": %v, "area": "%v", "region": "%v"}'::jsonb ORDER BY tm DESC LIMIT 1`, arm.ID, arm.Area, arm.Region)
			err = db.QueryRow(sqlStr).Scan(&tempDev.Time, &tempDev.ID, &tempDev.Text, &tempDev.Type)
			if err != nil {
				logger.Error.Println("|Message: Incorrect data ", err.Error())
				return u.Message(http.StatusInternalServerError, "incorrect data. Please report it to Admin")
			}
			listDevicesLog = append(listDevicesLog, tempDev)
		}
		mapDevice[string(rawByte)] = listDevicesLog
	}

	resp := u.Message(http.StatusOK, "get device Log")
	resp.Obj["deviceLogs"] = mapDevice
	return resp
}
