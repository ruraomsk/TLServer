package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	u "github.com/JanFant/TLServer/utils"
	"github.com/pkg/errors"
)

//RoleAccess информация наборах ролей и полномочий
type RoleAccess struct {
	Roles      []Role       `json:"roles"`
	Permission []Permission `json:"permissions"`
}

//Role массив ролей
type Role struct {
	Name string `json:"name"`        //название роли
	Perm []int  `json:"permissions"` //массив полномочий
}

//Privilege brah
type Privilege struct {
	Role   Role     `json:"role"`
	Region string   `json:"region"` //регион пользователя
	Area   []string `json:"area"`   //массив районов пользователя
}

//Permission структура полномойчий содержит ID, команду и описание команды
type Permission struct {
	ID          int    `json:"id"`          //ID порядковый номер
	Command     string `json:"command"`     //название команды
	Visible     bool   `json:"visible"`     //флаг отображения пользователю
	Description string `json:"description"` //описание команды
}

//shortPermission структура полномойчий содержит ID, команду и описание команды урезанный вид для отправки пользователю
type shortPermission struct {
	ID          int    `json:"id"`          //ID порядковый номер
	Description string `json:"description"` //описание команды
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
	sqlStr = fmt.Sprintf("select login, w_time, privilege from public.accounts where login != '%s'", mapContx["login"])
	if !strings.EqualFold(privilege.Region, "*") {
		sqlStr += fmt.Sprintf(`and privilege::jsonb @> '{"region":"%s"}'::jsonb`, privilege.Region)
	}
	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		var tempSA = ShortAccount{}
		err := rowsTL.Scan(&tempSA.Login, &tempSA.Wtime, &tempSA.Privilege)
		if err != nil {
			//logger.Info.Println("DisplayInfoForAdmin: Что-то не так с запросом", err)
			return u.Message(false, "Display info: Bad request")
		}
		var tempPrivilege = Privilege{}
		err = tempPrivilege.ConvertToJson(tempSA.Privilege)
		if err != nil {
			//logger.Info.Println("DisplayInfoForAdmin: Что-то не так со строкой привилегий", err)
			return u.Message(false, "Display info: Privilege json error")
		}
		tempSA.Role.Name = tempPrivilege.Role.Name

		//выбираю привелегии которые не ключены в шаблон роли

		CacheInfo.mux.Lock()
		for _, val1 := range tempPrivilege.Role.Perm {
			flag1, flag2 := false, false
			for _, val2 := range CacheInfo.mapRoles[tempSA.Role.Name] {
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
		CacheInfo.mux.Unlock()

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
	var roles []string
	CacheInfo.mux.Lock()
	defer CacheInfo.mux.Unlock()
	if mapContx["role"] == "Super" {
		roles = append(roles, "Admin")
	} else {
		for roleName, _ := range CacheInfo.mapRoles {
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
	resp["areaInfo"] = chosenArea

	//собираю в кучу разрешения без указания команд
	chosenPermisson := make(map[int]shortPermission)
	for key, value := range CacheInfo.mapPermisson {
		if value.Visible {
			var shValue shortPermission
			shValue.transform(value)
			chosenPermisson[key] = shValue
		}
	}
	resp["permInfo"] = chosenPermisson

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
func (newPrivilege *Privilege) WriteRoleInBD(login string) (err error) {
	privilegeStr, _ := json.Marshal(newPrivilege)
	return GetDB().Exec(fmt.Sprintf("update %s set privilege = '%s' where login = '%s'", GlobalConfig.DBConfig.AccountTable, string(privilegeStr), login)).Error
}

//ReadFromBD прочитать данные из бд и разобрать
func (newPrivilege *Privilege) ReadFromBD(login string) error {
	var privilegeStr string
	sqlStr := fmt.Sprintf("select privilege from %v where login = '%v'", GlobalConfig.DBConfig.AccountTable, login)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&privilegeStr)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(privilegeStr), newPrivilege)
	if err != nil {
		return err
	}
	return nil
}

//ConvertToJson из строки в структуру
func (newPrivilege *Privilege) ConvertToJson(privilegeStr string) (err error) {
	err = json.Unmarshal([]byte(privilegeStr), newPrivilege)
	if err != nil {
		return err
	}
	return nil
}

func NewPrivilegeF(role, region string, area []string) *Privilege {
	var newPrivilege Privilege
	if _, ok := CacheInfo.mapRoles[role]; ok {
		newPrivilege.Role.Name = role
	} else {
		newPrivilege.Role.Name = "Viewer"
	}

	for _, permission := range CacheInfo.mapRoles[newPrivilege.Role.Name] {
		newPrivilege.Role.Perm = append(newPrivilege.Role.Perm, permission)
	}

	if region == "" {
		newPrivilege.Region = "0"
	} else {
		newPrivilege.Region = region
	}

	if len(region) == 0 {
		newPrivilege.Area = []string{"0"}
	} else {
		newPrivilege.Area = area
	}

	return &newPrivilege
}

//RoleCheck проверка полученной роли на соответствие заданной и разрешение на выполнение действия
func NewRoleCheck(mapContx map[string]string, act int) (accept bool, err error) {
	privilege := Privilege{}
	//Проверил соответствует ли роль которую мне дали с ролью установленной в БД
	err = privilege.ReadFromBD(mapContx["login"])
	if err != nil {
		return false, err
	}
	if privilege.Role.Name != mapContx["role"] {
		err = errors.New("Access denied")
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
	a := RoleAccess{}
	_ = a.ReadRoleAccessFile()
	resp["RoleAccess"] = a
	GetDB().Exec(fmt.Sprintf("delete from %v where login = 'TestRole'", GlobalConfig.DBConfig.AccountTable))

	account := &Account{}
	account.Login = "TestRole"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WTime = 69
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege := NewPrivilegeF("Admin", "*", []string{"*"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	priv2 := NewPrivilegeF("", "", []string{""})
	_ = priv2.ReadFromBD("TestRole")
	resp["priv1"] = priv2

	privilege.Role.Perm = append(priv2.Role.Perm, 13, 69)
	_ = privilege.WriteRoleInBD("TestRole")

	priv3 := NewPrivilegeF("", "", []string{""})
	_ = priv3.ReadFromBD("TestRole")
	resp["priv2"] = priv3

	var mapContx = make(map[string]string)
	mapContx["login"] = account.Login
	mapContx["role"] = privilege.Role.Name

	resp["1"], _ = NewRoleCheck(mapContx, 22)
	resp["2"], _ = NewRoleCheck(mapContx, 43)
	resp["3"], _ = NewRoleCheck(mapContx, 1)
	mapContx["login"] = "1"
	resp["4"], _ = NewRoleCheck(mapContx, 1)
	return
}
