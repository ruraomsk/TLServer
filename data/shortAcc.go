package data

import (
	"fmt"
	"strconv"
	"time"
)

type ShortAccount struct {
	Login     string     `json:"login"`
	Wtime     int        `json:"wtime"`
	Password  string     `json:"password"`
	Role      string     `json:"role"`
	Privilege string     `json:"-"`
	Region    RegionInfo `json:"region"`
	Area      []AreaInfo `json:"area"`
}

func (shortAcc *ShortAccount) ConvertShortToAcc() (account Account, privilege Privilege) {
	account = Account{}
	privilege = Privilege{}
	account.Password = shortAcc.Password
	account.Login = shortAcc.Login
	account.WTime = time.Duration(shortAcc.Wtime)
	privilege.Region = shortAcc.Region.Num
	privilege.Role = shortAcc.Role
	for _, area := range shortAcc.Area {
		privilege.Area = append(privilege.Area, area.Num)
	}
	return account, privilege
}

func (shortAcc *ShortAccount) ValidCreate(role string, region string) (err error) {
	//проверка полученной роли
	if _, ok := CacheInfo.mapRoles[shortAcc.Role]; !ok || shortAcc.Role == "Super" {
		return fmt.Errorf("role not found")
	}
	//проверка кто создает
	if role == "RegAdmin" {
		if shortAcc.Role == "Admin" {
			return fmt.Errorf("this role cannot be created")
		}
		if num, _ := strconv.Atoi(region); shortAcc.Region.Num != num {
			return fmt.Errorf("regions dn't match")
		}
	}
	//проверка региона
	//у всех кроме админа регион не равен 0
	if shortAcc.Role != "Admin" {
		if shortAcc.Region.Num == 0 {
			return fmt.Errorf("region is incorrect")
		}
	}
	//регион должен существовать
	if _, ok := CacheInfo.mapRegion[shortAcc.Region.Num]; !ok {
		return fmt.Errorf("region not found")
	}
	//все области для этого региона должны существовать
	for _, area := range shortAcc.Area {
		if _, ok := CacheInfo.mapArea[CacheInfo.mapRegion[shortAcc.Region.Num]][area.Num]; !ok {
			return fmt.Errorf("area not found")
		}
	}
	//проверка времени работы
	if shortAcc.Wtime < 2 {
		return fmt.Errorf("working time should be indicated more than 2 hours")
	}

	return nil
}
