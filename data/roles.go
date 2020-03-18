package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	u "github.com/JanFant/TLServer/utils"
	"github.com/pkg/errors"
)

var RoleInfo RoleData

type RoleData struct {
	mux          sync.Mutex
	MapRoles     map[string][]int     //роли
	MapPermisson map[int]Permission   //привелегии
	MapRoutes    map[string]RouteInfo //маршруты
}

//RoleAccess информация наборах ролей и полномочий
type RoleAccess struct {
	Roles      []Role       `json:"roles"`
	Permission []Permission `json:"permissions"`
	Routes     []RouteInfo  `json:"routes"`
}

//Role массив ролей
type Role struct {
	Name string `json:"name"`        //название роли
	Perm []int  `json:"permissions"` //массив полномочий
}

//Privilege brah
type Privilege struct {
	Role         Role     `json:"role"`   //информация о роли пользователя
	Region       string   `json:"region"` //регион пользователя
	Area         []string `json:"area"`   //массив районов пользователя
	PrivilegeStr string   `json:"-"`      //строка для декодирования
}

//Permission структура полномойчий содержит ID, команду и описание команды
type Permission struct {
	ID          int    `json:"id"`          //ID порядковый номер
	Visible     bool   `json:"visible"`     //флаг отображения пользователю
	Description string `json:"description"` //описание команды
}

//shortPermission структура полномойчий содержит ID, команду и описание команды урезанный вид для отправки пользователю
type shortPermission struct {
	ID          int    `json:"id"`          //ID порядковый номер
	Description string `json:"description"` //описание команды
}

