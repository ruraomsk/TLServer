package data

import (
	"encoding/json"
	"github.com/ruraomsk/TLServer/internal/model/config"
	"io/ioutil"
	"os"
	"strconv"
)

//CrossesJSON структура в json
type CrossesJSON struct {
	Regions []RegionCross `json:"regions"`
}

//RegionCross описание региона в json
type RegionCross struct {
	Region string      `json:"region"` //текстовое название региона
	Id     int         `json:"id"`     //числовое значение региона
	Areas  []AreaCross `json:"areas"`
}

//AreaCross описание района в json
type AreaCross struct {
	Area         string        `json:"area"` //текстовое название района
	Id           int           `json:"id"`   //числовое значение района
	Descriptions []Description `json:"descriptions"`
}

//Description описания перекрестка в json
type Description struct {
	Description string `json:"description"` //текстовое описание перекрестка
	Id          int    `json:"id"`          //номер перекрестка в бд
}

//crossInfo необходимая информация для запроса из бд
type crossInfo struct {
	region   string
	area     string
	id       int
	describe string
}

//CreateCrossesJSON создание json файла с содержанием всех перекрестков с их местом положения и описания
func CreateCrossesJSON() {
	db := GetDB("CreateCrossesJSON")
	defer FreeDB("CreateCrossesJSON")

	cross := new(CrossesJSON)
	var tempCrosses []crossInfo

	rows, _ := db.Query(`SELECT region, area, id, describ FROM public.cross`)
	for rows.Next() {
		var (
			tempCross crossInfo
		)
		_ = rows.Scan(&tempCross.region, &tempCross.area, &tempCross.id, &tempCross.describe)
		tempCrosses = append(tempCrosses, tempCross)
	}

	CacheInfo.Mux.Lock()
	for idReg, nameReg := range CacheInfo.MapRegion {
		if idReg == "*" {
			continue
		}
		intIdR, _ := strconv.Atoi(idReg)
		rc := RegionCross{Id: intIdR, Region: nameReg}
		for idArea, nameArea := range CacheInfo.MapArea[nameReg] {
			intIdA, _ := strconv.Atoi(idArea)
			ac := AreaCross{Id: intIdA, Area: nameArea, Descriptions: make([]Description, 0)}
			for _, tCross := range tempCrosses {
				if tCross.region == idReg && tCross.area == idArea {
					dc := Description{Id: tCross.id, Description: tCross.describe}
					ac.Descriptions = append(ac.Descriptions, dc)
				}
			}
			rc.Areas = append(rc.Areas, ac)
		}
		cross.Regions = append(cross.Regions, rc)
	}
	CacheInfo.Mux.Unlock()

	file, _ := json.MarshalIndent(cross, "", " ")
	filePath := config.GlobalConfig.StaticPath + "/cross/crossesInfo.json"
	_ = ioutil.WriteFile(filePath, file, os.ModePerm)
}
