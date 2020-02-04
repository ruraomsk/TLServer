package data

import (
	u "../utils"
	"fmt"
	"os"
	"strings"
)

//SelectedData
type SelectedData struct {
	SelectedData map[string]map[string][]CheckData `json:"selected"`
	PngSettings  PngSettings                       `json:"pngSettings"`
}

//CheckData структура проверки для перекрестков
type CheckData struct {
	ID        string `json:"ID"`
	PngStatus bool   `json:"pngStatus"`
	SvgStatus bool   `json:"svgStatus"`
}

//PngSettings настройки размеров создаваемой map.png
type PngSettings struct {
	SizeX string `json:"sizeX"`
	SizeY string `json:"sizeY"`
	Z     string `json:"z"`
}

func (checkData *CheckData) setStatusTrue() {
	checkData.SvgStatus = true
	checkData.PngStatus = true
}

func (checkData *CheckData) setStatusFalse() {
	checkData.SvgStatus = false
	checkData.PngStatus = false
}

//stockData заполняет поля из env файла
func (set *PngSettings) stockData() {
	set.SizeX = os.Getenv("png_sizeX")
	set.SizeY = os.Getenv("png_sizeY")
	set.Z = os.Getenv("png_Z")
}

//MainCrossCreator формираю необходимые данные для начальной странички с деревом
func MainCrossCreator() map[string]interface{} {
	//CacheInfoDataUpdate()
	tfData := GetAllTrafficLights()
	mapRegAreaCross := make(map[string]map[string][]TrafficLights)
	for numReg, nameReg := range CacheInfo.mapRegion {
		if strings.Contains(numReg, "*") {
			continue
		}
		for numArea, nameArea := range CacheInfo.mapArea[nameReg] {
			if strings.Contains(numArea, "Все регионы") {
				continue
			}
			if _, ok := mapRegAreaCross[nameReg]; !ok {
				mapRegAreaCross[nameReg] = make(map[string][]TrafficLights)
			}
			for _, tempTF := range tfData {
				if (tempTF.Region.Num == numReg) && (tempTF.Area.Num == numArea) {
					mapRegAreaCross[nameReg][nameArea] = append(mapRegAreaCross[nameReg][nameArea], tempTF)
				}
			}

		}
	}
	var settings PngSettings
	settings.stockData()

	resp := u.Message(true, "Cross creator main page, formed a crosses map")
	resp["crossMap"] = mapRegAreaCross
	resp["pngSettings"] = settings
	resp["region"] = CacheInfo.mapRegion
	resp["area"] = CacheInfo.mapArea
	return resp
}

//CheckCrossDirFromBD проверяет ВСЕ перекрестки из БД на наличие каталого для них и заполнения
func CheckCrossDirFromBD() map[string]interface{} {
	//CacheInfoDataUpdate()
	tfData := GetAllTrafficLights()
	path := os.Getenv("views_path") + "//cross"
	var tempTF []TrafficLights
	for _, tfLight := range tfData {
		_, err1 := os.Stat(path + fmt.Sprintf("//%v//%v//%v//map.png", tfLight.Region.Num, tfLight.Area.Num, tfLight.ID))
		_, err2 := os.Stat(path + fmt.Sprintf("//%v//%v//%v//cross.svg", tfLight.Region.Num, tfLight.Area.Num, tfLight.ID))
		if os.IsNotExist(err1) || os.IsNotExist(err2) {
			tempTF = append(tempTF, tfLight)
		}
	}
	resp := make(map[string]interface{})
	resp["tf"] = tempTF
	return resp
}

//CheckCrossDirSelected проверяет региона/районы/перекрестки которые запросил пользователь
func CheckCrossFileSelected(selectedData map[string]map[string][]CheckData) map[string]interface{} {
	path := os.Getenv("views_path") + "//cross"
	for numFirst, firstMap := range selectedData {
		for numSecond, secondMap := range firstMap {
			for numCheck, check := range secondMap {
				_, err1 := os.Stat(path + fmt.Sprintf("//%v//%v//%v//map.png", numFirst, numSecond, check.ID))
				if !os.IsNotExist(err1) {
					selectedData[numFirst][numSecond][numCheck].PngStatus = true
				}
				_, err2 := os.Stat(path + fmt.Sprintf("//%v//%v//%v//cross.svg", numFirst, numSecond, check.ID))
				if !os.IsNotExist(err2) {
					selectedData[numFirst][numSecond][numCheck].SvgStatus = true
				}
			}
		}
	}
	resp := make(map[string]interface{})
	resp["verifiedData"] = selectedData
	return resp
}

//MakeSelectedDir создание каталогов и файлов png + svg у выбранных
func MakeSelectedDir(selData SelectedData) map[string]interface{} {
	if selData.PngSettings.SizeX == "" || selData.PngSettings.SizeY == "" || selData.PngSettings.Z == "" {
		selData.PngSettings.stockData()
	}
	path := os.Getenv("views_path") + "//cross"
	for numFirst, firstMap := range selData.SelectedData {
		for numSecond, secondMap := range firstMap {
			for numCheck, check := range secondMap {
				err := os.MkdirAll(path+fmt.Sprintf("//%v//%v//%v", numFirst, numSecond, check.ID), os.ModePerm)
				if err != nil {
					selData.SelectedData[numFirst][numSecond][numCheck].setStatusFalse()
					continue
				}
				if !selData.SelectedData[numFirst][numSecond][numCheck].PngStatus {
					fmt.Println("png")
				}
				if !selData.SelectedData[numFirst][numSecond][numCheck].SvgStatus {
					fmt.Println("svg")
				}
			}
		}
	}
	resp := make(map[string]interface{})
	resp["makeData"] = selData
	return resp
}

var (
	str1 = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
			<svg
			xmlns:svg="http://www.w3.org/2000/svg"
			xmlns="http://www.w3.org/2000/svg"
			width="450mm"
			height="450mm"
			viewBox="0 0 450 450">
			<foreignObject x="5" y="5" width="100" height="450">
			<div xmlns="http://www.w3.org/1999/xhtml"
		style="font-size:8px;font-family:sans-serif">`
	str2 = `</div>
			</foreignObject>
 			</svg>`
)

func createPng(path string) (err error) {
	//url := fmt.Sprintf("https://static-maps.yandex.ru/1.x/?ll=%3.15f,%3.15f&z=19&l=map&size=450,450", TL.Points.Y, TL.Points.X)
	return err
}

func createSvg(path string) (err error) {

	return err
}
