package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"../logger"
)

//CacheInfo глобальная переменная для обращения к данным
var CacheInfo CacheData

//CacheData Данные для обновления в определенный период
type CacheData struct {
	mux       sync.Mutex
	mapRegion map[string]string            //регионы
	mapArea   map[string]map[string]string //районы
	mapTLSost map[int]string               //светофоры
	mapRoles  map[string]Permissions       //роли
}

//RegionInfo расшифровка региона
type RegionInfo struct {
	Num        string `json:"num"`        //уникальный номер региона
	NameRegion string `json:"nameRegion"` //расшифровка номера
}

//AreaInfo расшифровка района
type AreaInfo struct {
	Num      string `json:"num"`      //уникальный номер зоны
	NameArea string `json:"nameArea"` //расшифровка номера
}

//InfoTL массив для приема пользовательского запроса
type InfoTL struct {
	Sost []TLSostInfo `json:"sost"`
}

//TLSostInfo информация о состояния светофоров
type TLSostInfo struct {
	Num         int    `json:"num"`
	Description string `json:"description"`
}

//CacheDataUpdate обновление данных из бд, период обновления 1 час
func CacheDataUpdate() {
	CacheInfo.mapRoles = make(map[string]Permissions)
	BusyArmInfo.mapBusyArm = make(map[BusyArm]EditCrossInfo)
	for {
		CacheInfoDataUpdate()
		//создадим суперпользователя если таблица только была создана
		if FirstCreate {
			FirstCreate = false
			// Супер пользователь
			_ = SuperCreate()
		}

		time.Sleep(time.Hour)
	}
}

//CacheInfoDataUpdate заполнение структуры cacheInfo
func CacheInfoDataUpdate() {
	var err error
	CacheInfo.mux.Lock()
	CleanMapBusyArm()
	CacheInfo.mapRegion, CacheInfo.mapArea, err = GetRegionInfo()
	CacheInfo.mapTLSost, err = getTLSost()
	err = getRoles()
	CacheInfo.mux.Unlock()
	if err != nil {
		logger.Error.Println(fmt.Sprintf("|Message: Error reading data cache: %s", err.Error()))
	}
}

//GetRegionInfo получить таблицу регионов
func GetRegionInfo() (region map[string]string, area map[string]map[string]string, err error) {
	region = make(map[string]string)
	area = make(map[string]map[string]string)
	sqlStr := fmt.Sprintf("select region, nameregion, area, namearea from %s", os.Getenv("region_table"))
	rows, err := GetDB().Raw(sqlStr).Rows()
	if err != nil {
		return CacheInfo.mapRegion, CacheInfo.mapArea, err
	}
	for rows.Next() {
		var (
			tempReg  = &RegionInfo{}
			tempArea = &AreaInfo{}
		)
		err = rows.Scan(&tempReg.Num, &tempReg.NameRegion, &tempArea.Num, &tempArea.NameArea)
		if err != nil {
			return nil, nil, err
		}
		if _, ok := region[tempReg.Num]; !ok {
			region[tempReg.Num] = tempReg.NameRegion
		}

		if _, ok := area[tempReg.NameRegion][tempArea.Num]; !ok {
			if _, ok := area[tempReg.NameRegion]; !ok {
				area[tempReg.NameRegion] = make(map[string]string)
			}
			area[tempReg.NameRegion][tempArea.Num] = tempArea.NameArea
		}
	}
	if _, ok := region["*"]; !ok {
		region["*"] = "Все регионы"
	}
	if _, ok := area["Все регионы"]["*"]; !ok {
		area["Все регионы"] = make(map[string]string)
		area["Все регионы"]["*"] = "Все районы"
	}

	return region, area, err
}

//getTLSost получить данные о состоянии светофоров
func getTLSost() (TLsost map[int]string, err error) {
	TLsost = make(map[int]string)
	file, err := ioutil.ReadFile("./cachefile/TLsost.json")
	if err != nil {
		return nil, err
	}
	temp := new(InfoTL)
	if err := json.Unmarshal(file, &temp); err != nil {
		return nil, err
	}
	for _, sost := range temp.Sost {
		if _, ok := TLsost[sost.Num]; !ok {
			TLsost[sost.Num] = sost.Description
		}
	}
	return TLsost, err
}

//getRoles получать данны о ролях
func getRoles() (err error) {
	var temp = Roles{}
	err = temp.ReadRoleFile()
	if err != nil {
		return err
	}
	for _, role := range temp.Roles {
		if _, ok := CacheInfo.mapRoles[role.Name]; !ok {
			CacheInfo.mapRoles[role.Name] = role.Perm
		}
	}
	return err
}

//SetRegionInfo установить в структуру номер и имя региона по номеру
func (region *RegionInfo) SetRegionInfo(num string) {
	region.Num = num
	CacheInfo.mux.Lock()
	region.NameRegion = CacheInfo.mapRegion[num]
	CacheInfo.mux.Unlock()
}

//SetAreaInfo установить в структуру номер и имя района по номеру района и региона
func (area *AreaInfo) SetAreaInfo(numReg, numArea string) {
	area.Num = numArea
	CacheInfo.mux.Lock()
	area.NameArea = CacheInfo.mapArea[CacheInfo.mapRegion[numReg]][numArea]
	CacheInfo.mux.Unlock()
}
