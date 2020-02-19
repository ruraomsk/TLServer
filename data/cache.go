package data

import (
	"../logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

//CacheInfo глобальная переменная для обращения к данным
var CacheInfo CacheData

//CacheData Данные для обновления в определенный период
type CacheData struct {
	mux        sync.Mutex
	mapRegion  map[string]string
	mapArea    map[string]map[string]string
	mapTLSost  map[int]string
	mapRoles   map[string]Permissions
	mapBusyArm map[BusyArm]EditCrossInfo
}

type BusyArm struct {
	Region string `json:"region"`
	Area   string `json:"area"`
	ID     int    `json:"ID"`
}

//EditCrossInfo информация о пользователе занявшем перекресток на изменение
type EditCrossInfo struct {
	Login    string `json:"login"`
	EditFlag bool   `json:"editFlag"`
	Kick     bool   `json:"kick"`
	time     time.Time
}

//RegionInfo расшифровка региона
type RegionInfo struct {
	Num        string `json:"num"`        //уникальный номер региона
	NameRegion string `json:"nameRegion"` //расшифровка номера
}

//AreaInfo расшифровка раона
type AreaInfo struct {
	Num      string `json:"num"`      //уникальный номер зоны
	NameArea string `json:"nameArea"` //расшифровка номера
}

//TLSostInfo состояние
type InfoTL struct {
	Sost []TLSostInfo `json:"sost"`
}

//InfoTL информация о состояния светофоров
type TLSostInfo struct {
	Num         int    `json:"num"`
	Description string `json:"description"`
}

//CacheDataUpdate обновление данных из бд, период обновления 1 час
func CacheDataUpdate() {
	CacheInfo.mapRoles = make(map[string]Permissions)
	CacheInfo.mapBusyArm = make(map[BusyArm]EditCrossInfo)
	go func() {
		for {
			fmt.Println(CacheInfo.mapBusyArm)
			time.Sleep(time.Second * 5)
		}
	}()
	go func() {
		for {
			CleanMapBusyArm()
			time.Sleep(time.Second * 20)
		}
	}()
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
	CacheInfo.mapRegion, CacheInfo.mapArea, err = GetRegionInfo()
	CacheInfo.mapTLSost, err = getTLSost()
	err = getRoles()
	CacheInfo.mux.Unlock()
	if err != nil {
		logger.Error.Println(fmt.Sprintf("|Message: Error reading data cache: %s", err.Error()))
	}
}

func CleanMapBusyArm() {
	for busyArm, editCross := range CacheInfo.mapBusyArm {
		if editCross.time.Add(time.Second * 10).Before(time.Now()) {
			delete(CacheInfo.mapBusyArm, busyArm)
		}
	}
}

func BusyArmDelete(arm BusyArm) map[string]interface{} {
	resp := make(map[string]interface{})
	delete(CacheInfo.mapBusyArm, arm)
	resp["ArmDelete"] = arm
	return resp
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
	region.NameRegion = CacheInfo.mapRegion[num]
}

//SetAreaInfo установить в структуру номер и имя района по номеру района и региона
func (area *AreaInfo) SetAreaInfo(numReg, numArea string) {
	area.Num = numArea
	area.NameArea = CacheInfo.mapArea[CacheInfo.mapRegion[numReg]][numArea]
}
