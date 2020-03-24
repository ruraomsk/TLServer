package data

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	u "github.com/JanFant/TLServer/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

//Token (JWT) структура токена доступа
type Token struct {
	UserID     uint   //Уникальный ID пользователя
	Login      string //Уникальный логин пользователя
	IP         string //IP пользователя
	Role       string //Роль
	Permission []int  //Привелегии
	Region     string //Регион пользователя
	jwt.StandardClaims
}

//Account структура для аккаунтов пользователя
type Account struct {
	gorm.Model
	Login    string        `json:"login",sql:"login"`       //Имя пользователя
	Password string        `json:"password"`                //Пароль
	BoxPoint BoxPoint      `json:"boxPoint",sql:"-"`        //Точки области отображения
	WorkTime time.Duration `json:"workTime",sql:"workTime"` //Время работы пользователя в часах
	YaMapKey string        `json:"ya_key",sql:"-"`          //Ключ доступа к яндекс карте
	Token    string        `json:"token",sql:"-"`           //Токен пользователя
}

//Login обработчик авторизации пользователя в системе
func Login(login, password, ip string) map[string]interface{} {
	ipSplit := strings.Split(ip, ":")
	account := &Account{}
	//Забираю из базы запись с подходящей почтой
	err := GetDB().Table("accounts").Where("login = ?", login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//logger.Warning.Println("IP: " + ip + " Login: " + login + " Message: " + "Login not found")
			return u.Message(false, fmt.Sprintf("Login: %s not found", login))
		}
		return u.Message(false, "Connection to DB error. Please try again")
	}

	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		// logger.Error.Println("IP: " + ip + " Login: " + login + " Message: " + "Privilege error")
		return u.Message(false, fmt.Sprintf("Privilege error. Login(%s)", login))
	}

	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(false, fmt.Sprintf("Invalid login credentials. Login(%s)", account.Login))
	}
	//Залогинились, создаем токен
	account.Password = ""
	tk := &Token{UserID: account.ID, Login: account.Login, IP: ipSplit[0], Role: privilege.Role.Name, Region: privilege.Region, Permission: privilege.Role.Perm}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = time.Now().Add(time.Hour * account.WorkTime).Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(GlobalConfig.TokenPassword))
	account.Token = tokenString
	//сохраняем токен в БД чтобы точно знать что дейтвителен только 1 токен
	GetDB().Exec("update public.accounts set token = ? where login = ?", account.Token, account.Login)

	//Формируем ответ
	resp := u.Message(true, "Logged In")
	resp["login"] = account.Login
	resp["token"] = tokenString
	return resp
}

//LogOut выход из учетной записи
func LogOut(mapContx map[string]string) map[string]interface{} {
	err := GetDB().Exec("update public.accounts set token = ? where login = ?", "", mapContx["login"]).Error
	if err != nil {
		return u.Message(false, "Connection to DB error. Please try again")
	}
	resp := u.Message(true, "Log out")
	return resp
}

//Validate проверка аккаунда в бд
func (account *Account) Validate() (map[string]interface{}, bool) {
	if account.Login != regexp.QuoteMeta(account.Login) {
		return u.Message(false, "Login contains invalid characters"), false
	}
	if account.Password != regexp.QuoteMeta(account.Password) {
		return u.Message(false, "Password contains invalid characters"), false
	}
	if len(account.Password) < 6 {
		return u.Message(false, "Password is required"), false
	}
	//логин аккаунта должен быть уникальным
	temp := &Account{}
	err := GetDB().Table("accounts").Where("login = ?", account.Login).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error, please try again"), false
	}
	if temp.Login != "" {
		return u.Message(false, "Login already in use by another user."), false
	}
	return u.Message(false, "Requirement passed"), true
}

//Create создание аккаунта для пользователей
func (account *Account) Create(privilege Privilege) map[string]interface{} {
	if resp, ok := account.Validate(); !ok {
		return resp
	}
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	GetDB().Table("accounts").Create(account)
	if account.ID <= 0 {
		return u.Message(false, "Failed to create account, connection error.")
	}
	RoleInfo.mux.Lock()
	privilege.Role.Perm = append(privilege.Role.Perm, RoleInfo.MapRoles[privilege.Role.Name]...)
	RoleInfo.mux.Unlock()
	if err := privilege.WriteRoleInBD(account.Login); err != nil {
		return u.Message(false, "Connection to DB error. Please try again")
	}
	account.Password = ""
	resp := u.Message(true, "Account has been created")
	resp["login"] = account.Login
	return resp
}

//Update обновление данных аккаунта
func (account *Account) Update(privilege Privilege) map[string]interface{} {
	RoleInfo.mux.Lock()
	privilege.Role.Perm = append(privilege.Role.Perm, RoleInfo.MapRoles[privilege.Role.Name]...)
	RoleInfo.mux.Unlock()
	privStr, _ := json.Marshal(privilege)
	updateStr := fmt.Sprintf("update public.accounts set privilege = '%s',work_time = %d where login = '%s'", string(privStr), account.WorkTime, account.Login)
	err := GetDB().Exec(updateStr).Error
	if err != nil {
		resp := u.Message(false, fmt.Sprintf("Account update error: %s", err.Error()))
		return resp
	}
	resp := u.Message(true, "Account has updated")
	return resp
}

