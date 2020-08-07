package maps

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/binding"
)

//SelectTL возвращает массив в котором содержатся светофоры, которые попали в указанную область
func SelectTL(db *sqlx.DB) (tfdata []data.TrafficLights) {
	var dgis string
	rowsTL, err := db.Query(`SELECT region, area, subarea, id, idevice, dgis, describ, status, state->'arrays'->'SetDK' FROM public.cross`)
	if err != nil {
		logger.Error.Println("|Message: db not respond", err.Error())
		return nil
	}
	for rowsTL.Next() {
		var (
			temp      = data.TrafficLights{}
			tempSetDK binding.SetDK
			dkStr     string
		)
		err := rowsTL.Scan(&temp.Region.Num, &temp.Area.Num, &temp.Subarea, &temp.ID, &temp.Idevice, &dgis, &temp.Description, &temp.Sost.Num, &dkStr)
		if err != nil {
			logger.Error.Println("|Message: No result at these points", err.Error())
			return nil
		}
		_ = json.Unmarshal([]byte(dkStr), &tempSetDK)
		temp.Phases = tempSetDK.GetPhases()
		temp.Points.StrToFloat(dgis)
		data.CacheInfo.Mux.Lock()
		temp.Region.NameRegion = data.CacheInfo.MapRegion[temp.Region.Num]
		temp.Area.NameArea = data.CacheInfo.MapArea[temp.Region.NameRegion][temp.Area.Num]
		temp.Sost.Description = data.CacheInfo.MapTLSost[temp.Sost.Num]
		data.CacheInfo.Mux.Unlock()
		tfdata = append(tfdata, temp)
	}
	return tfdata
}

//MapOpenInfo сбор всех данных для отображения информации на карте
func MapOpenInfo(db *sqlx.DB) (obj map[string]interface{}) {
	obj = make(map[string]interface{})

	location := &data.Locations{}
	box, _ := location.MakeBoxPoint()
	obj["boxPoint"] = &box
	obj["tflight"] = SelectTL(db)
	obj["authorizedFlag"] = false

	//собираю в кучу регионы для отображения
	chosenRegion := make(map[string]string)
	data.CacheInfo.Mux.Lock()
	for first, second := range data.CacheInfo.MapRegion {
		chosenRegion[first] = second
	}
	delete(chosenRegion, "*")
	obj["regionInfo"] = chosenRegion

	//собираю в кучу районы для отображения
	chosenArea := make(map[string]map[string]string)
	for first, second := range data.CacheInfo.MapArea {
		chosenArea[first] = make(map[string]string)
		chosenArea[first] = second
	}
	delete(chosenArea, "Все регионы")
	data.CacheInfo.Mux.Unlock()
	obj["areaInfo"] = chosenArea
	return
}
