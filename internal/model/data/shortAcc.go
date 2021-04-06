package data

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

//ShortAccount удобная структура аккаунта для обмена с пользователем
type ShortAccount struct {
	Login       string     `json:"login"`       //логин пользователя
	WorkTime    int        `json:"workTime"`    //время сеанса пользователя
	Description string     `json:"description"` //описание
	Password    string     `json:"password"`    //пароль пользователя
	Role        Role       `json:"role"`        //роль пользователя
	Privilege   string     `json:"-"`           //привилегии (не уходят на верх)
	Region      RegionInfo `json:"region"`      //регион работы пользователя
	Area        []AreaInfo `json:"area"`        //районы работы пользователя
}

//PassChange структура для изменения пароля
type PassChange struct {
	OldPW string `json:"oldPW"` //старый пароль
	NewPW string `json:"newPW"` //новый пароль
}

//ConvertShortToAcc преобразование информации об аккаунте
func (shortAcc *ShortAccount) ConvertShortToAcc() (account Account, privilege Privilege) {
	account = Account{}
	privilege = Privilege{}
	account.Password = shortAcc.Password
	account.Description = shortAcc.Description
	account.Login = shortAcc.Login
	account.WorkTime = time.Duration(shortAcc.WorkTime)
	privilege.Region = shortAcc.Region.Num
	privilege.Role = shortAcc.Role
	for _, area := range shortAcc.Area {
		privilege.Area = append(privilege.Area, area.Num)
	}
	return account, privilege
}

//ValidCreate проверка данных полученных от пользователя на создание нового пользователя
func (shortAcc *ShortAccount) ValidCreate(role string, region string) (err error) {
	//проверка полученной роли
	RoleInfo.Mux.Lock()
	if _, ok := RoleInfo.MapRoles[shortAcc.Role.Name]; !ok {
		return errors.New("role not found")
	}
	RoleInfo.Mux.Unlock()
	//проверка кто создает
	if role == "RegAdmin" {
		if shortAcc.Role.Name == "Admin" || shortAcc.Role.Name == role {
			return errors.New("this role cannot be created")
		}
		if !strings.EqualFold(shortAcc.Region.Num, region) {
			return errors.New("regions don't match")
		}
	}
	//проверка региона
	//у всех кроме админа регион не равен 0
	if shortAcc.Role.Name != "Admin" {
		if strings.EqualFold(shortAcc.Region.Num, "*") {
			return errors.New("region is incorrect")
		}
	}
	//регион должен существовать
	CacheInfo.Mux.Lock()
	if _, ok := CacheInfo.MapRegion[shortAcc.Region.Num]; !ok {
		return errors.New("region not found")
	}
	//все области для этого региона должны существовать
	for _, area := range shortAcc.Area {
		if _, ok := CacheInfo.MapArea[CacheInfo.MapRegion[shortAcc.Region.Num]][area.Num]; !ok {
			return errors.New("area not found")
		}
	}
	CacheInfo.Mux.Unlock()
	//проверка времени работы
	if shortAcc.WorkTime < 2 {
		return errors.New("Working time should be indicated more than 2 hours")
	}

	return nil
}

//ValidDelete проверка данных полученных от пользователя на удаление аккаунта
func (shortAcc *ShortAccount) ValidDelete(role string, region string) (account *Account, err error) {
	account = &Account{}
	//Забираю из базы запись с подходящей почтой
	db := GetDB("ValidDelete")
	defer FreeDB("ValidDelete")

	rows, err := db.Query(`SELECT login, password, token, work_time FROM public.accounts WHERE login=$1`, shortAcc.Login)
	if rows == nil {
		return nil, errors.New(fmt.Sprintf("login: %s, not found", shortAcc.Login))
	}
	if err != nil {
		return nil, errors.New("connection to DB error")
	}
	for rows.Next() {
		_ = rows.Scan(&account.Login, &account.Password, &account.Token, &account.WorkTime)
	}

	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("privilege error. Login(%s)", account.Login))
	}

	if role == "RegAdmin" {
		if privilege.Role.Name == "Admin" || privilege.Role.Name == role {
			return nil, errors.New("this role cannot be deleted")
		}
		if !strings.EqualFold(privilege.Region, region) {
			return nil, errors.New("regions dn't match")
		}
	}

	return account, nil
}

//ValidChangePW проверка данных полученных от админа для смены паролей пользователя
func (shortAcc *ShortAccount) ValidChangePW(role string, region string) (account *Account, err error) {
	account = &Account{}
	//Забираю из базы запись с подходящим логином
	db := GetDB("ValidChangePW")
	defer FreeDB("ValidChangePW")

	rows, err := db.Query(`SELECT login, password, token, work_time FROM public.accounts WHERE login=$1`, shortAcc.Login)
	if rows == nil {
		return nil, errors.New(fmt.Sprintf("Пользователь: %s, не найден", shortAcc.Login))
	}
	if err != nil {
		return nil, errors.New("Ошибка соединения с базой")
	}
	for rows.Next() {
		_ = rows.Scan(&account.Login, &account.Password, &account.Token, &account.WorkTime)
	}
	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		//logger.Info.Println("Account: Bad privilege")
		return nil, errors.New(fmt.Sprintf("Ошибка чтения полномочий из БД. Пользователь(%s)", account.Login))
	}

	if role == "RegAdmin" {
		if privilege.Role.Name == "Admin" || privilege.Role.Name == role {
			return nil, errors.New("Вы не можете сбросить пароль для этого пользователя")
		}
		if !strings.EqualFold(shortAcc.Region.Num, region) {
			return nil, errors.New("Вы не можете сбросить пароль пользователю другого региона")
		}
	}

	return account, nil
}

//ValidOldNewPW проверка данных полученных от пользователя для изменения своего пароля
func (passChange *PassChange) ValidOldNewPW(login string) (account *Account, err error) {
	account = &Account{}
	//Забираю из базы запись с подходящей почтой
	db := GetDB("ValidOldNewPW")
	defer FreeDB("ValidOldNewPW")

	rows, err := db.Query(`SELECT login, password, token, work_time FROM public.accounts WHERE login=$1`, login)
	if rows == nil {
		return nil, errors.New(fmt.Sprintf("login: %s, not found", login))
	}
	if err != nil {
		return nil, errors.New("connection to DB error")
	}
	for rows.Next() {
		_ = rows.Scan(&account.Login, &account.Password, &account.Token, &account.WorkTime)
	}
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(passChange.OldPW))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, errors.New("invalid login credentials")
	}
	if passChange.NewPW != regexp.QuoteMeta(passChange.NewPW) {
		return nil, errors.New("password contains invalid characters")
	}
	if len(passChange.NewPW) < 6 {
		return nil, errors.New("password is required")
	}
	account.Password = passChange.NewPW

	return account, nil
}
