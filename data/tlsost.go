package data

import (
	"fmt"
	"os"
)

type TrafficLights struct {
	ID          string `json:"ID"`          //Уникальный ID светофора
	Regin       string `json:"regin"`       //Регион
	Description string `json:"description"` //Описание светофора
	Points      Point  `json:"points"`      //Координата где находится светофор
}

//GetLightsFromBD возвращает массив в котором содержатся светофоры, которые попали в указанную область
func GetLightsFromBD(point0 Point, point1 Point) (tfdata []TrafficLights) {
	var dgis string
	temp := &TrafficLights{}
	sqlquery := fmt.Sprintf("select region, id, dgis, describ from %s where box '((%3.15f,%3.15f),(%3.15f,%3.15f))'@> dgis", os.Getenv("gis_table"), point0.X, point0.Y, point1.X, point1.Y)
	rows, _ := GetDB().Raw(sqlquery).Rows()
	for rows.Next() {
		rows.Scan(&temp.Regin, &temp.ID, &dgis, &temp.Description)
		temp.Points.StrToFloat(dgis)
		tfdata = append(tfdata, *temp)
	}
	return
}
