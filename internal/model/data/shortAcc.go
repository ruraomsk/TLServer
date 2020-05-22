package data

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	u "github.com/JanFant/TLServer/utils"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

//ShortAccount удобная структура аккаунта для обмена с пользователем
type ShortAccount struct {
	Login     string     `json:"login"`    //логин пользователя
	WorkTime  int        `json:"workTime"` //время сеанса пользователя
	Password  string     `json:"password"` //пароль пользователя
	Role      Role       `json:"role"`     //роль пользователя
	Privilege string     `json:"-"`        //привелегии (не уходят на верх)
	Region    RegionInfo `json:"region"`   //регион работы пользователя
	Area      []AreaInfo `json:"area"`     //районы работы пользователя
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
	account.Login = shortAcc.Login
	account.WorkTime = time.Duration(shortAcc.WorkTime)
	privilege.Region = shortAcc.Region.Num
	privilege.Role = shortAcc.Role
	for _, area := range shortAcc.Area {
		privilege.Area = append(privilege.Area, area.Num)
	}
	return account, privilege
}

//DecodeRequest расшифровываем json данные полученные от пользователя
func (shortAcc *ShortAccount) DecodeRequest(w http.ResponseWriter, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(shortAcc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Incorrectly filled data"))
		return err
	}
	return nil
}

//ValidCreate проверка данных полученных от пользователя на создание нового пользователя
func (shortAcc *ShortAccount) ValidCreate(role string, region string) (err error) {
	//проверка полученной роли
	RoleInfo.Mux.Lock()
	if _, ok := RoleInfo.MapRoles[shortAcc.Role.Name]; !ok || shortAcc.Role.Name == "Super" {
		return errors.New("Role not found")
	}
	RoleInfo.Mux.Unlock()
	//проверка кто создает
	if role == "RegAdmin" {
		if shortAcc.Role.Name == "Admin" || shortAcc.Role.Name == role {
			return errors.New("This role cannot be created")
		}
		if !strings.EqualFold(shortAcc.Region.Num, region) {
			return errors.New("Regions don't match")
		}
	}
	//проверка региона
	//у всех кроме админа регион не равен 0
	if shortAcc.Role.Name != "Admin" {
		if strings.EqualFold(shortAcc.Region.Num, "*") {
			return errors.New("Region is incorrect")
		}
	}
	//регион должен существовать
	CacheInfo.Mux.Lock()
	if _, ok := CacheInfo.MapRegion[shortAcc.Region.Num]; !ok {
		return errors.New("Region not found")
	}
	//все области для этого региона должны существовать
	for _, area := range shortAcc.Area {
		if _, ok := CacheInfo.MapArea[CacheInfo.MapRegion[shortAcc.Region.Num]][area.Num]; !ok {
			return errors.New("Area not found")
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
	rows, err := GetDB().Query(`SELECT id, login, password, token, work_time FROM public.accounts WHERE login=$1`, shortAcc.Login)
	if rows == nil {
		return nil, errors.New(fmt.Sprintf("Login: %s, not found", shortAcc.Login))
	}
	if err != nil {
		return nil, errors.New("Connection to DB error")
	}
	for rows.Next() {
		_ = rows.Scan(&account.ID, &account.Login, &account.Password, &account.Token, &account.WorkTime)
	}

	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		//logger.Info.Println("Account: Bad privilege")
		return nil, errors.New(fmt.Sprintf("Privilege error. Login(%s)", account.Login))
	}

	if role == "RegAdmin" {
		if privilege.Role.Name == "Admin" || privilege.Role.Name == role {
			return nil, errors.New("This role cannot be deleted")
		}
		if !strings.EqualFold(privilege.Region, region) {
			return nil, errors.New("Regions dn't match")
		}
	}

	return account, nil
}

//ValidChangePW проверка данных полученных от админа для смены паролей пользователя
func (shortAcc *ShortAccount) ValidChangePW(role string, region string) (account *Account, err error) {
	account = &Account{}
	//Забираю из базы запись с подходящей почтой
	rows, err := GetDB().Query(`SELECT id, login, password, token, work_time FROM public.accounts WHERE login=$2`, shortAcc.Login)
	if rows == nil {
		return nil, errors.New(fmt.Sprintf("Login: %s, not found", shortAcc.Login))
	}
	if err != nil {
		return nil, errors.New("Connection to DB error")
	}
	for rows.Next() {
		_ = rows.Scan(&account.ID, &account.Login, &account.Password, &account.Token, &account.WorkTime)
	}
	account.Password = shortAcc.Password
	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		//logger.Info.Println("Account: Bad privilege")
		return nil, errors.New(fmt.Sprintf("Privilege error. Login(%s)", account.Login))
	}

	if role == "RegAdmin" {
		if privilege.Role.Name == "Admin" || privilege.Role.Name == role {
			return nil, errors.New("Cannot change the password for this user")
		}
		if !strings.EqualFold(shortAcc.Region.Num, region) {
			return nil, errors.New("Regions don't match")
		}
	}

	return account, nil
}

//ValidOldNewPW проверка данных полученных от пользователя для изменения своего пароля
func (passChange *PassChange) ValidOldNewPW(login string) (account *Account, err error) {
	account = &Account{}
	//Забираю из базы запись с подходящей почтой
	rows, err := GetDB().Query(`SELECT id, login, password, token, work_time FROM public.accounts WHERE login=$2`, login)
	if rows == nil {
		return nil, errors.New(fmt.Sprintf("Login: %s, not found", login))
	}
	if err != nil {
		return nil, errors.New("Connection to DB error")
	}
	for rows.Next() {
		_ = rows.Scan(&account.ID, &account.Login, &account.Password, &account.Token, &account.WorkTime)
	}
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(passChange.OldPW))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		//logger.Info.Println("Account: Invalid login credentials. ", login)
		return nil, errors.New("Invalid login credentials")
	}
	if passChange.NewPW != regexp.QuoteMeta(passChange.NewPW) {
		return nil, errors.New("Password contains invalid characters")
	}
	if len(passChange.NewPW) < 6 {
		return nil, errors.New("Password is required")
	}
	account.Password = passChange.NewPW

	return account, nil
}
