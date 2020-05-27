package crossEdit

import (
	"github.com/JanFant/newTLServer/internal/model/logger"
	"github.com/jmoiron/sqlx"
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
func FillingDeviceLogTable(db *sqlx.DB) {
	//если не кто не обращался за логами, заполняем таблицу раз в день
	timeTick := time.Tick(time.Hour * 24)
	_ = fillingTable(db)
	for {
		select {
		case fillStatus := <-FillingDeviceChan:
			{
				fillStatus.Status = fillingTable(db)
				FillingDeviceChan <- fillStatus
			}
		case <-timeTick:
			{
				_ = fillingTable(db)
			}
		}
	}
}

//fillingTable заполнение таблицы записями
func fillingTable(db *sqlx.DB) (status bool) {
	//запрос на уникальные не заполненные записи
	//sqlLogStr := fmt.Sprintf(`SELECT distinct id FROM %v where crossinfo is null`, GlobalConfig.DBConfig.LogDeviceTable)
	rowsDev, err := db.Query(`SELECT distinct id FROM public.logdevice WHERE crossinfo is null`)
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
		err = db.QueryRow(`SELECT region, area, id, describ FROM public.cross WHERE idevice = $1`, tempID).Scan(&TLight.Region, &TLight.Area, &TLight.ID, &TLight.Description)
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
		db.MustExec(`UPDATE public.logdevice SET  crossinfo = $1 WHERE id = $2 AND crossinfo IS NULL `)
	}
	return true
}
