package device

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"sync"
	"time"
)

var (
	GlobalDevices Devices
	GlobalDevEdit EditDevices
	sleepReadTime = time.Second * 1
)

type Devices struct {
	Mux        sync.Mutex
	MapDevices map[int]DevInfo
}

//DevInfo информация об устройстве
type DevInfo struct {
	Region     int                 //регион устройства
	Area       int                 //район устройства
	Controller agspudge.Controller //унформация устройства
}

type EditDevices struct {
	Mux        sync.Mutex
	MapDevices map[int]ControllerEdit
}

//ControllerEdit информация о редактировании устройства
type ControllerEdit struct {
	BusyCount  int  //количество устройств которые редактируются
	TurnOnFlag bool //флаг контроля отправки команды на редактирование
}

//StartReadDevices чтение БД таблицы public.devices
func StartReadDevices(db *sqlx.DB) {
	GlobalDevices.MapDevices = make(map[int]DevInfo, 0)
	GlobalDevEdit.MapDevices = make(map[int]ControllerEdit, 0)
	for {
		var (
			tempDevice = make(map[int]DevInfo)
		)
		rows, err := db.Query(`SELECT c.region, c.area, d.device FROM public.cross as c, public.devices as d WHERE c.idevice IN(d.id)`)
		if err != nil {
			continue
		}
		for rows.Next() {
			var (
				tempDev  DevInfo
				rowContr []byte
			)
			_ = rows.Scan(&tempDev.Region, &tempDev.Area, &rowContr)
			_ = json.Unmarshal(rowContr, &tempDev.Controller)
			tempDevice[tempDev.Controller.ID] = tempDev
		}
		//обновление девайса
		GlobalDevices.Mux.Lock()
		GlobalDevices.MapDevices = make(map[int]DevInfo, 0)
		GlobalDevices.MapDevices = tempDevice
		GlobalDevices.Mux.Unlock()

		//заполнение мапы управления
		GlobalDevEdit.Mux.Lock()
		for key := range tempDevice {
			if _, ok := GlobalDevEdit.MapDevices[key]; !ok {
				var nullEdit = ControllerEdit{BusyCount: 0, TurnOnFlag: false}
				GlobalDevEdit.MapDevices[key] = nullEdit
			}
		}
		GlobalDevEdit.Mux.Unlock()

		time.Sleep(sleepReadTime)
	}
}
