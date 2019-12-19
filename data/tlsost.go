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
	ID          string     `json:"ID"`          //Уникальный ID светофора
	Region      RegionInfo `json:"region"`      //Регион
	Idevice     int        `json:"idevice"`     //реальный номер устройства
	Sost        TLSostInfo `json:"tlsost"`      //состояние светофора
	Description string     `json:"description"` //Описание светофора
	Points      Point      `json:"points"`      //Координата где находится светофор
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
		sqlStr = sqlStr + fmt.Sprintf("where box '((%3.15f,%3.15f),(%3.15f,%3.15f))'@> dgis", point0.X, point0.Y, point1.X, point1.Y)
	}
	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		_ = rowsTL.Scan(&temp.Region.Num, &temp.ID, &temp.Idevice, &dgis, &temp.Description, &StateStr)
		temp.Points.StrToFloat(dgis)
		temp.Region.Name = CacheInfo.mapRegion[temp.Region.Num]

		//Состояние светофора!
		rState, err := ConvertStateStrToStruct(StateStr)
		if err != nil {
			logger.Info.Println("Не удалось разобрать информацию о перекрестке", err)
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

func GetCrossInfo(TLignt TrafficLights)map[string]interface{}{
	//TLignt.
	return nil
}