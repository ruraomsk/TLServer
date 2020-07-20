package data

import (
	"fmt"
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
	MapTLSost map[int]string               //светофоры

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
	FillMapAreaBox()
	err = getRoleAccess()
	if err != nil {
		logger.Error.Println(fmt.Sprintf("|Message: Error reading data cache: %s", err.Error()))
	}
}

//FillMapAreaBox заполнение мапы районов и регионов с координатами
func FillMapAreaBox() {
	FillCacheALB()
	CacheArea.Mux.Lock()
	defer CacheArea.Mux.Unlock()
	CacheInfo.Mux.Lock()
	defer CacheInfo.Mux.Unlock()
	CacheArea.Areas = make([]locations.AreaBox, 0)
	//запрос уникальных регионов и районов
	rows, _ := GetDB().Query("SELECT distinct on (region, area) region, area, Min(dgis[0]) as \"Y0\", Min(convTo360(dgis[1])) as \"X0\", Max(dgis[0]) as \"Y1\", Max(convTo360(dgis[1])) as \"X1\"  FROM public.\"cross\"  group by region, area")
	for rows.Next() {
		var temp locations.AreaBox
		_ = rows.Scan(&temp.Region, &temp.Area, &temp.Box.Point0.Y, &temp.Box.Point0.X, &temp.Box.Point1.Y, &temp.Box.Point1.X)
		temp.Sub = make([]locations.SybAreaBox, 0)
		CacheArea.Areas = append(CacheArea.Areas, temp)
	}

	//запрос уникальных подрайонов
	rows, _ = GetDB().Query("SELECT distinct on (region, area, subarea) region, area, subarea, Min(dgis[0]) as \"Y0\", Min(convTo360(dgis[1])) as \"X0\", Max(dgis[0]) as \"Y1\", Max(convTo360(dgis[1])) as \"X1\"  FROM public.\"cross\"  group by region, area, subarea")
	for rows.Next() {
		var (
			tempSyb locations.SybAreaBox
			reg     string
			area    string
		)
		_ = rows.Scan(&reg, &area, &tempSyb.SubArea, &tempSyb.Box.Point0.Y, &tempSyb.Box.Point0.X, &tempSyb.Box.Point1.Y, &tempSyb.Box.Point1.X)
		for num, areaBox := range CacheArea.Areas {
			if areaBox.Region == reg && areaBox.Area == area {
				CacheArea.Areas[num].Sub = append(CacheArea.Areas[num].Sub, tempSyb)
			}
		}

	}
	//заполним поля названиями
	for num := range CacheArea.Areas {
		CacheArea.Areas[num].Region = CacheInfo.MapRegion[CacheArea.Areas[num].Region]
		CacheArea.Areas[num].Area = CacheInfo.MapArea[CacheArea.Areas[num].Region][CacheArea.Areas[num].Area]
	}
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
	statusRow, err := GetDB().Query(`SELECT id, description FROM public.status`)
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
