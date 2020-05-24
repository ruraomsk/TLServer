package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JanFant/newTLServer/internal/model/config"
	"github.com/JanFant/newTLServer/internal/model/locations"
	u "github.com/JanFant/newTLServer/internal/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//Token (JWT) структура токена доступа
type Token struct {
	UserID     int    //Уникальный ID пользователя
	Login      string //Уникальный логин пользователя
	IP         string //IP пользователя
	Role       string //Роль
	Permission []int  //Привелегии
	Region     string //Регион пользователя
	jwt.StandardClaims
}

//Account структура аккаунта пользователя
type Account struct {
	ID       int                `json:"id",sql:"id"`             //уникальный номер пользователя
	Login    string             `json:"login",sql:"login"`       //Имя пользователя
	Password string             `json:"password"`                //Пароль
	BoxPoint locations.BoxPoint `json:"boxPoint",sql:"-"`        //Точки области отображения
	WorkTime time.Duration      `json:"workTime",sql:"workTime"` //Время работы пользователя в часах
	YaMapKey string             `json:"ya_key",sql:"-"`          //Ключ доступа к яндекс карте
	Token    string             `json:"token",sql:"-"`           //Токен пользователя
}

//login обработчик авторизации пользователя в системе
func Login(login, password, ip string) u.Response {
	ipSplit := strings.Split(ip, ":")
	account := &Account{}
	//Забираю из базы запись с подходящей почтой
	rows, err := GetDB().Query(`SELECT id, login, password FROM public.accounts WHERE login=$1`, login)
	if rows == nil {
		return u.Message(http.StatusUnauthorized, fmt.Sprintf("login: %s not found", login))
	}
	if err != nil {
		return u.Message(http.StatusInternalServerError, "Connection to DB error. Please try again")
	}
	for rows.Next() {
		_ = rows.Scan(&account.ID, &account.Login, &account.Password)
		fmt.Println(account)
	}

	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		return u.Message(http.StatusUnauthorized, fmt.Sprintf("Privilege error. login(%s)", login))
	}

	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(http.StatusUnauthorized, fmt.Sprintf("Invalid login credentials. login(%s)", account.Login))
	}
	//Залогинились, создаем токен
	account.Password = ""
	tk := &Token{UserID: account.ID, Login: account.Login, IP: ipSplit[0], Role: privilege.Role.Name, Region: privilege.Region, Permission: privilege.Role.Perm}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = time.Now().Add(time.Hour * account.WorkTime).Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(config.GlobalConfig.TokenPassword))
	account.Token = tokenString
	//сохраняем токен в БД чтобы точно знать что дейтвителен только 1 токен

	_, err = GetDB().Exec(`UPDATE public.accounts SET token = $1 WHERE login = $2`, account.Token, account.Login)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "Connection to DB error. Please try again")
	}

	//Формируем ответ
	resp := u.Message(http.StatusOK, "Logged In")
	resp.Obj["login"] = account.Login
	resp.Obj["token"] = tokenString
	return resp
}

//LogOut выход из учетной записи
func LogOut(mapContx map[string]string) u.Response {
	_, err := GetDB().Exec("UPDATE public.accounts SET token = $1 where login = $2", "", mapContx["login"])
	if err != nil {
		return u.Message(http.StatusInternalServerError, "Connection to DB error. Please try again")
	}
	resp := u.Message(http.StatusOK, "Log out")
	return resp
}

//Validate проверка аккаунда в бд
func (data *Account) Validate1() error {
	err := validation.ValidateStruct(data,
		validation.Field(&data.Login, validation.Required, is.ASCII, validation.Length(4, 100)),
		validation.Field(&data.Password, validation.Required, is.ASCII, validation.Length(6, 100)),
	)
	if err != nil {
		return err
	}
	//логин аккаунта должен быть уникальным
	temp := &Account{}
	rows, err := GetDB().Query(`SELECT id, login, password FROM public.accounts WHERE login=$1`, data.Login)
	if rows != nil && err != nil {
		return errors.New("connection error, please try again")
	}
	if temp.Login != "" {
		return errors.New("login already in use by another user")
	}
	return nil
}

