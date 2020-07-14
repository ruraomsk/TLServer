package data

import (
	"errors"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/locations"
	"github.com/JanFant/TLServer/logger"
)

//TrafficLights информация о светофоре
type TrafficLights struct {
	ID          int             `json:"ID"`          //Уникальный ID светофора
	Region      RegionInfo      `json:"region"`      //Регион
	Area        AreaInfo        `json:"area"`        //Район
	Subarea     int             `json:"subarea"`     //ПодРайон
	Idevice     int             `json:"idevice"`     //Реальный номер устройства
	Sost        TLSostInfo      `json:"tlsost"`      //Состояние светофора
	Description string          `json:"description"` //Описание светофора
	Phases      []int           `json:"phases"`      //Доступные фазы
	Points      locations.Point `json:"points"`      //Координата где находится светофор
}

//Locations информация о запрашиваемом регионе и районе карты
type Locations struct {
	Region string   `json:"region"` //регион
	Area   []string `json:"area"`   //районы
}

//GetAllTrafficLights запрос информации об всех сфетофорах из БД
func GetAllTrafficLights() (tfData []TrafficLights) {
	var dgis string
	temp := &TrafficLights{}
	sqlStr := fmt.Sprintf("SELECT region, id, area, dgis, describ FROM public.cross")
	rows, err := GetDB().Query(sqlStr)
	if err != nil {
		logger.Error.Println("|Message: db not respond", err.Error())
		return nil
	}
	for rows.Next() {
		_ = rows.Scan(&temp.Region.Num, &temp.ID, &temp.Area.Num, &dgis, &temp.Description)
		temp.Points.StrToFloat(dgis)
		tfData = append(tfData, *temp)
	}
	return
}

//MakeBoxPoint расчет координат для перемещения по карте
func (l *Locations) MakeBoxPoint() (box locations.BoxPoint, err error) {
	var sqlStr = `SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross"`
	if l.Region != "" {
		tempStr := " WHERE "
		tempStr += fmt.Sprintf("region = %v AND area in (", l.Region)
		for numArea, area := range l.Area {
			if numArea == 0 {
				tempStr += fmt.Sprintf("%v", area)
			} else {
				tempStr += fmt.Sprintf(",%v", area)
			}
		}
		tempStr += ")"
		sqlStr += tempStr
	}
	row := GetDB().QueryRow(sqlStr)
	err = row.Scan(&box.Point0.Y, &box.Point0.X, &box.Point1.Y, &box.Point1.X)
	if err != nil {
		return box, errors.New(fmt.Sprintf("parserPoints. Request error: %s", err.Error()))
	}
	if box.Point0.X > 180 {
		box.Point0.X -= 360
	}
	if box.Point1.X > 180 {
		box.Point1.X -= 360
	}
	return
}
