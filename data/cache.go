package data

import (
	"../logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

//CacheInfo Данные для обновления в определенный период
var CacheInfo CacheData

type CacheData struct {
	mux       sync.Mutex
	mapRegion map[string]string
	mapArea   map[string]map[string]string
	mapTLSost map[int]string
	mapRoles  map[string]Permissions
}

//RegionInfo расшифровка региона
type RegionInfo struct {
	Num        string `json:"num"`        //уникальный номер региона
	NameRegion string `json:"nameRegion"` //расшифровка номера
}

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

func CacheDataUpdate() {
	var err error
	CacheInfo.mapRoles = make(map[string]Permissions)
	for {
		CacheInfo.mux.Lock()
		CacheInfo.mapRegion, CacheInfo.mapArea, err = GetRegionInfo()
		CacheInfo.mapTLSost, err = GetTLSost()
		err = GetRoles()

		CacheInfo.mux.Unlock()

		if err != nil {
			logger.Error.Println(fmt.Sprintf("|Message: Error reading data cache: %s", err.Error()))
		}
		//создадим суперпользователя если таблица только была создана
		if FirstCreate {
			FirstCreate = false
			// Супер пользователь
			_ = SuperCreate()
		}

		time.Sleep(time.Hour)
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
	return region, area, err
}

//GetTLSost получить данные о состоянии светофоров
func GetTLSost() (TLsost map[int]string, err error) {
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

func GetRoles() (err error) {
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

func (region *RegionInfo) SetRegionInfo(num string) {
	region.Num = num
	if strings.EqualFold(region.Num, "*") {
		region.NameRegion = "Все регионы"
	} else {
		region.NameRegion = CacheInfo.mapRegion[num]
	}
}

func (area *AreaInfo) SetAreaInfo(numReg, numArea string) {
	area.Num = numArea
	if strings.EqualFold(area.Num, "*") {
		area.NameArea = "Все районы"
	} else {
		area.NameArea = CacheInfo.mapArea[CacheInfo.mapRegion[numReg]][numArea]
	}

}
