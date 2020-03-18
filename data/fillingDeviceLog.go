package data

import (
	"encoding/json"
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

//tlInfo информация для записи
type tlInfo struct {
	Region      int    `json:"region"`
	Area        int    `json:"area"`
	ID          int    `json:"id"`
	Description string `json:"description"`
	structStr   string
}

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

//toStr конвертировать в строку
func (tlInfo *tlInfo) toStr() (str string, err error) {
	newByte, err := json.Marshal(tlInfo)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}

func (tlInfo *tlInfo) toStruct(str string) (err error) {
	err = json.Unmarshal([]byte(str), tlInfo)
	if err != nil {
		return err
	}
	return nil
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
		var TLight tlInfo
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
