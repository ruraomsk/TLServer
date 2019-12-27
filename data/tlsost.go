package data

import (
	"../logger"
	u "../utils"
	"encoding/json"
	"fmt"
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
	State       State      `json:"state"`       //Полное состояние светофора полученное от устройства
}

type State struct {
	Ck         int   `json:"ck",sql:"ck"`
	Nk         int   `json:"nk",sql:"nk"`
	Pk         int   `json:"pk",sql:"pk"`
	Arrays     []int `json:"arrays",sql:"Arrays"`
	Status     int   `json:"status",sql:"status"`
	Statistics []int `json:"statistics",sql:"Statistics"`
}

//GetLightsFromBD возвращает массив в котором содержатся светофоры, которые попали в указанную область
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
		tflight = SelectTL(point0, point1)
		//для второй области
		point0.Y = box.Point0.Y
		point0.X = -179.9999999999
		point1 = box.Point1
		tempTF := SelectTL(point0, point1)
		tflight = append(tflight, tempTF...)

	} else {
		tflight = SelectTL(box.Point0, box.Point1)
	}
	return tflight
}

func SelectTL(point0 Point, point1 Point) (tfdata []TrafficLights) {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)
	temp := &TrafficLights{}
	//tempState := &State{}
	sqlStr = fmt.Sprintf("select region, area, subarea, id, idevice, dgis, describ, state from %s where box '((%3.15f,%3.15f),(%3.15f,%3.15f))'@> dgis", os.Getenv("gis_table"), point0.Y, point0.X, point1.Y, point1.X)
	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		err := rowsTL.Scan(&temp.Region.Num, &temp.Area.Num, &temp.Subarea, &temp.ID, &temp.Idevice, &dgis, &temp.Description, &StateStr)
		if err != nil {
			logger.Info.Println("tlsost: Что-то не так с запросом", err)
			return nil
		}
		temp.Points.StrToFloat(dgis)
		temp.Region.NameRegion = CacheInfo.mapRegion[temp.Region.Num]
		temp.Area.NameArea = CacheInfo.mapArea[temp.Region.NameRegion][temp.Area.Num]
		//Состояние светофора!
		rState, err := ConvertStateStrToStruct(StateStr)
		if err != nil {
			logger.Info.Println("tlsost: Не удалось разобрать информацию о перекрестке", err)
		}
		temp.Sost.Num = rState.Status
		temp.Sost.Description = CacheInfo.mapTLSost[temp.Sost.Num]
		tfdata = append(tfdata, *temp)
	}

	return tfdata
}

func ConvertStateStrToStruct(str string) (rState State, err error) {
	if err := json.Unmarshal([]byte(str), &rState); err != nil {
		return rState, err
	}
	return rState, nil
}

func GetCrossInfo(TLignt TrafficLights) map[string]interface{} {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)
	sqlStr = fmt.Sprintf("select area, subarea, idevice, dgis, describ, state from %s where region = %d and id = %d", os.Getenv("gis_table"), TLignt.Region.Num, TLignt.ID)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &StateStr)
	if err != nil {
		logger.Info.Println("getCrossInfo: Что-то не так с запросом", err)
		return u.Message(false, "No result at these points")
	}
	TLignt.Points.StrToFloat(dgis)
	TLignt.Region.NameRegion = CacheInfo.mapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.mapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(StateStr)
	if err != nil {
		logger.Info.Println("getCrossInfo: Не удалось разобрать информацию о перекрестке", err)
	}
	TLignt.State = rState
	TLignt.Sost.Num = rState.Status
	TLignt.Sost.Description = CacheInfo.mapTLSost[TLignt.Sost.Num]
	resp := u.Message(true, "Cross information")
	resp["cross"] = TLignt
	return resp
}
