package exchangeData

import (
	"github.com/JanFant/TLServer/internal/model/device"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/jmoiron/sqlx"
	"net/http"
	"time"
)

type plansInfo struct {
	Idevice      int         `json:"id"`            //id устройства
	Name         string      `json:"name"`          //нет информации
	Type         string      `json:"type"`          //нет информации
	Active       bool        `json:"active"`        //устройство активно
	Created      time.Time   `json:"created"`       //+-
	Updated      time.Time   `json:"updated"`       //+-
	GroupId      int         `json:"group_id"`      //нет информации
	ControllerId int         `json:"controller_id"` //нет информации
	ProgGroups   []progGroup `json:"programs_groups"`
}

type progGroup struct {
	Days     string    `json:"days"` //нет информации
	Programs []program `json:"programs"`
}

type program struct {
	Id             int       `json:"id"`    //нет информации
	Mode           string    `json:"mode"`  //нет информации
	Shift          int       `json:"shift"` //нет информации
	Phases         []phase   `json:"phases"`
	PlanId         int       `json:"plan_id"` //нет информации
	Conditions     condition `json:"conditions"`
	CircleTime     int       `json:"circle_time"`     //нет информации
	ControllerId   int       `json:"controller_id"`   //нет информации
	ConditionsType string    `json:"conditions_type"` //нет информации
}

type phase struct {
	Max   int `json:"max"`         //нет информации
	Min   int `json:"min"`         //нет информации
	Id    int `json:"phase_id"`    //нет информации
	Time  int `json:"phase_time"`  //нет информации
	Order int `json:"phase_order"` //нет информации
}

type condition struct {
	Days      int    `json:"days"`       //нет информации
	EndTime   string `json:"end_time"`   //нет информации
	StartTime string `json:"start_time"` //нет информации
}

func GetPlans(iDevice []int, db *sqlx.DB) u.Response {
	var (
		PlansList       = make([]plansInfo, 0)
		mapActivDevices = make(map[int]device.DevInfo)
	)

	device.GlobalDevices.Mux.Lock()
	mapActivDevices = device.GlobalDevices.MapDevices
	device.GlobalDevices.Mux.Unlock()

	query, args, err := sqlx.In(`SELECT state FROM public.cross WHERE idevice IN (?)`, iDevice)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "error formatting IN query")
	}
	query = db.Rebind(query)
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "db server not response")
	}

	for rows.Next() {
		var (
			tempPlans plansInfo
			stateStr  string
		)
		tempPlans.ProgGroups = make([]progGroup, 1)

		err = rows.Scan(&stateStr)
		if err != nil {
			return u.Message(http.StatusInternalServerError, "error convert cross info")
		}

		state, _ := crossSock.ConvertStateStrToStruct(stateStr)
		//заполнение

		tempPlans.Idevice = state.IDevice
		//проверка включено ли устройство
		tempPlans.Active = false
		for key := range mapActivDevices {
			if tempPlans.Idevice == key {
				tempPlans.Active = true
			}
		}
		tempPlans.Created = time.Now()
		tempPlans.Updated = time.Now()

		PlansList = append(PlansList, tempPlans)
	}

	//хотят чтобы было уложено так через data/list (что поделать)
	globalData := make(map[string]interface{})
	globalData["data"] = PlansList
	resp := u.Response{Code: http.StatusOK, Obj: map[string]interface{}{}}
	resp.Obj["list"] = globalData
	return resp
}
