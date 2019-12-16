package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
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

//func GetRoles() (Roles map[int]string, err error) {
//	TLsost = make(map[int]string)
//	file, err := ioutil.ReadFile("./cachefile/TLsost.js")
//	if err != nil {
//		return nil, err
//	}
//	temp := new(InfoTL)
//	if err := json.Unmarshal(file, &temp); err != nil {
//		return nil, err
//	}
//	for _, sost := range temp.Sost {
//		if _, ok := TLsost[sost.Num]; !ok {
//			TLsost[sost.Num] = sost.Description
//		}
//	}
//	return TLsost, err
//}
func (roles *Roles) CreateRole() (err error) {
	var temp = new(Role)
	var temp1 = new(Permission)
	temp.Name = "Super"
	temp1.ID = 1
	temp1.Command = "CreateUser"
	temp1.Description = "Создание пользователя"
	temp.Perm.Permissions = append(temp.Perm.Permissions, *temp1)
	temp1.ID = 2
	temp1.Command = "DeleteUser"
	temp1.Description = "Удаление пользователя"
	temp.Perm.Permissions = append(temp.Perm.Permissions, *temp1)
	temp1.ID = 3
	temp1.Command = "UpdateUser"
	temp1.Description = "Обновление учетных данных"
	temp.Perm.Permissions = append(temp.Perm.Permissions, *temp1)
	temp1.ID = 4
	temp1.Command = "ViewCross"
	temp1.Description = "Отображение перекрестков"
	temp.Perm.Permissions = append(temp.Perm.Permissions, *temp1)

	roles.Roles = append(roles.Roles, *temp)
	file, _ := json.Marshal(roles)
	ioutil.WriteFile("test.json", file, os.ModePerm)
	return err
}

func (roles *Roles) ReadRoleFile() (err error) {
	file, err := ioutil.ReadFile("./Role.json")
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
	file, err := ioutil.ReadFile("./Permissions.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(file, perm)
	if err != nil {
		return err
	}
	return err
}