//RouteInfo информация о всех расписанных маршрутах
type RouteInfo struct {
	ID          int    `json:"id"`
	Permission  int    `json:"permission"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

//DisplayInfoForAdmin отображение информации о пользователях для администраторов
func (privilege *Privilege) DisplayInfoForAdmin(mapContx map[string]string) map[string]interface{} {
	var (
		sqlStr   string
		shortAcc []ShortAccount
	)
	err := privilege.ReadFromBD(mapContx["login"])
	if err != nil {
		//logger.Info.Println("DisplayInfoForAdmin: Не смог считать привилегии пользователя", err)
		return u.Message(false, "Display info: Privilege error")
	}
	sqlStr = fmt.Sprintf("select login, work_time, privilege from public.accounts where login != '%s'", mapContx["login"])
	if !strings.EqualFold(privilege.Region, "*") {
		sqlStr += fmt.Sprintf(`and privilege::jsonb @> '{"region":"%s"}'::jsonb`, privilege.Region)
	}
	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		var tempSA = ShortAccount{}
		err := rowsTL.Scan(&tempSA.Login, &tempSA.WorkTime, &tempSA.Privilege)
		if err != nil {
			//logger.Info.Println("DisplayInfoForAdmin: Что-то не так с запросом", err)
			return u.Message(false, "Display info: Bad request")
		}
		var tempPrivilege = Privilege{}
		tempPrivilege.PrivilegeStr = tempSA.Privilege
		err = tempPrivilege.ConvertToJson()
		if err != nil {
			//logger.Info.Println("DisplayInfoForAdmin: Что-то не так со строкой привилегий", err)
			return u.Message(false, "Display info: Privilege json error")
		}
		tempSA.Role.Name = tempPrivilege.Role.Name

		//выбираю привелегии которые не ключены в шаблон роли

		RoleInfo.mux.Lock()
		for _, val1 := range tempPrivilege.Role.Perm {
			flag1, flag2 := false, false
			for _, val2 := range RoleInfo.MapRoles[tempSA.Role.Name] {
				if val2 == val1 {
					flag1 = true
					break
				}
			}
			for _, val3 := range tempSA.Role.Perm {
				if val3 == val1 {
					flag2 = true
					break
				}
			}
			if !flag1 && !flag2 {
				tempSA.Role.Perm = append(tempSA.Role.Perm, val1)
			}
		}
		RoleInfo.mux.Unlock()

		if tempSA.Role.Perm == nil {
			tempSA.Role.Perm = make([]int, 0)
		}
		tempSA.Region.SetRegionInfo(tempPrivilege.Region)
		for _, num := range tempPrivilege.Area {
			tempArea := AreaInfo{}
			tempArea.SetAreaInfo(tempSA.Region.Num, num)
			tempSA.Area = append(tempSA.Area, tempArea)
		}
		if tempSA.Role.Name != "Super" {
			shortAcc = append(shortAcc, tempSA)
		}
	}

	resp := u.Message(true, "Display information for Admins")

	//собираем в кучу роли
	RoleInfo.mux.Lock()
	var roles []string
	if mapContx["role"] == "Super" {
		roles = append(roles, "Admin")
	} else {
		for roleName, _ := range RoleInfo.MapRoles {
			if roleName != "Super" {
				if (mapContx["role"] == "Admin") && (roleName == "Admin") {
					continue
				}
				if (mapContx["role"] == "RegAdmin") && ((roleName == "Admin") || (roleName == "RegAdmin")) {
					continue
				}
				roles = append(roles, roleName)
			}
		}
	}
	resp["roles"] = roles

	//собираю в кучу разрешения без указания команд
	chosenPermisson := make(map[int]shortPermission)
	for key, value := range RoleInfo.MapPermisson {
		if value.Visible {
			var shValue shortPermission
			shValue.transform(value)
			chosenPermisson[key] = shValue
		}
	}
	resp["permInfo"] = chosenPermisson
	RoleInfo.mux.Unlock()

	CacheInfo.mux.Lock()
	//собираю в кучу регионы для отображения
	chosenRegion := make(map[string]string)
	if mapContx["role"] != "Super" {
		if mapContx["role"] != "RegAdmin" {
			for first, second := range CacheInfo.mapRegion {
				chosenRegion[first] = second
			}
		} else {
			chosenRegion[mapContx["region"]] = CacheInfo.mapRegion[mapContx["region"]]
		}
		delete(chosenRegion, "*")
	} else {
		chosenRegion["*"] = CacheInfo.mapRegion["*"]
	}
	resp["regionInfo"] = chosenRegion

	//собираю в кучу районы для отображения
	chosenArea := make(map[string]map[string]string)
	for key, value := range CacheInfo.mapArea {
		chosenArea[key] = make(map[string]string)
		chosenArea[key] = value
	}
	if mapContx["role"] != "Super" {
		delete(chosenArea, "Все регионы")
	}
	CacheInfo.mux.Unlock()
	resp["areaInfo"] = chosenArea

	resp["accInfo"] = shortAcc
	return resp
}

func (shPerm *shortPermission) transform(perm Permission) {
	shPerm.Description = perm.Description
	shPerm.ID = perm.ID
}

func (roleAccess *RoleAccess) ReadRoleAccessFile() (err error) {
	file, err := ioutil.ReadFile(GlobalConfig.CachePath + "//RoleAccess.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, roleAccess)
	if err != nil {
		return err
	}
	return err
}

//ToSqlStrUpdate запись привелегий в базу
func (privilege *Privilege) WriteRoleInBD(login string) (err error) {
	privilegeStr, _ := json.Marshal(privilege)
	return GetDB().Exec(fmt.Sprintf("update %s set privilege = '%s' where login = '%s'", GlobalConfig.DBConfig.AccountTable, string(privilegeStr), login)).Error
}

//ReadFromBD прочитать данные из бд и разобрать
func (privilege *Privilege) ReadFromBD(login string) error {
	var privilegeStr string
	sqlStr := fmt.Sprintf("select privilege from %v where login = '%v'", GlobalConfig.DBConfig.AccountTable, login)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&privilegeStr)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(privilegeStr), privilege)
	if err != nil {
		return err
	}
	return nil
}

//ConvertToJson из строки в структуру
func (privilege *Privilege) ConvertToJson() (err error) {
	err = json.Unmarshal([]byte(privilege.PrivilegeStr), privilege)
	if err != nil {
		return err
	}
	return nil
}

func NewPrivilege(role, region string, area []string) *Privilege {
	var privilege Privilege
	RoleInfo.mux.Lock()
	if _, ok := RoleInfo.MapRoles[role]; ok {
		privilege.Role.Name = role
	} else {
		privilege.Role.Name = "Viewer"
	}

	for _, permission := range RoleInfo.MapRoles[privilege.Role.Name] {
		privilege.Role.Perm = append(privilege.Role.Perm, permission)
	}
	RoleInfo.mux.Unlock()
	if region == "" {
		privilege.Region = "0"
	} else {
		privilege.Region = region
	}

	if len(region) == 0 {
		privilege.Area = []string{"0"}
	} else {
		privilege.Area = area
	}

	return &privilege
}

//RoleCheck проверка полученной роли на соответствие заданной и разрешение на выполнение действия
func AccessCheck(login string, act int) (accept bool, err error) {
	privilege := Privilege{}
	//Проверил соответствует ли роль которую мне дали с ролью установленной в БД
	err = privilege.ReadFromBD(login)
	if err != nil {
		return false, err
	}

	//Проверяю можно ли делать этой роле данное действие
	for _, perm := range privilege.Role.Perm {
		if perm == act {
			return true, nil
		}
	}
	err = errors.New("Access denied")
	return false, err
}

func TestNewRoleSystem() (resp map[string]interface{}) {
	resp = make(map[string]interface{})
	resp["1"] = RoleInfo.MapRoles
	resp["2"] = RoleInfo.MapPermisson
	resp["3"] = RoleInfo.MapRoutes

	//a := RoleAccess{}
	//_ = a.ReadRoleAccessFile()
	//resp["RoleAccess"] = a
	//GetDB().Exec(fmt.Sprintf("delete from %v where login = 'TestRole'", GlobalConfig.DBConfig.AccountTable))
	//
	//account := &Account{}
	//account.Login = "TestRole"
	////Отдаем ключ для yandex map
	//account.YaMapKey = GlobalConfig.YaKey
	//account.WorkTime = 69
	//account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege := NewPrivilege("Admin", "*", []string{"*"})
	//GetDB().Table("accounts").Create(account)
	//////Записываю координаты в базу!!!
	//_ = privilege.WriteRoleInBD(account.Login)
	//
	//priv2 := NewPrivilege("", "", []string{""})
	//_ = priv2.ReadFromBD("TestRole")
	//resp["priv1"] = priv2
	//
	//privilege.Role.Perm = append(priv2.Role.Perm, 13, 69)
	//_ = privilege.WriteRoleInBD("TestRole")
	//
	//priv3 := NewPrivilege("", "", []string{""})
	//_ = priv3.ReadFromBD("TestRole")
	//resp["priv2"] = priv3
	//
	//var mapContx = make(map[string]string)
	//mapContx["login"] = account.Login
	//mapContx["role"] = privilege.Role.Name
	//
	//resp["1"], _ = NewRoleCheck(mapContx, 22)
	//resp["2"], _ = NewRoleCheck(mapContx, 43)
	//resp["3"], _ = NewRoleCheck(mapContx, 1)
	//mapContx["login"] = "1"
	//resp["4"], _ = NewRoleCheck(mapContx, 1)
	return
}