//Validate проверка аккаунда в бд
func (data *Account) Validate() (u.Response, bool) {

	if data.Login != regexp.QuoteMeta(data.Login) {
		return u.Message(http.StatusBadRequest, "login contains invalid characters"), false
	}
	if data.Password != regexp.QuoteMeta(data.Password) {
		return u.Message(http.StatusBadRequest, "Password contains invalid characters"), false
	}
	if len(data.Password) < 6 {
		return u.Message(http.StatusBadRequest, "Password is required"), false
	}
	//логин аккаунта должен быть уникальным
	temp := &Account{}
	rows, err := GetDB().Query(`SELECT id, login, password FROM public.accounts WHERE login=$1`, data.Login)
	if rows != nil && err != nil {
		return u.Message(http.StatusInternalServerError, "Connection error, please try again"), false
	}
	if temp.Login != "" {
		return u.Message(http.StatusBadRequest, "login already in use by another user."), false
	}
	return u.Message(http.StatusOK, "Requirement passed"), true
}

//Create создание аккаунта для пользователей
func (data *Account) Create(privilege Privilege) u.Response {
	if err := data.Validate1(); err != nil {
		return u.Message(http.StatusBadRequest, err.Error())
	}
	//Отдаем ключ для yandex map
	data.YaMapKey = config.GlobalConfig.YaKey
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	data.Password = string(hashedPassword)

	row := GetDB().QueryRow(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4) RETURNING id`,
		data.Login, data.Password, data.WorkTime, data.YaMapKey)
	if err := row.Scan(&data.ID); err != nil {
		return u.Message(http.StatusInternalServerError, err.Error())
	}
	if data.ID <= 0 {
		return u.Message(http.StatusBadRequest, "Failed to create data, connection error.")
	}
	RoleInfo.Mux.Lock()
	privilege.Role.Perm = append(privilege.Role.Perm, RoleInfo.MapRoles[privilege.Role.Name]...)
	RoleInfo.Mux.Unlock()
	if err := privilege.WriteRoleInBD(data.Login); err != nil {
		return u.Message(http.StatusBadRequest, "Connection to DB error. Please try again")
	}
	data.Password = ""
	resp := u.Message(http.StatusOK, "Account has been created")
	resp.Obj["login"] = data.Login
	return resp
}

//Update обновление данных аккаунта
func (data *Account) Update(privilege Privilege) u.Response {
	RoleInfo.Mux.Lock()
	privilege.Role.Perm = append(privilege.Role.Perm, RoleInfo.MapRoles[privilege.Role.Name]...)
	RoleInfo.Mux.Unlock()
	privStr, _ := json.Marshal(privilege)
	_, err := GetDB().Exec(`UPDATE public.accounts SET privilege = $1, work_time = $2 WHERE login = $3`, string(privStr), data.WorkTime, data.Login)
	if err != nil {
		resp := u.Message(http.StatusInternalServerError, fmt.Sprintf("Account update error: %s", err.Error()))
		return resp
	}
	resp := u.Message(http.StatusOK, "Account has updated")
	return resp
}

//Delete удаление аккаунта из БД
func (data *Account) Delete() u.Response {
	_, err := GetDB().Exec(`DELETE FROM public.accounts WHERE login = $3`, data.Login)
	if err != nil {
		resp := u.Message(http.StatusInternalServerError, "data deletion error "+err.Error())
		return resp
	}
	resp := u.Message(http.StatusOK, "account deleted")
	return resp
}

//ChangePW изменение пароля пользователя
func (data *Account) ChangePW() u.Response {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	data.Password = string(hashedPassword)
	//sqlStr := fmt.Sprintf("update public.accounts set password = '%s' where login = '%s';UPDATE public.accounts SET token='' WHERE login='%s'", data.Password, data.Login, data.Login)
	_, err := GetDB().Exec(`UPDATE public.accounts SET password = $1 WHERE login = $2 ; UPDATE public.accounts SET token = '' WHERE login = $3`, data.Password, data.Login, data.Login)
	if err != nil {
		resp := u.Message(http.StatusInternalServerError, "password change error "+err.Error())
		return resp
	}
	resp := u.Message(http.StatusOK, "password changed")
	return resp
}

//ParserBoxPointsUser заполняет BoxPoint
func (data *Account) ParserBoxPointsUser() (err error) {
	var (
		boxpoint  = locations.BoxPoint{}
		privilege = Privilege{}
	)
	err = privilege.ReadFromBD(data.Login)
	if err != nil {
		return errors.New(fmt.Sprintf("ParserPoints. Privilege error: %s", err.Error()))
	}
	var sqlString = `SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross"`
	if !strings.EqualFold(privilege.Region, "*") {
		sqlString = sqlString + fmt.Sprintf(" where region = %s;", privilege.Region)
	}
	row := GetDB().QueryRow(sqlString)
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

	data.BoxPoint = boxpoint
	return nil
}

