package xctrl

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/xcontrol"
)

//getXctrl формирует массив записей из таблицы xctrl
func getXctrl(db *sqlx.DB) ([]xcontrol.State, error) {
	rows, err := db.Query(`SELECT state FROM public.xctrl`)
	if err != nil {
		return nil, err
	}
	var allXctrl []xcontrol.State
	for rows.Next() {
		var (
			strXctrl string
			temp     xcontrol.State
		)
		err := rows.Scan(&strXctrl)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(strXctrl), &temp)
		allXctrl = append(allXctrl, temp)
	}
	return allXctrl, nil
}

//writeXctrl запис массива state в базу
func writeXctrl(states []xcontrol.State, db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, state := range states {
		strState, _ := json.Marshal(state)
		_, err := tx.Exec(`UPDATE public.xctrl SET state = $1 WHERE region = $2 AND area = $3 AND subarea = $4`, string(strState), state.Region, state.Area, state.SubArea)
		if err != nil {
			err = tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func getSubAreaTF(region, area, sub int, db *sqlx.DB) (tfdata []data.TrafficLights, err error) {
	rowsTL, err := db.Query(`SELECT region, area, subarea, id, describ FROM public.cross WHERE region = $1 AND area = $2 AND subarea = $3`, region, area, sub)
	if err != nil {
		logger.Error.Println("|Message: db not respond", err.Error())
		return nil, err
	}
	for rowsTL.Next() {
		var temp = data.TrafficLights{}
		err := rowsTL.Scan(&temp.Region.Num, &temp.Area.Num, &temp.Subarea, &temp.ID, &temp.Description)
		if err != nil {
			logger.Error.Println("|Message: No result at these points", err.Error())
			return nil, err
		}
		data.CacheInfo.Mux.Lock()
		temp.Region.NameRegion = data.CacheInfo.MapRegion[temp.Region.Num]
		temp.Area.NameArea = data.CacheInfo.MapArea[temp.Region.NameRegion][temp.Area.Num]
		temp.Sost.Description = data.CacheInfo.MapTLSost[temp.Sost.Num].Description
		temp.Sost.Control = data.CacheInfo.MapTLSost[temp.Sost.Num].Control
		data.CacheInfo.Mux.Unlock()
		tfdata = append(tfdata, temp)
	}
	return tfdata, nil
}