//Delete удаление аккаунта из БД
func (account *Account) Delete() map[string]interface{} {
	sqlStr := fmt.Sprintf("DELETE FROM public.accounts WHERE login = '%s';", account.Login)
	err := GetDB().Exec(sqlStr).Error
	if err != nil {
		resp := u.Message(true, "account deletion error "+err.Error())
		return resp
	}
	resp := u.Message(true, "Account deleted")
	return resp
}

//ChangePW изменение пароля пользователя
func (account *Account) ChangePW() map[string]interface{} {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	sqlStr := fmt.Sprintf("update public.accounts set password = '%s' where login = '%s';UPDATE public.accounts SET token='' WHERE login='%s'", account.Password, account.Login, account.Login)
	err := GetDB().Exec(sqlStr).Error
	if err != nil {
		resp := u.Message(true, "password change error "+err.Error())
		return resp
	}
	resp := u.Message(true, "Password changed")
	return resp
}

//ParserBoxPointsUser заполняет BoxPoint
func (account *Account) ParserBoxPointsUser() (err error) {
	var (
		boxpoint  = BoxPoint{}
		privilege = Privilege{}
	)
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		return errors.New(fmt.Sprintf("ParserPoints. Privilege error: %s", err.Error()))
	}
	var sqlString = `SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross"`
	if !strings.EqualFold(privilege.Region, "*") {
		sqlString = sqlString + fmt.Sprintf(" where region = %s;", privilege.Region)
	}
	row := GetDB().Raw(sqlString).Row()
	err = row.Scan(&boxpoint.Point0.Y, &boxpoint.Point0.X, &boxpoint.Point1.Y, &boxpoint.Point1.X)
	if err != nil {
		return errors.New(fmt.Sprintf("ParserPoints. Request error: %s", err.Error()))
	}
	if boxpoint.Point0.X > 180 {
		boxpoint.Point0.X -= 360
	}
	if boxpoint.Point1.X > 180 {
		boxpoint.Point1.X -= 360
	}

	account.BoxPoint = boxpoint
	return nil
}

//GetInfoForUser собираю информацию для пользователя который только что залогинился
func (account *Account) GetInfoForUser() map[string]interface{} {
	err := GetDB().Table("accounts").Where("login = ?", account.Login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			//logger.Info.Println("Account: Invalid token, log in again, ", account.Login)
			return u.Message(false, "Invalid token, log in again")
		}
		return u.Message(false, "Connection to DB error. Please log in again")
	}
	err = account.ParserBoxPointsUser()
	if err != nil {
		return u.Message(false, err.Error())
	}
	tflight := GetLightsFromBD(account.BoxPoint)
	resp := u.Message(true, "Loading content for the main page")
	resp["ya_map"] = account.YaMapKey
	resp["boxPoint"] = account.BoxPoint
	resp["tflight"] = tflight

	//собираю в кучу регионы для отображения
	chosenRegion := make(map[string]string)
	CacheInfo.mux.Lock()
	for first, second := range CacheInfo.mapRegion {
		chosenRegion[first] = second
	}
	delete(chosenRegion, "*")
	resp["regionInfo"] = chosenRegion

	//собираю в кучу районы для отображения
	chosenArea := make(map[string]map[string]string)
	for first, second := range CacheInfo.mapArea {
		chosenArea[first] = make(map[string]string)
		chosenArea[first] = second
	}
	delete(chosenArea, "Все регионы")
	CacheInfo.mux.Unlock()
	resp["areaInfo"] = chosenArea

	return resp
}

//SuperCreate создание суперпользователя
func SuperCreate() (err error) {
	account := &Account{}
	account.Login = "Super"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 24
	account.Password = "$2a$10$ZCWyIEfEVF3KGj6OUtIeSOQ3WexMjuAZ43VSO6T.QqOndn4HN1J6C"
	//privilege := Privilege{}
	privilege := NewPrivilege("Super", "*", []string{"*"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	//!!!!! Другие пользователи Для ОТЛАДКИ потом УДАЛИТЬ все что ниже
	account = &Account{}
	account.Login = "Moscow"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Sachalin"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "3", []string{"1"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Cykotka"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "2", []string{"1", "2", "3"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "All"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 1000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Rura"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "MMM"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Admin"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "RegAdmin"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "User"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("User", "2", []string{"2"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Viewer"
	//Отдаем ключ для yandex map
	account.YaMapKey = GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Viewer", "3", []string{"1"})
	GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	//!!!!! НЕ забудь удалить все что вышел
	fmt.Println("Super created!")
	return err
}
