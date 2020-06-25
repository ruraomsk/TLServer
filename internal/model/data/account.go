package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/internal/model/locations"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/dgrijalva/jwt-go"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

//Token (JWT) структура токена доступа
type Token struct {
	UserID      int    //Уникальный ID пользователя
	Login       string //Уникальный логин пользователя
	IP          string //IP пользователя
	Description string //какая-то хуйня!!!?!??!?!?!?!?!!?
	Role        string //Роль
	Permission  []int  //Привелегии
	Region      string //Регион пользователя
	jwt.StandardClaims
}

//Account структура аккаунта пользователя
type Account struct {
	ID          int                `json:"id",sql:"id"`                   //уникальный номер пользователя
	Description string             `json:"description",sql:"description"` //какая-то хуйня!!!?!??!?!?!?!?!!?
	Login       string             `json:"login",sql:"login"`             //Имя пользователя
	Password    string             `json:"password"`                      //Пароль
	BoxPoint    locations.BoxPoint `json:"boxPoint",sql:"-"`              //Точки области отображения
	WorkTime    time.Duration      `json:"workTime",sql:"workTime"`       //Время работы пользователя в часах
	YaMapKey    string             `json:"ya_key",sql:"-"`                //Ключ доступа к яндекс карте
	Token       string             `json:"token",sql:"-"`                 //Токен пользователя
}

var AutomaticLogin = "TechAutomatic"
var errorConnectDB = "соединение с БД потеряно"

//login обработчик авторизации пользователя в системе
func Login(login, password, ip string) MapSokResponse {
	ipSplit := strings.Split(ip, ":")
	account := &Account{}
	//Забираю из базы запись с подходящей почтой
	rows, err := GetDB().Query(`SELECT id, login, password, work_time, description FROM public.accounts WHERE login=$1`, login)
	if rows == nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = fmt.Sprintf("Invalid login credentials. login(%s)", login)
		return resp
	}
	if err != nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = "connection to DB error. Please try again"
		return resp
	}
	for rows.Next() {
		_ = rows.Scan(&account.ID, &account.Login, &account.Password, &account.WorkTime, &account.Description)
	}

	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = fmt.Sprintf("Invalid login credentials. login(%s)", login)
		return resp
	}

	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = fmt.Sprintf("Invalid login credentials. login(%s)", account.Login)
		return resp
	}
	//Залогинились, создаем токен
	account.Password = ""
	tk := &Token{UserID: account.ID, Login: account.Login, IP: ipSplit[0], Role: privilege.Role.Name, Region: privilege.Region, Permission: privilege.Role.Perm, Description: account.Description}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = time.Now().Add(time.Hour * account.WorkTime).Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(license.LicenseFields.TokenPass))
	account.Token = tokenString
	//сохраняем токен в БД чтобы точно знать что дейтвителен только 1 токен

	_, err = GetDB().Exec(`UPDATE public.accounts SET token = $1 WHERE login = $2`, account.Token, account.Login)
	if err != nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = "connection to DB error. Please try again"
		return resp
	}

	//Формируем ответ
	resp := newMapMess(typeLogin, nil, nil)
	resp.Data["login"] = account.Login
	resp.Data["token"] = tokenString
	resp.Data["manageFlag"], _ = AccessCheck(login, privilege.Role.Name, 2)
	resp.Data["logDeviceFlag"], _ = AccessCheck(login, privilege.Role.Name, 5)
	resp.Data["authorizedFlag"] = true
	resp.Data["description"] = account.Description
	return resp
}

//LogOut выход из учетной записи
func LogOut(login string) MapSokResponse {
	_, err := GetDB().Exec("UPDATE public.accounts SET token = $1 where login = $2", "", login)
	if err != nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = "connection to DB error. Please try again"
		return resp
	}
	return newMapMess(typeLogOut, nil, nil)
}

