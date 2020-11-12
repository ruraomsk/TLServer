package exchangeData

import (
	"github.com/JanFant/TLServer/internal/model/device"
	"github.com/JanFant/TLServer/internal/sockets/techArm"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var Automatic = "AutomaticD"

type controllersInfo struct {
	Idevice      int       `json:"id"`           //idevice устройства
	Ip           string    `json:"ip"`           //ip gprs
	Port         int       `json:"port"`         //port gprs
	Name         string    `json:"name"`         //название перекрестка (совпадает с address)
	Address      string    `json:"address"`      //название перекрестка (дубликат)
	Active       bool      `json:"active"`       //устройство активно
	Created      time.Time `json:"created"`      //когда создан (нет данных)
	Updated      time.Time `json:"updated"`      //когда изменен (нет данных)
	GraphId      int       `json:"graph_id"`     //не знаю что такое
	Point        point     `json:"location"`     //позиция устройства
	Manufacturer string    `json:"manufacturer"` //производитель (автоматика)
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
		tempController.Name = tempController.Address
		tempController.Ip = techArm.GPRSInfo.IP
		tempController.Port, _ = strconv.Atoi(techArm.GPRSInfo.Port)
		tempController.Manufacturer = Automatic

		tempController.Created = time.Now()
		tempController.Updated = time.Now()
		//граф не знаю чем заполнить
		tempController.GraphId = 0

		//проверка включен ли светофор
		tempController.Active = false
		for key := range mapActivDevices {
			if tempController.Idevice == key {
				tempController.Active = true
			}
		}

		controllerList = append(controllerList, tempController)
	}

	//хотят чтобы было уложено так через data/list (что поделать)
	globalData := make(map[string]interface{})
	globalData["data"] = controllerList
	resp := u.Response{Code: http.StatusOK, Obj: map[string]interface{}{}}
	resp.Obj["list"] = globalData
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
