package data

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"os"
	"strconv"
)

type CrossesJSON struct {
	Regions []RegionCross `json:"regions"`
}

type RegionCross struct {
	Region string      `json:"region"`
	Id     int         `json:"id"`
	Areas  []AreaCross `json:"areas"`
}

type AreaCross struct {
	Area         string        `json:"area"`
	Id           int           `json:"id"`
	Descriptions []Description `json:"descriptions"`
}

type Description struct {
	Description string `json:"description"`
	Id          int    `json:"id"`
}

type crossInfo struct {
	region   string
	area     string
	id       int
	describe string
}

//CreateCrossesJSON создание json файла с содержанием всех перекрестков с их местом положения и описания
func CreateCrossesJSON(db *sqlx.DB) {
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
