package data

import (
	"fmt"
	"sync"
	"time"

	"github.com/JanFant/TLServer/logger"
)

//CacheInfo глобальная переменная для обращения к данным
var CacheInfo CacheData

//CacheData Данные для обновления в определенный период
type CacheData struct {
	mux       sync.Mutex
	mapRegion map[string]string            //регионы
	mapArea   map[string]map[string]string //районы
	mapTLSost map[int]string               //светофоры

}

//RegionInfo расшифровка региона
type RegionInfo struct {
	Num        string `json:"num"`        //уникальный номер региона
	NameRegion string `json:"nameRegion"` //расшифровка номера
}

//AreaInfo информация о районе
type AreaInfo struct {
	Num      string `json:"num"`      //уникальный номер района
	NameArea string `json:"nameArea"` //расшифровка номера
}

//InfoTL массив для приема пользовательского запроса
type InfoTL struct {
	Sost []TLSostInfo `json:"sost"`
}

//TLSostInfo информация о состояния светофоров
type TLSostInfo struct {
	Num         int    `json:"num"`         //номер состояния
	Description string `json:"description"` //описание состояния
}

//CacheDataUpdate обновление данных из бд, период обновления 1 час
func CacheDataUpdate() {
	RoleInfo.MapRoles = make(map[string][]int)
	RoleInfo.MapPermisson = make(map[int]Permission)
	RoleInfo.MapRoutes = make(map[string]RouteInfo)
	BusyArmInfo.mapBusyArm = make(map[BusyArm]EditCrossInfo)
	go FillingDeviceLogTable()
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
	CacheInfo.mux.Unlock()
	err = getRoleAccess()
	if err != nil {
		logger.Error.Println(fmt.Sprintf("|Message: Error reading data cache: %s", err.Error()))
	}
}

//GetRegionInfo получить таблицу регионов
func GetRegionInfo() (region map[string]string, area map[string]map[string]string, err error) {
	region = make(map[string]string)
	area = make(map[string]map[string]string)
	sqlStr := fmt.Sprintf("select region, nameregion, area, namearea from %s", GlobalConfig.DBConfig.RegionTable)
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
	sqlStr := fmt.Sprintf(`Select id, description from %v`, GlobalConfig.DBConfig.StatusTable)
	statusRow, err := GetDB().Raw(sqlStr).Rows()
	if err != nil {
		logger.Error.Println("|Message: GetTLSost StatusTable error : ", err.Error())
		return nil, err
	}
	for statusRow.Next() {
		var (
			id   int
			desc string
		)
		err := statusRow.Scan(&id, &desc)
		if err != nil {
			logger.Error.Println("|Message: No result at these points", err.Error())
			return nil, err
		}
		if _, ok := TLsost[id]; !ok {
			TLsost[id] = desc
		}
	}
	return TLsost, err
}

//getRoleAccess получить информацию о ролях из файла RoleAccess.json
func getRoleAccess() (err error) {
	var temp = RoleAccess{}
	err = temp.ReadRoleAccessFile()
	if err != nil {
		return err
	}
	RoleInfo.mux.Lock()
	for _, role := range temp.Roles {
		if _, ok := RoleInfo.MapRoles[role.Name]; !ok {
			RoleInfo.MapRoles[role.Name] = role.Perm
		}
	}
	for _, perm := range temp.Permission {
		if _, ok := RoleInfo.MapPermisson[perm.ID]; !ok {
			RoleInfo.MapPermisson[perm.ID] = perm
		}
	}
	for _, route := range temp.Routes {
		if _, ok := RoleInfo.MapRoutes[route.Path]; !ok {
			RoleInfo.MapRoutes[route.Path] = route
		}
	}
	RoleInfo.mux.Unlock()
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