//GetInfoForUser собор информацию для пользователя, который авторизировался
func (data *Account) GetInfoForUser() u.Response {
	rows, err := GetDB().Query(`SELECT id, login, password FROM public.accounts WHERE login=$1`, data.Login)
	if rows == nil {
		return u.Message(http.StatusUnauthorized, fmt.Sprintf("login: %s not found", data.Login))
	}
	if err != nil {
		return u.Message(http.StatusInternalServerError, "Connection to DB error. Please try again")
	}
	for rows.Next() {
		_ = rows.Scan(&data.ID, &data.Login, &data.Password)
	}

	err = data.ParserBoxPointsUser()
	if err != nil {
		return u.Message(http.StatusInternalServerError, err.Error())
	}
	tflight := GetLightsFromBD(data.BoxPoint)
	resp := u.Message(http.StatusOK, "loading content for the main page")
	resp.Obj["ya_map"] = data.YaMapKey
	resp.Obj["boxPoint"] = data.BoxPoint
	resp.Obj["tflight"] = tflight

	//собираю в кучу регионы для отображения
	chosenRegion := make(map[string]string)
	CacheInfo.Mux.Lock()
	for first, second := range CacheInfo.MapRegion {
		chosenRegion[first] = second
	}
	delete(chosenRegion, "*")
	resp.Obj["regionInfo"] = chosenRegion

	//собираю в кучу районы для отображения
	chosenArea := make(map[string]map[string]string)
	for first, second := range CacheInfo.MapArea {
		chosenArea[first] = make(map[string]string)
		chosenArea[first] = second
	}
	delete(chosenArea, "Все регионы")
	CacheInfo.Mux.Unlock()
	resp.Obj["areaInfo"] = chosenArea

	return resp
}

//SuperCreate создание суперпользователя
func SuperCreate() {
	account := &Account{}
	account.Login = "Super"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 24
	account.Password = "$2a$10$ZCWyIEfEVF3KGj6OUtIeSOQ3WexMjuAZ43VSO6T.QqOndn4HN1J6C"
	//privilege := Privilege{}
	privilege := NewPrivilege("Super", "*", []string{"*"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))

	//!!!!! Другие пользователи Для ОТЛАДКИ потом УДАЛИТЬ все что ниже
	account = &Account{}
	account.Login = "Moscow"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))

	account = &Account{}
	account.Login = "Sachalin"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "3", []string{"1"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))

	account = &Account{}
	account.Login = "Cykotka"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "2", []string{"1", "2", "3"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))

	account = &Account{}
	account.Login = "All"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 1000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Rura"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Alex_B"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 24
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "MMM"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Admin"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "RegAdmin"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "User"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("User", "2", []string{"2"})
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	//roles.GetDB().Table("accounts").Create(data)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Viewer"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = NewPrivilege("Viewer", "3", []string{"1"})
	//roles.GetDB().Table("accounts").Create(data)
	_, _ = GetDB().Exec(`INSERT INTO  public.accounts (login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.YaMapKey)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", data.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	//!!!!! НЕ забудь удалить все что вышел
	fmt.Println("Super created!")
}
