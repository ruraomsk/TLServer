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
	Roles      []NewRole    `json:"roles"`
	Permission []Permission `json:"permissions"`
}

//NewRole массив ролей
type NewRole struct {
	Name string `json:"name"`        //название роли
	Perm []int  `json:"permissions"` //массив полномочий
}

//NewPrivilege brah
type NewPrivilege struct {
	NewRole NewRole  `json:"role"`
	Region  string   `json:"region"` //регион пользователя
	Area    []string `json:"area"`   //массив районов пользователя
}

//Roles массив ролей
type Roles struct {
	Roles []Role `json:"roles"`
}

//Role структура содержащая называние роли и ее привелегии
type Role struct {
	Name string      `json:"name"`       //название роли
	Perm Permissions `json:"permission"` //массив полномочий
}

//Permissions массив полномочий
type Permissions struct {
	Permissions []Permission `json:"permissions"`
}

//Permission структура полномойчий содержит ID, команду и описание команды
type Permission struct {
	ID          int    `json:"id"`          //ID порядковый номер
	Command     string `json:"command"`     //название команды
	Description string `json:"description"` //описание команды
}

//Privilege структура привилегий содержит роль, регион, и массив районов
type Privilege struct {
	Role   string   `json:"role"`   //роль пользователя
	Region string   `json:"region"` //регион пользователя
	Area   []string `json:"area"`   //массив районов пользователя
}

//func (roles *Roles) CreateRole() (err error) {
//	var tempRole = new(Role)
//	var tempPermission = new(Permission)
//	tempRole.Name = "Super"
//	tempPermission.ID = 1
//	tempPermission.Command = "CreateUser"
//	tempPermission.Description = "Создание пользователя"
//	tempRole.Perm.Permissions = append(tempRole.Perm.Permissions, *tempPermission)
//
//	roles.Roles = append(roles.Roles, *tempRole)
//	file, _ := json.Marshal(roles)
//	ioutil.WriteFile("test.json", file, os.ModePerm)
//	return err
//}

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
		tempSA.Role = tempPrivilege.Role
		tempSA.Region.SetRegionInfo(tempPrivilege.Region)
		for _, num := range tempPrivilege.Area {
			tempArea := AreaInfo{}
			tempArea.SetAreaInfo(tempSA.Region.Num, num)
			tempSA.Area = append(tempSA.Area, tempArea)
		}
		if tempSA.Role != "Super" {
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
	for first, second := range CacheInfo.mapArea {
		chosenArea[first] = make(map[string]string)
		chosenArea[first] = second
	}
	if mapContx["role"] != "Super" {
		delete(chosenArea, "Все регионы")
	}
	resp["areaInfo"] = chosenArea

	resp["accInfo"] = shortAcc
	return resp
}

//RoleCheck проверка полученной роли на соответствие заданной и разрешение на выполнение действия
func RoleCheck(mapContx map[string]string, act string) (accept bool, err error) {
	privilege := Privilege{}
	//Проверил соответствует ли роль которую мне дали с ролью установленной в БД
	err = privilege.ReadFromBD(mapContx["login"])
	if err != nil {
		return false, err
	}
	if privilege.Role != mapContx["role"] {
		err = errors.New("Access denied")
		return false, err
	}

	CacheInfo.mux.Lock()
	defer CacheInfo.mux.Unlock()
	//Проверяю можно ли делать этой роле данное действие
	for _, perm := range CacheInfo.mapRoles[mapContx["role"]].Permissions {
		if perm.Command == act {
			return true, nil
		}
	}
	err = errors.New("Access denied")
	return false, err
}

//ReadFromBD прочитать данные из бд и разобрать
func (privilege *Privilege) ReadFromBD(login string) error {
	var privilegestr string
	sqlStr := fmt.Sprintf("select privilege from public.accounts where login = '%s'", login)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&privilegestr)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(privilegestr), privilege)
	if err != nil {
		return err
	}
	return nil
}

