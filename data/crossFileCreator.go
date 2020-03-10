package data

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/JanFant/TLServer/logger"
	u "github.com/JanFant/TLServer/utils"
)

//SelectedData общая струкрута для обмена
type SelectedData struct {
	SelectedData map[string]map[string][]CheckData `json:"selected"`    //хранилище перекрестков которые были выбраны
	PngSettings  PngSettings                       `json:"pngSettings"` //настройки для создания map.png
}

//CheckData структура проверки для перекрестков
type CheckData struct {
	ID        string `json:"ID"`        //ID устройства
	PngStatus bool   `json:"pngStatus"` //флаг наличия map.png
	SvgStatus bool   `json:"svgStatus"` //флаг наличия cross.svg
}

//PngSettings настройки размеров создаваемой map.png
type PngSettings struct {
	SizeX int `json:"sizeX",toml:"png_sizeX"` //размер картинки по координате X
	SizeY int `json:"sizeY",toml:"png_sizeY"` //размер картинки по координате Y
	Z     int `json:"z",toml:"png_Z"`         //величина отдаление
}

//setStatusTrue установить значение в True
func (checkData *CheckData) setStatusTrue() {
	checkData.SvgStatus = true
	checkData.PngStatus = true
}

//setStatusFalse установить значение в False
func (checkData *CheckData) setStatusFalse() {
	checkData.SvgStatus = false
	checkData.PngStatus = false
}

//stockData заполняет поля из env файла
func (set *PngSettings) stockData() {
	set.SizeX = GlobalConfig.PngSettings.SizeX
	set.SizeY = GlobalConfig.PngSettings.SizeY
	set.Z = GlobalConfig.PngSettings.Z
}

//MainCrossCreator формираю необходимые данные для начальной странички с деревом
func MainCrossCreator() map[string]interface{} {
	//CacheInfoDataUpdate()
	tfData := GetAllTrafficLights()
	mapRegAreaCross := make(map[string]map[string][]TrafficLights)
	CacheInfo.mux.Lock()
	defer CacheInfo.mux.Unlock()
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
	path := GlobalConfig.ViewsPath + "//cross"
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

//CheckCrossFileSelected проверяет региона/районы/перекрестки которые запросил пользователь
func CheckCrossFileSelected(selectedData map[string]map[string][]CheckData) map[string]interface{} {
	path := GlobalConfig.ViewsPath + "//cross"
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
	var (
		message []string
		count   = 0
	)
	sizeX := selData.PngSettings.SizeX
	sizeY := selData.PngSettings.SizeY
	if selData.PngSettings.SizeX == 0 || selData.PngSettings.SizeY == 0 || selData.PngSettings.Z == 0 || sizeX > 450 || sizeX < 0 || sizeY < 0 || sizeY > 450 {
		selData.PngSettings.stockData()
	}
	path := GlobalConfig.ViewsPath + "//cross"
	for numFirst, firstMap := range selData.SelectedData {
		for numSecond, secondMap := range firstMap {
			for numCheck, check := range secondMap {
				err := os.MkdirAll(path+fmt.Sprintf("//%v//%v//%v", numFirst, numSecond, check.ID), os.ModePerm)
				if err != nil {
					selData.SelectedData[numFirst][numSecond][numCheck].setStatusFalse()
					continue
				}
				if !selData.SelectedData[numFirst][numSecond][numCheck].PngStatus {
					point, err := TakePointFromBD(numFirst, numSecond, check.ID)
					if err != nil {
						logger.Error.Println(fmt.Sprintf("|Message: No result at point: (%v//%v//%v)", numFirst, numSecond, check.ID))
						if count == 0 {
							message = append(message, "Не созданны")
						}
						count++
						message = append(message, fmt.Sprintf("%v: (%v//%v//%v)", count, numFirst, numSecond, check.ID))
						continue
					}
					err = createPng(numFirst, numSecond, check.ID, selData.PngSettings, point)
					if err != nil {
						logger.Error.Println(fmt.Sprintf("|Message: Can't create map.png path = %v//%v//%v", numFirst, numSecond, check.ID))
						continue
					}
					selData.SelectedData[numFirst][numSecond][numCheck].PngStatus = true
				}
			}
		}
	}
	resp := make(map[string]interface{})
	if len(message) == 0 {
		resp = u.Message(true, "Everything was created")
	} else {
		resp = u.Message(false, "There are not created")
		resp["notCreated"] = message
	}
	resp["makeData"] = selData
	return resp
}

//ShortCreateDirPng создание директории по одной
func ShortCreateDirPng(region, area, id int, pointStr string) bool {
	var (
		pngSettings PngSettings
		point       Point
	)
	pngSettings.stockData()
	point.StrToFloat(pointStr)
	path := GlobalConfig.ViewsPath + "//cross"
	_ = os.MkdirAll(path+fmt.Sprintf("//%v//%v//%v", region, area, id), os.ModePerm)
	err := createPng(strconv.Itoa(region), strconv.Itoa(area), strconv.Itoa(id), pngSettings, point)
	if err != nil {
		logger.Error.Println(fmt.Sprintf("|Message: Can't create map.png path = %v//%v//%v", region, area, id))
		return false
	}
	_, err = os.Stat(path + fmt.Sprintf("//%v//%v//%v//cross.svg", region, area, id))
	if os.IsNotExist(err) {
		file1, err := os.Create(path + fmt.Sprintf("//%v//%v//%v//cross.svg", region, area, id))
		if err != nil {
			return false
		}
		defer file1.Close()
		_, err = fmt.Fprintln(file1, str1, str2)
		if err != nil {
			return false
		}
	}
	return true
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

//createPng создание png файла
func createPng(numReg, numArea, id string, settings PngSettings, point Point) (err error) {
	url := fmt.Sprintf("https://static-maps.yandex.ru/1.x/?ll=%3.15f,%3.15f&z=%v&l=map&size=%v,%v", point.X, point.Y, settings.Z, settings.SizeX, settings.SizeY)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	filePath := GlobalConfig.ViewsPath + "//cross" + "//" + numReg + "//" + numArea + "//" + id + "//"
	//open a file for writing
	file, err := os.Create(filePath + "map.png")
	if err != nil {
		return err
	}
	defer file.Close()
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}