//Validate проверка аккаунда в бд
func (data *Account) Validate() error {
	err := validation.ValidateStruct(data,
		validation.Field(&data.Login, validation.Required, is.ASCII, validation.Length(4, 100)),
		validation.Field(&data.Password, validation.Required, is.ASCII, validation.Length(6, 100)),
		validation.Field(&data.Description, validation.Required, is.ASCII, validation.Length(1, 255)),
	)
	if data.Login == "Global" {
		return errors.New("этот логин не может быть создан")
	}
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

//Create создание аккаунта для пользователей
func (data *Account) Create(privilege Privilege) u.Response {
	var count int
	if err := GetDB().QueryRow(`SELECT count(*) FROM public.accounts`).Scan(&count); err != nil {
		return u.Message(http.StatusInternalServerError, errorConnectDB)
	}
	if (count - 1) >= license.LicenseFields.NumAcc {
		return u.Message(http.StatusOK, "ограничение по количеству аккаунтов")
	}
	if err := data.Validate(); err != nil {
		return u.Message(http.StatusBadRequest, errorConnectDB)
	}
	//Отдаем ключ для yandex map
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	data.Password = string(hashedPassword)
	row := GetDB().QueryRow(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4) RETURNING id`,
		data.Login, data.Password, data.WorkTime, data.Description)
	if err := row.Scan(&data.ID); err != nil {
		return u.Message(http.StatusInternalServerError, errorConnectDB)
	}
	if data.ID <= 0 {
		return u.Message(http.StatusBadRequest, "ошибка создания пользователя")
	}
	RoleInfo.Mux.Lock()
	privilege.Role.Perm = append(privilege.Role.Perm, RoleInfo.MapRoles[privilege.Role.Name]...)
	RoleInfo.Mux.Unlock()
	if err := privilege.WriteRoleInBD(data.Login); err != nil {
		return u.Message(http.StatusBadRequest, errorConnectDB)
	}
	data.Password = ""
	resp := u.Message(http.StatusOK, "аккаунт создан")
	resp.Obj["login"] = data.Login
	return resp
}

//Update обновление данных аккаунта
func (data *Account) Update(privilege Privilege) u.Response {
	RoleInfo.Mux.Lock()
	privilege.Role.Perm = append(privilege.Role.Perm, RoleInfo.MapRoles[privilege.Role.Name]...)
	RoleInfo.Mux.Unlock()
	privStr, _ := json.Marshal(privilege)
	_, err := GetDB().Exec(`UPDATE public.accounts SET privilege = $1, work_time = $2, description = $3 WHERE login = $4`, string(privStr), data.WorkTime, data.Description, data.Login)
	if err != nil {
		resp := u.Message(http.StatusInternalServerError, fmt.Sprintf("Account update error: %s", err.Error()))
		return resp
	}
	resp := u.Message(http.StatusOK, "Account has updated")
	return resp
}

//Delete удаление аккаунта из БД
func (data *Account) Delete() u.Response {
	_, err := GetDB().Exec(`DELETE FROM public.accounts WHERE login = $1`, data.Login)
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

	sqlStr := fmt.Sprintf(`UPDATE public.accounts SET password = '%v' WHERE login = '%v'; UPDATE public.accounts SET token = '' WHERE login = '%v'`, data.Password, data.Login, data.Login)
	_, err := GetDB().Exec(sqlStr)
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
		return fmt.Errorf("parserPoints. Privilege error: %s", err.Error())
	}
	var sqlString = `SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross"`
	if !strings.EqualFold(privilege.Region, "*") {
		sqlString = sqlString + fmt.Sprintf(" where region = %s;", privilege.Region)
	}
	row := GetDB().QueryRow(sqlString)
	err = row.Scan(&boxpoint.Point0.Y, &boxpoint.Point0.X, &boxpoint.Point1.Y, &boxpoint.Point1.X)
	if err != nil {
		return fmt.Errorf("parserPoints. Request error: %s", err.Error())
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

//SuperCreate создание суперпользователя
func SuperCreate() {
	account := &Account{}
	account.Login = AutomaticLogin
	//Отдаем ключ для yandex map
	account.WorkTime = 24
	account.Password = "$2a$10$ZCWyIEfEVF3KGj6OUtIeSOQ3WexMjuAZ43VSO6T.QqOndn4HN1J6C"
	account.Description = "Tech"
	privilege := NewPrivilege("Admin", "*", []string{"*"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	//!!!!! Другие пользователи Для ОТЛАДКИ потом УДАЛИТЬ все что ниже
	account = &Account{}
	account.Login = "Moscow"
	//Отдаем ключ для yandex map
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "Moscow ADs"
	privilege = NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Sachalin"
	//Отдаем ключ для yandex map
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "Sachalin"
	privilege = NewPrivilege("RegAdmin", "3", []string{"1"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)

	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Cykotka"
	//Отдаем ключ для yandex map
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "Cykotka"
	privilege = NewPrivilege("RegAdmin", "2", []string{"1", "2", "3"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)

	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "All"
	//Отдаем ключ для yandex map
	account.WorkTime = 1000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "ALLLLLLAAAAALLLLAAA"
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Marina"
	//Отдаем ключ для yandex map
	account.WorkTime = 1000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "Marina ARM"
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Rura"
	//Отдаем ключ для yandex map
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "BoSS"
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Alex_B"
	//Отдаем ключ для yandex map
	account.WorkTime = 24
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "Alex_B Description"
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "MMM"
	//Отдаем ключ для yandex map
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "MMMMMMMMMMMMM"
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Admin"
	//Отдаем ключ для yandex map
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "Admin KEKW"
	privilege = NewPrivilege("Admin", "*", []string{"*"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "RegAdmin"
	//Отдаем ключ для yandex map
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "RegA"
	privilege = NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "User"
	//Отдаем ключ для yandex map
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "User"
	privilege = NewPrivilege("User", "2", []string{"2"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Viewer"
	//Отдаем ключ для yandex map
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	account.Description = "View ASD"
	privilege = NewPrivilege("Viewer", "3", []string{"1"})
	GetDB().MustExec(`INSERT INTO  public.accounts (login, password, work_time, description) VALUES ($1, $2, $3, $4)`,
		account.Login, account.Password, account.WorkTime, account.Description)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)

	//!!!!! НЕ забудь удалить все что вышел
	fmt.Println("Super created!")
}