//ConvertToJson преобразуем структуру в строку для записи в БД
func (privilege *Privilege) ConvertToJson(privilegeStr string) (err error) {
	err = json.Unmarshal([]byte(privilegeStr), privilege)
	if err != nil {
		return err
	}
	return nil
}

//AddPrivilege когдато нужно будет редактировать привелегии наверно...
func (privilege *Privilege) AddPrivilege(privilegeStr, login string) (err error) {
	err = json.Unmarshal([]byte(privilegeStr), privilege)
	if err != nil {
		return err
	}
	return nil
}

//ToSqlStrUpdate запись привелегий в базу
func (privilege *Privilege) ToSqlStrUpdate(table, login string) string {
	privilegeStr, _ := json.Marshal(privilege)
	return fmt.Sprintf("update %s set privilege = '%s' where login = '%s'", table, string(privilegeStr), login)
}

//ReadRoleFile прочитать файл role.json
func (roles *Roles) ReadRoleFile() (err error) {
	file, err := ioutil.ReadFile("./cachefile/Role.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, roles)
	if err != nil {
		return err
	}
	return err
}

//ReadPermissionsFile прочитать файл permissions.json
func (perm *Permissions) ReadPermissionsFile() (err error) {
	file, err := ioutil.ReadFile("./cachefile/Permissions.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, perm)
	if err != nil {
		return err
	}
	return err
}

// ----------------------------------------------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------------------------------------------

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
func (newPrivilege *NewPrivilege) WriteRoleInBD(login string) (err error) {
	privilegeStr, _ := json.Marshal(newPrivilege)
	return GetDB().Exec(fmt.Sprintf("update %s set privilege = '%s' where login = '%s'", GlobalConfig.DBConfig.AccountTable, string(privilegeStr), login)).Error
}

//ReadFromBD прочитать данные из бд и разобрать
func (newPrivilege *NewPrivilege) ReadFromBD(login string) error {
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

func NewPrivilegeF(role, region string, area []string) *NewPrivilege {
	var newPrivilege NewPrivilege
	if _, ok := CacheInfo.mapRoles[role]; ok {
		newPrivilege.NewRole.Name = role
	} else {
		newPrivilege.NewRole.Name = "Viewer"
	}

	temp := CacheInfo.mapRoles[newPrivilege.NewRole.Name]
	for _, permission := range temp.Permissions {
		newPrivilege.NewRole.Perm = append(newPrivilege.NewRole.Perm, permission.ID)
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

	privilege.NewRole.Perm = append(priv2.NewRole.Perm, 21, 22, 25)
	_ = privilege.WriteRoleInBD("TestRole")

	priv3 := NewPrivilegeF("", "", []string{""})
	_ = priv3.ReadFromBD("TestRole")
	resp["priv2"] = priv3

	var mapContx = make(map[string]string)
	mapContx["login"] = account.Login
	mapContx["role"] = privilege.NewRole.Name

	resp["1"], _ = NewRoleCheck(mapContx, 22)
	resp["2"], _ = NewRoleCheck(mapContx, 43)
	resp["3"], _ = NewRoleCheck(mapContx, 1)
	mapContx["login"] = "1"
	resp["4"], _ = NewRoleCheck(mapContx, 1)
	return
}

//RoleCheck проверка полученной роли на соответствие заданной и разрешение на выполнение действия
func NewRoleCheck(mapContx map[string]string, act int) (accept bool, err error) {
	privilege := NewPrivilege{}
	//Проверил соответствует ли роль которую мне дали с ролью установленной в БД
	err = privilege.ReadFromBD(mapContx["login"])
	if err != nil {
		return false, err
	}
	if privilege.NewRole.Name != mapContx["role"] {
		err = errors.New("Access denied")
		return false, err
	}

	//Проверяю можно ли делать этой роле данное действие
	for _, perm := range privilege.NewRole.Perm {
		if perm == act {
			return true, nil
		}
	}
	err = errors.New("Access denied")
	return false, err
}
