package data

import "fmt"

type TrafficLights struct {
	ID          string `json:"ID"`
	Description string `json:"description"`
	Points      Point  `json:"points"`
}

func GetLightsFromBD(point0 Point, point1 Point) (tfdata []TrafficLights) {
	var dgis string
	temp := &TrafficLights{}
	sqlquery := fmt.Sprintf("select id, dgis, describ from dev_gis where box '((%3.15f,%3.15f),(%3.15f,%3.15f))'@> dgis", point0.X, point0.Y, point1.X, point1.Y)
	fmt.Println(sqlquery)
	rows, _ := GetDB().Raw(sqlquery).Rows()
	for rows.Next() {
		rows.Scan(&temp.ID, &dgis, &temp.Description)
		temp.Points.StrToFloat(dgis)
		tfdata = append(tfdata, *temp)
	}
	return
}
