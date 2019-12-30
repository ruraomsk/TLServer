package data

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
)

type Roles struct {
	Roles []Role `json:"roles"`
}

type Role struct {
	Name string      `json:"name"`
	Perm Permissions `json:"permission"`
}

type Permissions struct {
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	ID          int    `json:"id"`
	Command     string `json:"command"`
	Description string `json:"description"`
}

type Privilege struct {
	Role   string `json:"role"`
	Region int    `json:"region"`
	Area   []int  `json:"area"`
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
