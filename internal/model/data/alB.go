package data

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/locations"
)

var CacheALB []locations.Point

func FillCacheALB() {
	var TempALB locations.Points
	//запрос уникальных регионов и районов
	rows, _ := GetDB().Query(`SELECT  region, area, dgis FROM public."cross" where region = 1 and area = 1`)
	for rows.Next() {
		var (
			temp = struct {
				Region string          `json:"region"`
				Area   string          `json:"area"`
				Point  locations.Point `json:"point"`
			}{}
			tempPointStr string
			tempPoint    locations.Point
		)
		_ = rows.Scan(&temp.Region, &temp.Area, &tempPointStr)
		tempPoint.StrToFloat(tempPointStr)
		temp.Point.X = tempPoint.X
		temp.Point.Y = tempPoint.Y

		TempALB = append(TempALB, temp.Point)
	}
	CacheALB = TempALB.ConvexHull()
	fmt.Println(CacheALB)
}
