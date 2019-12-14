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
	mux    sync.Mutex
	Region map[int]string
	TLSost map[int]string
}

//RegionInfo расшифровка региона
type RegionInfo struct {
	Num  int    //уникальный номер региона
	Name string //расшифровка номера
}

//TLSostInfo состояние
type InfoTL  struct {
	Sost []TLSostInfo `json:"sost"`
}

//InfoTL информация о состояния светофоров
type TLSostInfo struct {
	Num         int    `json:"num"`
	Description string `json:"description"`
}

func CacheDataUpdate() {
	var err error
	for {
		CacheInfo.mux.Lock()
		CacheInfo.Region, err = GetRegionInfo()
		CacheInfo.TLSost, err = GetTLSost()
		CacheInfo.mux.Unlock()
		if err != nil {
			logger.Info.Println("Произошла ошибка в чтении cache данных :", err)
		}
		time.Sleep(time.Hour)
	}

}

//GetRegionInfo получить таблицу регионов
func GetRegionInfo() (region map[int]string, err error) {
	region = make(map[int]string)
	sqlStr := fmt.Sprintf("select region, name from %s", os.Getenv("region_table"))
	rows, _ := GetDB().Raw(sqlStr).Rows()
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

func GetTLSost() (TLsost map[int]string, err error) {
	TLsost = make(map[int]string)
	file, err := ioutil.ReadFile("./TLsost.js")
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
