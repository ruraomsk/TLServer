package data

import (
	"github.com/JanFant/TLServer/internal/model/locations"
	"sync"
	"time"

	"github.com/JanFant/TLServer/logger"
)

//CacheInfo глобальная переменная для обращения к данным
var CacheInfo CacheData

//CacheArea глобальная переменная для обращения к данным области
var CacheArea locations.AreaOnMap

//CacheData Данные для обновления в определенный период
type CacheData struct {
	Mux       sync.Mutex
	MapRegion map[string]string            //регионы
	MapArea   map[string]map[string]string //районы
	MapTLSost map[int]TLSostInfo           //светофоры

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
	Control     bool   `json:"control"`     //признак управления в этом режиме
}

//CacheDataUpdate обновление данных из бд, период обновления 1 час
func CacheDataUpdate() {
	RoleInfo.Mux.Lock()
	RoleInfo.MapRoles = make(map[string][]int)
	RoleInfo.MapPermisson = make(map[int]Permission)
	RoleInfo.MapRoutes = make(map[string]RouteInfo)
	RoleInfo.Mux.Unlock()
	for {
		CacheInfoDataUpdate()
		//создадим суперпользователя если таблица только была создана
		if FirstCreate {
			FirstCreate = false
			// Супер пользователь
			SuperCreate()
		}
		time.Sleep(time.Hour)
	}
}

//CacheInfoDataUpdate заполнение структуры cacheInfo
func CacheInfoDataUpdate() {
	var err error
	CacheInfo.Mux.Lock()
	CacheInfo.MapRegion, CacheInfo.MapArea, err = GetRegionInfo()
	CacheInfo.MapTLSost, err = getTLSost()
	CacheInfo.Mux.Unlock()
	FillMapAreaZone()
	err = getRoleAccess()
	if err != nil {
		logger.Error.Printf("|IP: Server  |Login: Server |Resource: Server |Message: %v \n", err.Error())
	}
}

//FillMapAreaBox заполнение мапы районов и регионов с координатами
func FillMapAreaZone() {
	tempAreaCache := make([]locations.AreaZone, 0)
	//запрос уникальных регионов и районов
	rows, err := GetDB().Query(`SELECT distinct on (region, area) region, area, array_agg(dgis) as aDgis   FROM public.cross  group by region, area`)
	if err != nil {
		logger.Error.Printf("|IP: Server  |Login: Server |Resource: Server |Message: %v \n", err.Error())
		return
	}
	for rows.Next() {
		var (
			temp     locations.AreaZone
			arrayStr string
		)
		_ = rows.Scan(&temp.Region, &temp.Area, &arrayStr)
		temp.Zone.ParseFromStr(arrayStr)
		temp.Zone = temp.Zone.ConvexHull()
		temp.Sub = make([]locations.SybAreaZone, 0)
		tempAreaCache = append(tempAreaCache, temp)
	}

	//запрос уникальных подрайонов
	rows, err = GetDB().Query(`SELECT distinct on (region, area, subarea) region, area, subarea, array_agg(dgis) as aDgis FROM public.cross  group by region, area, subarea`)
	if err != nil {
		logger.Error.Printf("|IP: Server  |Login: Server |Resource: Server |Message: %v \n", err.Error())
		return
	}
	for rows.Next() {
		var (
			tempSyb  locations.SybAreaZone
			reg      string
			area     string
			arrayStr string
		)
		_ = rows.Scan(&reg, &area, &tempSyb.SubArea, &arrayStr)
		tempSyb.Zone.ParseFromStr(arrayStr)
		if len(tempSyb.Zone) > 1 {
			tempSyb.Zone = tempSyb.Zone.ConvexHull()
		}
		for num, areaBox := range tempAreaCache {
			if areaBox.Region == reg && areaBox.Area == area {
				tempAreaCache[num].Sub = append(tempAreaCache[num].Sub, tempSyb)
			}
		}

	}

	CacheInfo.Mux.Lock()
	//заполним поля названиями
	for num := range tempAreaCache {
		tempAreaCache[num].Region = CacheInfo.MapRegion[tempAreaCache[num].Region]
		tempAreaCache[num].Area = CacheInfo.MapArea[tempAreaCache[num].Region][tempAreaCache[num].Area]
	}
	CacheInfo.Mux.Unlock()

	CacheArea.Mux.Lock()
	CacheArea.Areas = tempAreaCache
	CacheArea.Mux.Unlock()

	go CreateCrossesJSON(db)
}

//GetRegionInfo получить таблицу регионов
func GetRegionInfo() (region map[string]string, area map[string]map[string]string, err error) {
	region = make(map[string]string)
	area = make(map[string]map[string]string)
	rows, err := GetDB().Query(`SELECT region, nameregion, area, namearea FROM public.region`)
	if err != nil {
		return CacheInfo.MapRegion, CacheInfo.MapArea, err
	}
	for rows.Next() {
		var (
			tempReg  = &RegionInfo{}
			tempArea = &AreaInfo{}
		)
		err = rows.Scan(&tempReg.Num, &tempReg.NameRegion, &tempArea.Num, &tempArea.NameArea)
		if err != nil {
			return CacheInfo.MapRegion, CacheInfo.MapArea, err
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
func getTLSost() (TLsost map[int]TLSostInfo, err error) {
	TLsost = make(map[int]TLSostInfo, 0)
	statusRow, err := GetDB().Query(`SELECT id, description, control FROM public.status`)
	if err != nil {
		logger.Error.Printf("|IP: Server  |Login: Server |Resource: Server |Message: %v \n", err.Error())
		return TLsost, err
	}
	for statusRow.Next() {
		var tempTL TLSostInfo
		err := statusRow.Scan(&tempTL.Num, &tempTL.Description, &tempTL.Control)
		if err != nil {
			logger.Error.Printf("|IP: Server  |Login: Server |Resource: Server |Message: %v \n", err.Error())
			return TLsost, err
		}
		if _, ok := TLsost[tempTL.Num]; !ok {
			TLsost[tempTL.Num] = tempTL
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
	RoleInfo.Mux.Lock()
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
	RoleInfo.Mux.Unlock()
	return err
}

//SetRegionInfo установить в структуру номер и имя региона по номеру
func (region *RegionInfo) SetRegionInfo(num string) {
	region.Num = num
	CacheInfo.Mux.Lock()
	region.NameRegion = CacheInfo.MapRegion[num]
	CacheInfo.Mux.Unlock()
}

//SetAreaInfo установить в структуру номер и имя района по номеру района и региона
func (area *AreaInfo) SetAreaInfo(numReg, numArea string) {
	area.Num = numArea
	CacheInfo.Mux.Lock()
	area.NameArea = CacheInfo.MapArea[CacheInfo.MapRegion[numReg]][numArea]
	CacheInfo.Mux.Unlock()
}
