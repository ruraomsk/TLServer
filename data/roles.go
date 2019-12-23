package data

import (
	"encoding/json"
	"fmt"
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
	Role   Role       `json:"role"`
	Region RegionInfo `json:"region"`
	IDs    []int      `json:"IDs"`
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

func (privilege *Privilege) ReadFromBD() {
	//login := "MMM"

	//GetDB().Table("account").Where("login = ?", login).
}

func (privilege *Privilege) AddPrivilege(privilegeStr, login string) (err error) {
	//sqlstr := fmt.Sprintf("alter table public.accounts add privilege jsonb where login = %s", login)
	//err = GetDB().Exec(sqlstr).Error
	//if err != nil{
	//	return err
	//}
	err = json.Unmarshal([]byte(privilegeStr), privilege)
	if err != nil {
		return err
	}
	return nil
}

func (privilege *Privilege) ToSqlStrUpdate(table, login string) (string) {
	for _,perm := range CacheInfo.mapRoles[privilege.Role.Name].Permissions{
		privilege.Role.Perm.Permissions = append(privilege.Role.Perm.Permissions,perm)
	}
	privilegeStr, _ := json.Marshal(privilege)
	return fmt.Sprintf("update %s set privilege = '%s' where login = '%s'", table, string(privilegeStr), login)
}

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
