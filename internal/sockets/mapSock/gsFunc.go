package mapSock

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/routeGS"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
)

//getAllModes вернуть из базы все маршруты
func getAllModes(db *sqlx.DB) interface{} {
	var (
		modes = make([]routeGS.Route, 0)
	)
	rows, err := db.Query(`SELECT id, description, box, listtl, region FROM public.routes`)
	if err != nil {
		logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
		return modes
	}
	for rows.Next() {
		var (
			temp            routeGS.Route
			listSrt, boxStr string
		)
		err := rows.Scan(&temp.Id, &temp.Description, &boxStr, &listSrt, &temp.Region)
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
		modes = append(modes, temp)
	}
	return modes
}
