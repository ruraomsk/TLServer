package data

import (
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"time"
)

//FillingInfo информация о запросе
type FillingInfo struct {
	User   string //пользователь запросивший заполнение таблицы
	Status bool   //статус выполнения запроса
}

//FillingDeviceChan канал запроса на заполнение таблицы логов устройства
var FillingDeviceChan = make(chan FillingInfo)

//FillingDeviceLogTable заполнение таблины логов устройств
func FillingDeviceLogTable() {
	timeTick := time.Tick(time.Hour * 24)
	_ = fillingTable()
	for {
		select {
		case fillStatus := <-FillingDeviceChan:
			{
				fillStatus.Status = fillingTable()
				FillingDeviceChan <- fillStatus
			}
		case <-timeTick:
			{
				_ = fillingTable()
			}
		}
	}
}

//fillingTable заполнение таблицы записями
func fillingTable() (status bool) {
	//запрос на уникальные не заполненные записи
	sqlLogStr := fmt.Sprintf(`SELECT distinct id FROM %v where crossinfo is null`, GlobalConfig.DBConfig.LogDeviceTable)
	rowsDev, err := GetDB().Raw(sqlLogStr).Rows()
	if err != nil {
		logger.Error.Println("|Message: LogDevice table error: ", err.Error())
		return false
	}
	for rowsDev.Next() {
		var tempID int
		err := rowsDev.Scan(&tempID)
		if err != nil {
			logger.Error.Println("|Message: LogDevice scan ID from table: ", err.Error())
			return false
		}
		//запрос на информацию о девайсах которые не заполнены
		var TLight BusyArm
		sqlCrossStr := fmt.Sprintf(`SELECT region, area, id, describ FROM %v where idevice = %v`, GlobalConfig.DBConfig.CrossTable, tempID)
		err = GetDB().Raw(sqlCrossStr).Row().Scan(&TLight.Region, &TLight.Area, &TLight.ID, &TLight.Description)
		if err != nil {
			logger.Error.Println("|Message: LogDevice CrossTable error: ", err.Error())
			return false
		}
		TLight.structStr, err = TLight.toStr()
		if err != nil {
			logger.Error.Println("|Message: LogDevice convert json error: ", err.Error())
			return false
		}
		//запрос на заполнение информацией
		sqlLogUpdateStr := fmt.Sprintf(`UPDATE %v SET crossinfo= '%v' WHERE id = %v and crossinfo is null`, GlobalConfig.DBConfig.LogDeviceTable, TLight.structStr, tempID)
		GetDB().Exec(sqlLogUpdateStr)
	}
	return true
}
