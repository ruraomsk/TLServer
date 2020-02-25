package data

import (
	"../logger"
	u "../utils"
	"encoding/json"
	"errors"
	"fmt"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"os"
)

//TrafficLights информация о светофоре
type TrafficLights struct {
	ID          int        `json:"ID"`          //Уникальный ID светофора
	Region      RegionInfo `json:"region"`      //Регион
	Area        AreaInfo   `json:"area"`        //Район
	Subarea     int        `json:"subarea"`     //ПодРайон
	Idevice     int        `json:"idevice"`     //Реальный номер устройства
	Sost        TLSostInfo `json:"tlsost"`      //Состояние светофора
	Description string     `json:"description"` //Описание светофора
	Points      Point      `json:"points"`      //Координата где находится светофор
}

//Locations информация о запрашиваемом регионе и районе карты
type Locations struct {
	Region string   `json:"region"` //регион
	Area   []string `json:"area"`   //районы
}

//GetLightsFromBD определяем область отображения светофоров
func GetLightsFromBD(box BoxPoint) (tfdata []TrafficLights) {
	var tflight = []TrafficLights{}
	if (box.Point1.X > -180 && box.Point1.X < 0) && (box.Point0.X > 0 && box.Point0.X < 180) {
		var (
			point0 Point
			point1 Point
		)
		//для первую область
		point0 = box.Point0
		point1.Y = box.Point1.Y
		point1.X = 179.9999999999
		tflight = SelectTL(point0, point1, false)
		//для второй области
		point0.Y = box.Point0.Y
		point0.X = -179.9999999999
		point1 = box.Point1
		tempTF := SelectTL(point0, point1, false)
		tflight = append(tflight, tempTF...)

	} else if int(box.Point0.X) == int(box.Point1.X) {
		tflight = SelectTL(box.Point0, box.Point1, true)
	} else {
		tflight = SelectTL(box.Point0, box.Point1, false)
	}
	return tflight
}

//SelectTL возвращает массив в котором содержатся светофоры, которые попали в указанную область
func SelectTL(point0 Point, point1 Point, equalPoint bool) (tfdata []TrafficLights) {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)

	temp := &TrafficLights{}
	if equalPoint {
		sqlStr = fmt.Sprintf("select region, area, subarea, id, idevice, dgis, describ, state from %s", os.Getenv("gis_table"))
	} else {
		sqlStr = fmt.Sprintf("select region, area, subarea, id, idevice, dgis, describ, state from %s where box '((%3.15f,%3.15f),(%3.15f,%3.15f))'@> dgis", os.Getenv("gis_table"), point0.Y, point0.X, point1.Y, point1.X)
	}
	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		err := rowsTL.Scan(&temp.Region.Num, &temp.Area.Num, &temp.Subarea, &temp.ID, &temp.Idevice, &dgis, &temp.Description, &StateStr)
		if err != nil {
			logger.Error.Println("|Message: No result at these points", err.Error())
			return nil
		}
		temp.Points.StrToFloat(dgis)
		//Состояние светофора!
		rState, err := ConvertStateStrToStruct(StateStr)
		if err != nil {
			logger.Error.Println("|Message: Failed to parse cross information", err.Error())
			return nil
		}
		CacheInfo.mux.Lock()
		temp.Region.NameRegion = CacheInfo.mapRegion[temp.Region.Num]
		temp.Area.NameArea = CacheInfo.mapArea[temp.Region.NameRegion][temp.Area.Num]
		temp.Sost.Description = CacheInfo.mapTLSost[temp.Sost.Num]
		CacheInfo.mux.Unlock()
		temp.Sost.Num = rState.StatusDevice
		tfdata = append(tfdata, *temp)
	}

	return tfdata
}

//GetAllTrafficLights запрос информации об всех сфетофорах из БД
func GetAllTrafficLights() (tfData []TrafficLights) {
	var dgis string
	temp := &TrafficLights{}
	sqlquery := fmt.Sprintf("select region, id, area, dgis, describ from %s", os.Getenv("gis_table"))
	rows, _ := GetDB().Raw(sqlquery).Rows()
	for rows.Next() {
		rows.Scan(&temp.Region.Num, &temp.ID, &temp.Area.Num, &dgis, &temp.Description)
		temp.Points.StrToFloat(dgis)
		tfData = append(tfData, *temp)
	}
	return
}

//ConvertStateStrToStruct разбор данных полученных из БД в нужную структуру
func ConvertStateStrToStruct(str string) (rState agS_pudge.Cross, err error) {
	if err := json.Unmarshal([]byte(str), &rState); err != nil {
		return rState, err
	}
	return rState, nil
}

//GetCrossInfo сбор информации для пользователя и выбранном перекрестке
func GetCrossInfo(TLignt TrafficLights) map[string]interface{} {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)
	sqlStr = fmt.Sprintf("select area, subarea, idevice, dgis, describ, state from %v where region = %v and id = %v and area = %v", os.Getenv("gis_table"), TLignt.Region.Num, TLignt.ID, TLignt.Area.Num)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &StateStr)
	if err != nil {
		logger.Error.Println("|Message: No result at these points", err.Error())
		return u.Message(false, "No result at these points")
	}
	TLignt.Points.StrToFloat(dgis)
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(StateStr)
	if err != nil {
		logger.Error.Println("|Message: Failed to parse cross information", err.Error())
		return u.Message(false, "Failed to parse cross information")
	}
	CacheInfo.mux.Lock()
	TLignt.Region.NameRegion = CacheInfo.mapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.mapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	TLignt.Sost.Num = rState.StatusDevice
	TLignt.Sost.Description = CacheInfo.mapTLSost[TLignt.Sost.Num]
	CacheInfo.mux.Unlock()
	resp := u.Message(true, "Cross information")
	resp["DontWrite"] = "true"
	resp["cross"] = TLignt
	resp["state"] = rState
	return resp
}

//MakeBoxPoint расчет координат для перемешения по карте
func (location *Locations) MakeBoxPoint() (box BoxPoint, err error) {
	var sqlStr = `SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross"`
	tempStr := " where "
	tempStr += fmt.Sprintf("region = %v and area in (", location.Region)
	for numArea, area := range location.Area {
		if numArea == 0 {
			tempStr += fmt.Sprintf("%v", area)
		} else {
			tempStr += fmt.Sprintf(",%v", area)
		}
	}
	tempStr += ")"
	sqlStr += tempStr
	row := GetDB().Raw(sqlStr).Row()
	err = row.Scan(&box.Point0.Y, &box.Point0.X, &box.Point1.Y, &box.Point1.X)
	if err != nil {
		return box, errors.New(fmt.Sprintf("ParserPoints. Request error: %s", err.Error()))
	}
	if box.Point0.X > 180 {
		box.Point0.X -= 360
	}
	if box.Point1.X > 180 {
		box.Point1.X -= 360
	}
	return
}
