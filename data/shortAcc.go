package data

import (
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
