package greenStreet

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/model/routeGS"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"strconv"
)

//executeRoute управление светофорами
type executeRoute struct {
	Devices []int `json:"devices"`
	TurnOn  bool  `json:"turnOn"`
}

func getPhases(devices []int, db *sqlx.DB) []*Phase {
	result := make([]*Phase, 0)
	for _, i := range devices {
		rows, err := db.Query(`SELECT device FROM public.devices where id=` + strconv.Itoa(i))
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return result
		}
		var s []byte
		var state pudge.Controller
		for rows.Next() {
			rows.Scan(&s)
			err = json.Unmarshal(s, &state)
			if err != nil {
				logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
				return result
			}
			result = append(result, &Phase{Device: i, Phase: state.DK.FDK})
			break
		}
	}
	//logger.Debug.Printf("devs %v  %v",devices,result)
	return result
}

//getAllModes вернуть из базы все маршруты
func getAllModes(db *sqlx.DB) interface{} {
	var (
		modes = make([]routeGS.Route, 0)
	)
	rows, err := db.Query(`SELECT description, box, listtl, region FROM public.routes`)
	if err != nil {
		logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
		return modes
	}
	for rows.Next() {
		var (
			temp            routeGS.Route
			listSrt, boxStr string
		)
		err := rows.Scan(&temp.Description, &boxStr, &listSrt, &temp.Region)
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return modes
		}
		err = json.Unmarshal([]byte(listSrt), &temp.List)
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return modes
		}
		err = json.Unmarshal([]byte(boxStr), &temp.Box)
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return modes
		}

		if len(temp.List) == 0 {
			temp.List = make([]routeGS.RouteTL, 0)
		}
		for numR, route := range temp.List {
			rowRoute := db.QueryRow(`SELECT describ FROM public.cross WHERE region = $1 AND area = $2 AND id = $3`, route.Pos.Region, route.Pos.Area, route.Pos.Id)
			_ = rowRoute.Scan(&temp.List[numR].Description)
		}
		modes = append(modes, temp)
	}
	return modes
}
