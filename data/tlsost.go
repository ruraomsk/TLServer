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
	Idevice     int        `json:"idevice"`     //Реальный номер устройства
	Sost        TLSostInfo `json:"tlsost"`      //Состояние светофора
	Description string     `json:"description"` //Описание светофора
	Points      Point      `json:"points"`      //Координата где находится светофор
	State       State      `json:"state"`       //Полное состояние светофора полученное от устройства
}

type State struct {
	Ck         int    `json:"ck",sql:"ck"`
	Nk         int    `json:"nk",sql:"nk"`
	Pk         int    `json:"pk",sql:"pk"`
	Arrays     string `json:"arrays",sql:"Arrays"`
	Status     int    `json:"status",sql:"status"`
	Statistics string `json:"statistics",sql:"Statistics"`
}

//GetLightsFromBD возвращает массив в котором содержатся светофоры, которые попали в указанную область
func GetLightsFromBD(point0 Point, point1 Point) (tfdata []TrafficLights) {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)
	temp := &TrafficLights{}
	//tempState := &State{}
	sqlStr = fmt.Sprintf("select region, id, idevice, dgis, describ, state from %s ", os.Getenv("gis_table"))
	if !((point0.X == 0) && (point0.Y == 0) && (point1.X == 0) && (point1.Y == 0)) {
		sqlStr = sqlStr + fmt.Sprintf("where box '((%3.15f,%3.15f),(%3.15f,%3.15f))'@> dgis", point0.Y, point0.X, point1.Y, point1.X)
	}
	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		err := rowsTL.Scan(&temp.Region.Num, &temp.ID, &temp.Idevice, &dgis, &temp.Description, &StateStr)
		if err != nil {
			logger.Info.Println("tlsost: Что-то не так с запросом", err)
			return nil
		}
		temp.Points.StrToFloat(dgis)
		temp.Region.Name = CacheInfo.mapRegion[temp.Region.Num]

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

//UpdateTLightInfo обновить данные о светофорах вощедших в область
func UpdateTLightInfo(box BoxPoint) map[string]interface{} {
	resp := u.Message(true, "Update box data")
	//кривая работа с 180 меридианом приходится его обрезать (postgreqsl не может )
	if box.Point1.X > -180 && box.Point1.X < -160 {
		box.Point1.X = 179.999999999999999
	}
	tflight := GetLightsFromBD(box.Point0, box.Point1)
	resp["tflight"] = tflight
	return resp
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
	sqlStr = fmt.Sprintf("select idevice, dgis, describ, state from %s where region = %d and id = %d", os.Getenv("gis_table"), TLignt.Region.Num, TLignt.ID)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&TLignt.Idevice, &dgis, &TLignt.Description, &StateStr)
	if err != nil {
		resp := u.Message(false, "No result at these points")
		return resp
	}
	TLignt.Points.StrToFloat(dgis)
	TLignt.Region.Name = CacheInfo.mapRegion[TLignt.Region.Num]

	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(StateStr)
	if err != nil {
		logger.Info.Println("tlsost: Не удалось разобрать информацию о перекрестке", err)
	}
	TLignt.State = rState
	TLignt.Sost.Num = rState.Status
	TLignt.Sost.Description = CacheInfo.mapTLSost[TLignt.Sost.Num]
	resp := u.Message(true, "Cross information")
	resp["cross"] = TLignt
	return resp
}
