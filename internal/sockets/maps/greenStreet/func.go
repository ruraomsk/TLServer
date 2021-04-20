package greenStreet

import (
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/routeGS"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"strconv"
	"sync"
)

//executeRoute управление светофорами
type executeRoute struct {
	Devices []int `json:"devices"`
	TurnOn  bool  `json:"turnOn"`
}

var mutex sync.Mutex

func getPhases(devices []int) []*Phase {
	mutex.Lock()
	db, id := data.GetDB()
	defer func() {
		data.FreeDB(id)
		mutex.Unlock()
	}()
	result := make([]*Phase, 0)
	//logger.Debug.Printf("devices %v ",devices)
	for _, i := range devices {
		rows, err := db.Query(`SELECT device FROM public.devices where id=` + strconv.Itoa(i))
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return result
		}
		//logger.Debug.Printf("getPhases after select ")
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
			//logger.Debug.Printf("getPhases append %d ",i)
		}
		rows.Close()
	}
	//logger.Debug.Printf("devs %v",result)
	return result
}

//getAllModes вернуть из базы все маршруты
func getAllModes() interface{} {
	db, id := data.GetDB()
	db1, id1 := data.GetDB()
	defer func() {
		data.FreeDB(id)
		data.FreeDB(id1)
	}()
	var (
		modes = make([]routeGS.Route, 0)
	)
	rows, err := db.Query(`SELECT description, box, listtl, region FROM public.routes`)
	if err != nil {
		logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
		return modes
	}
	//logger.Debug.Printf("getAllModes after select")
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
		//logger.Debug.Printf("getAllModes temp.List %v",temp.List)
		for numR, route := range temp.List {

			w := fmt.Sprintf("SELECT describ FROM public.cross WHERE region = %s AND area = %s AND id = %d;", route.Pos.Region, route.Pos.Area, route.Pos.Id)
			//logger.Debug.Print(w)
			rows1, err := db1.Query(w)
			if err != nil {
				logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
				return modes
			}

			for rows1.Next() {
				rows1.Scan(&temp.List[numR].Description)
			}
		}
		modes = append(modes, temp)
		//logger.Debug.Printf("getAllModes append modes")
	}
	//logger.Debug.Printf("getAllModes out")
	return modes
}
