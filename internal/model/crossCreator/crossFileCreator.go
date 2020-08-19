package crossCreator

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/JanFant/TLServer/internal/model/data"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/JanFant/TLServer/internal/model/locations"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
)

//SelectedData общая структура обмена
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
	SizeX int `json:"sizeX",toml:"png_sizeX"` //размер по координате X
	SizeY int `json:"sizeY",toml:"png_sizeY"` //размер по координате Y
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
	set.SizeX = 450
	set.SizeY = 450
	set.Z = 19
}

//MainCrossCreator формирование необходимых данных для начальной странички с деревом
func MainCrossCreator() u.Response {
	//CacheInfoDataUpdate()
	tfData := data.GetAllTrafficLights()
	mapRegAreaCross := make(map[string]map[string][]data.TrafficLights)
	data.CacheInfo.Mux.Lock()
	defer data.CacheInfo.Mux.Unlock()
	for numReg, nameReg := range data.CacheInfo.MapRegion {
		if strings.Contains(numReg, "*") {
			continue
		}
		for numArea, nameArea := range data.CacheInfo.MapArea[nameReg] {
			if strings.Contains(numArea, "Все регионы") {
				continue
			}
			if _, ok := mapRegAreaCross[nameReg]; !ok {
				mapRegAreaCross[nameReg] = make(map[string][]data.TrafficLights)
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

	resp := u.Message(http.StatusOK, "cross creator main page, formed a crosses map")
	resp.Obj["crossMap"] = mapRegAreaCross
	resp.Obj["pngSettings"] = settings
	resp.Obj["region"] = data.CacheInfo.MapRegion
	resp.Obj["area"] = data.CacheInfo.MapArea
	return resp
}

//CheckCrossDirFromBD проверяет ВСЕ перекрестки из БД на наличие каталогов для них
func CheckCrossDirFromBD() u.Response {
	//CacheInfoDataUpdate()
	tfData := data.GetAllTrafficLights()
	path := config.GlobalConfig.StaticPath + "//cross"
	var tempTF []data.TrafficLights
	for _, tfLight := range tfData {
		_, err1 := os.Stat(path + fmt.Sprintf("//%v//%v//%v//map.png", tfLight.Region.Num, tfLight.Area.Num, tfLight.ID))
		_, err2 := os.Stat(path + fmt.Sprintf("//%v//%v//%v//cross.svg", tfLight.Region.Num, tfLight.Area.Num, tfLight.ID))
		if os.IsNotExist(err1) || os.IsNotExist(err2) {
			data.CacheInfo.Mux.Lock()
			tfLight.Region.NameRegion = data.CacheInfo.MapRegion[tfLight.Region.Num]
			tfLight.Area.NameArea = data.CacheInfo.MapArea[tfLight.Region.NameRegion][tfLight.Area.Num]
			data.CacheInfo.Mux.Unlock()
			tempTF = append(tempTF, tfLight)
		}
	}
	resp := u.Message(http.StatusOK, "all cross")
	resp.Obj["tf"] = tempTF
	return resp
}

//CheckCrossFileSelected проверяет региона/районы/перекрестки которые запросил пользователь
func CheckCrossFileSelected(selectedData map[string]map[string][]CheckData) u.Response {
	path := config.GlobalConfig.StaticPath + "//cross"
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
	resp := u.Message(http.StatusOK, "check cross")
	resp.Obj["verifiedData"] = selectedData
	return resp
}

//MakeSelectedDir создание каталогов и файлов png + svg у выбранных перекрестков
func MakeSelectedDir(selData SelectedData) u.Response {
	var (
		message []string
		count   = 0
	)
	sizeX := selData.PngSettings.SizeX
	sizeY := selData.PngSettings.SizeY
	if selData.PngSettings.SizeX == 0 || selData.PngSettings.SizeY == 0 || selData.PngSettings.Z == 0 || sizeX > 450 || sizeX < 0 || sizeY < 0 || sizeY > 450 {
		selData.PngSettings.stockData()
	}
	path := config.GlobalConfig.StaticPath + "//cross"
	for numFirst, firstMap := range selData.SelectedData {
		for numSecond, secondMap := range firstMap {
			for numCheck, check := range secondMap {
				err := os.MkdirAll(path+fmt.Sprintf("//%v//%v//%v", numFirst, numSecond, check.ID), os.ModePerm)
				if err != nil {
					selData.SelectedData[numFirst][numSecond][numCheck].setStatusFalse()
					continue
				}
				if !selData.SelectedData[numFirst][numSecond][numCheck].PngStatus {
					point, err := locations.TakePointFromBD(numFirst, numSecond, check.ID, data.GetDB())
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
	resp := u.Message(http.StatusOK, "check cross")
	if len(message) == 0 {
		resp.Obj["notCreated"] = message
	}
	resp.Obj["makeData"] = selData
	return resp
}

//ShortCreateDirPng создание каталога
func ShortCreateDirPng(region, area, id, z int, pointStr string) bool {
	var (
		pngSettings PngSettings
		point       locations.Point
	)
	pngSettings.stockData()
	point.StrToFloat(pointStr)
	path := config.GlobalConfig.StaticPath + "//cross"
	_ = os.MkdirAll(path+fmt.Sprintf("//%v//%v//%v", region, area, id), os.ModePerm)
	if z > 0 {
		pngSettings.Z = z
	}
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
		_, err = fmt.Fprintln(file1, svgStr)
		if err != nil {
			return false
		}
	}
	return true
}

var (
	svgStr = `<svg height="450" width="450" 
		xmlns:svg="http://www.w3.org/2000/svg" 
		xmlns:xlink="http://www.w3.org/1999/xlink" 
		xmlns:xsl="http://www.w3.org/1999/XSL/Transform" 
		xmlns="http://www.w3.org/2000/svg" 
		xlinkn="http://www.w3.org/1999/xlink" 
		viewBox="0 0 450 450">
 		<circle cx="225" cy="225" r="14" 
            	stroke="black"
            	stroke-width="1"  
            	fill="blue" />
    		<text id = "midText" x="225" y="225" 
          	font-size="11"
          	dy=".3em"
          	text-anchor="middle" 
          	stroke="white" 
          	stroke-width="1px"
          	> 0Ф </text>
     	<script>
			function setPhase(phase){
    			var midText = document.getElementById('midText'); midText.textContent = phase + "Ф";
			}
 	</script>
	</svg>`
)

//createPng создание png файла
func createPng(numReg, numArea, id string, settings PngSettings, point locations.Point) (err error) {
	url := fmt.Sprintf("https://static-maps.yandex.ru/1.x/?ll=%3.15f,%3.15f&z=%v&l=map&size=%v,%v", point.X, point.Y, settings.Z, settings.SizeX, settings.SizeY)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	filePath := config.GlobalConfig.StaticPath + "//cross" + "//" + numReg + "//" + numArea + "//" + id + "//"
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
