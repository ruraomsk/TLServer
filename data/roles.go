package data

import (
	"encoding/json"
	"io/ioutil"
)

type Roles struct {
	Roles []Role
}

type Role struct {
	Name string      `json:"Name"`
	Perm Permissions `json:"Permission"`
}

type Permissions struct {
	Permissions []Permission
}

type Permission struct {
	ID          int    `json:"id"`
	Command     string `json:"command"`
	Description string `json:"description"`
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
