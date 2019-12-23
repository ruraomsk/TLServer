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

//CacheInfo Данные для обновления в определенный период
var CacheInfo CacheData

type CacheData struct {
	mux       sync.Mutex
	mapRegion map[int]string
	mapTLSost map[int]string
	mapRoles  map[string]Permissions
}

//RegionInfo расшифровка региона
type RegionInfo struct {
	Num  int    `json:"num"`  //уникальный номер региона
	Name string `json:"name"` //расшифровка номера
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
		CacheInfo.mapRegion, err = GetRegionInfo()
		CacheInfo.mapTLSost, err = GetTLSost()
		err = GetRoles()

		CacheInfo.mux.Unlock()
		if err != nil {
			logger.Info.Println("Cache: Произошла ошибка в чтении cache данных :", err)
		}
		time.Sleep(time.Hour)
	}

}

//GetRegionInfo получить таблицу регионов
func GetRegionInfo() (region map[int]string, err error) {
	region = make(map[int]string)
	sqlStr := fmt.Sprintf("select region, name from %s", os.Getenv("region_table"))
	rows, err := GetDB().Raw(sqlStr).Rows()
	if err != nil {
		return CacheInfo.mapRegion, err
	}
	for rows.Next() {
		temp := &RegionInfo{}
		err = rows.Scan(&temp.Num, &temp.Name)
		if err != nil {
			return nil, err
		}
		if _, ok := region[temp.Num]; !ok {
			region[temp.Num] = temp.Name
		}
	}
	return region, err
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
