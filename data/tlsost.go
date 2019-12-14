package data

import (
	u "../utils"
	"fmt"
	"os"
)

//TrafficLights информация о светофоре
type TrafficLights struct {
	ID          string     `json:"ID"`          //Уникальный ID светофора
	Region      RegionInfo `json:"region"`      //Регион
	Idevice     int        `json:"idevice"`     //реальный номер устройства
	Description string     `json:"description"` //Описание светофора
	Points      Point      `json:"points"`      //Координата где находится светофор
}

//GetLightsFromBD возвращает массив в котором содержатся светофоры, которые попали в указанную область
func GetLightsFromBD(point0 Point, point1 Point) (tfdata []TrafficLights) {
	var (
		dgis   string
		sqlStr string
	)
	temp := &TrafficLights{}
	sqlStr = fmt.Sprintf("select region, id, idevice, dgis, describ from %s ", os.Getenv("gis_table"))
	if !((point0.X == 0) && (point0.Y == 0) && (point1.X == 0) && (point1.Y == 0)) {
		sqlStr = sqlStr + fmt.Sprintf("where box '((%3.15f,%3.15f),(%3.15f,%3.15f))'@> dgis", point0.X, point0.Y, point1.X, point1.Y)
	}
	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		_ = rowsTL.Scan(&temp.Region.Num, &temp.ID, &temp.Idevice, &dgis, &temp.Description)
		temp.Points.StrToFloat(dgis)
		temp.Region.Name =  CacheInfo.Region[temp.Region.Num]
		tfdata = append(tfdata, *temp)
	}

	return tfdata
}

//UpdateTLightInfo обновить данные о светофорах вощедших в область
func UpdateTLightInfo(box BoxPoint) map[string]interface{} {
	resp := u.Message(true, "Update box data")
	tflight := GetLightsFromBD(box.Point0, box.Point1)
	resp["tflight"] = tflight
	return resp
}
