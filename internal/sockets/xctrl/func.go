package xctrl

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/xcontrol"
	"sort"
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

//changeXctrl запись массива state в базу
func changeXctrl(states []xcontrol.State, db *sqlx.DB) error {
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

//createXctrl создание state в базу
func createXctrl(state xcontrol.State, db *sqlx.DB) error {
	strState, _ := json.Marshal(state)
	_, err := db.Exec(`INSERT INTO public.xctrl (region, area, subarea, state) VALUES ($1, $2, $3, $4)`, state.Region, state.Area, state.SubArea, string(strState))
	if err != nil {
		return err
	}
	return nil
}

//deleteXctrl удаление state из базы
func deleteXctrl(reg, area, sub int, db *sqlx.DB) error {
	_, err := db.Exec(`DELETE FROM public.xctrl WHERE region = $1 AND area = $2 AND subarea= $3`, reg, area, sub)
	if err != nil {
		return err
	}
	return nil
}

//getAreaTF генерация сортированного массива светофоров из базы
func getAreaTF(region, area int, db *sqlx.DB) (tfdata []data.TrafficLights, err error) {
	rowsTL, err := db.Query(`SELECT region, area, subarea, id, describ FROM public.cross WHERE region = $1 AND area = $2`, region, area)
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
	sort.Slice(tfdata, func(i, j int) bool {
		return tfdata[i].ID < tfdata[j].ID
	})
	return tfdata, nil
}
