package exchangeData

import (
	"github.com/JanFant/TLServer/internal/model/device"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"
	"strings"
)

type controllersInfo struct {
	Idevice int    `json:"id"`       //idevice устройства
	Address string `json:"address"`  //название перекрестка
	Active  bool   `json:"active"`   //устройство активно
	Point   point  `json:"location"` //позиция устройства
}

type point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func GetController(db *sqlx.DB) u.Response {
	var (
		controllerList  = make([]controllersInfo, 0)
		mapActivDevices = make(map[int]device.DevInfo)
	)

	rows, err := db.Query(`SELECT idevice, describ, dgis FROM public.cross`)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "db server not response")
	}

	device.GlobalDevices.Mux.Lock()
	mapActivDevices = device.GlobalDevices.MapDevices
	device.GlobalDevices.Mux.Unlock()

	for rows.Next() {
		var (
			tempController controllersInfo
			dgis           string
		)
		err = rows.Scan(&tempController.Idevice, &tempController.Address, &dgis)
		if err != nil {
			return u.Message(http.StatusInternalServerError, "error convert cross info")
		}
		//заполняем оставшиеся поля
		tempController.Point.StrToFloat(dgis)

		//проверка включен ли светофор
		tempController.Active = false
		for key := range mapActivDevices {
			if tempController.Idevice == key {
				tempController.Active = true
			}
		}

		controllerList = append(controllerList, tempController)
	}

	resp := u.Response{Code: http.StatusOK, Obj: map[string]interface{}{}}
	resp.Obj["data"] = controllerList
	return resp
}

//StrToFloat преобразует строку, полученную из БД в структуру Point
func (points *point) StrToFloat(str string) {
	str = strings.TrimPrefix(str, "(")
	str = strings.TrimSuffix(str, ")")
	temp := strings.Split(str, ",")
	if len(temp) != 2 {
		points.Y, points.X = 0, 0
		return
	}
	for num, part := range temp {
		temp[num] = strings.TrimSpace(part)
	}
	points.Y, _ = strconv.ParseFloat(temp[0], 64)
	points.X, _ = strconv.ParseFloat(temp[1], 64)
}
