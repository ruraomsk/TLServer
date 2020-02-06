package data

import (
	"../logger"
	u "../utils"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//ShortAccount удобная структура аккаунта для обмена с пользователем
type ShortAccount struct {
	Login     string     `json:"login"`
	Wtime     int        `json:"wtime"`
	Password  string     `json:"password"`
	Role      string     `json:"role"`
	Privilege string     `json:"-"`
	Region    RegionInfo `json:"region"`
	Area      []AreaInfo `json:"area"`
}

//PassChange структура для изменения пароля
type PassChange struct {
	OldPW string `json:"oldPW"`
	NewPW string `json:"newPW"`
}

//ConvertShortToAcc преобразование 2х структур информации об аккаунте в одну целую
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

//DecodeRequest расшифровываем json данные полученные от пользователя
func (shortAcc *ShortAccount) DecodeRequest(w http.ResponseWriter, r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(shortAcc)
	if err != nil {
		//logger.Info.Println("ActParser, Add: Incorrectly filled data ", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, r, u.Message(false, "Incorrectly filled data"))
		return err
	}
	return nil
}

//ValidCreate проверка данных полученных от пользователя на создание нового пользователя
func (shortAcc *ShortAccount) ValidCreate(role string, region string) (err error) {
	//проверка полученной роли
	if _, ok := CacheInfo.mapRoles[shortAcc.Role]; !ok || shortAcc.Role == "Super" {
		return errors.New("Role not found")
	}
	//проверка кто создает
	if role == "RegAdmin" {
		if shortAcc.Role == "Admin" || shortAcc.Role == role {
			return errors.New("This role cannot be created")
		}
		if !strings.EqualFold(shortAcc.Region.Num, region) {
			return errors.New("Regions don't match")
		}
	}
	//проверка региона
	//у всех кроме админа регион не равен 0
	if shortAcc.Role != "Admin" {
		if strings.EqualFold(shortAcc.Region.Num, "*") {
			return errors.New("Region is incorrect")
		}
	}
	//регион должен существовать
	if _, ok := CacheInfo.mapRegion[shortAcc.Region.Num]; !ok {
		return errors.New("Region not found")
	}
	//все области для этого региона должны существовать
	for _, area := range shortAcc.Area {
		if _, ok := CacheInfo.mapArea[CacheInfo.mapRegion[shortAcc.Region.Num]][area.Num]; !ok {
			return errors.New("Area not found")
		}
	}
	//проверка времени работы
	if shortAcc.Wtime < 2 {
		return errors.New("Working time should be indicated more than 2 hours")
	}

	return nil
}

//ValidDelete проверка данных полученных от пользователя на удаление аккаунта
func (shortAcc *ShortAccount) ValidDelete(role string, region string) (account *Account, err error) {
	account = &Account{}
	//Забираю из базы запись с подходящей почтой
	err = GetDB().Table("accounts").Where("login = ?", shortAcc.Login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//logger.Info.Println("Account: Login not found: ", shortAcc.Login)
			return nil, errors.New(fmt.Sprintf("Login: %s, not found", shortAcc.Login))
		}
		//logger.Info.Println("Account: Connection to DB err")
		return nil, errors.New("Connection to DB error")
	}

	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		//logger.Info.Println("Account: Bad privilege")
		return nil, errors.New(fmt.Sprintf("Privilege error. Login(%s)", account.Login))
	}

	if role == "RegAdmin" {
		if privilege.Role == "Admin" || privilege.Role == role {
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
	err = GetDB().Table("accounts").Where("login = ?", shortAcc.Login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//logger.Info.Println("Account: Login not found: ", shortAcc.Login)
			return nil, errors.New("Login not found")
		}
		//logger.Info.Println("Account: Connection to DB err")
		return nil, errors.New("Connection to DB error")
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
		if privilege.Role == "Admin" || privilege.Role == role {
			return nil, errors.New("Cannot change the password for this user")
		}
		if !strings.EqualFold(shortAcc.Region.Num, region) {
			return nil, errors.New("Regions don't match")
		}
	}

	return account, nil
}

//ValidOldNewPW проверка данных полученных от пользователя из изменения своего пароля
func (passChange *PassChange) ValidOldNewPW(login string) (account *Account, err error) {
	account = &Account{}
	//Забираю из базы запись с подходящей почтой
	err = GetDB().Table("accounts").Where("login = ?", login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//logger.Info.Println("Account: Login not found: ", login)
			return nil, errors.New("Login not found")
		}
		logger.Error.Println("|Message: Connection to DB err")
		return nil, errors.New("Connection to DB error")
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
