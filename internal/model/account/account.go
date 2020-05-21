package account

import (
	"fmt"
	"github.com/JanFant/newTLServer/internal/app/config"
	"github.com/JanFant/newTLServer/internal/app/db"
	"github.com/JanFant/newTLServer/internal/model/roles"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"

	u "github.com/JanFant/newTLServer/internal/utils"
	"github.com/dgrijalva/jwt-go"
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
	ID       int    `json:"id",sql:"id"`       //уникальный номер пользователя
	Login    string `json:"login",sql:"login"` //Имя пользователя
	Password string `json:"password"`          //Пароль
	//BoxPoint roles.BoxPoint `json:"boxPoint",sql:"-"`        //Точки области отображения
	WorkTime time.Duration `json:"workTime",sql:"workTime"` //Время работы пользователя в часах
	YaMapKey string        `json:"ya_key",sql:"-"`          //Ключ доступа к яндекс карте
	Token    string        `json:"token",sql:"-"`           //Токен пользователя
}

//login обработчик авторизации пользователя в системе
func login(login, password, ip string) u.Response {
	ipSplit := strings.Split(ip, ":")
	account := &Account{}
	//Забираю из базы запись с подходящей почтой
	rows, err := db.GetDB().Query(`SELECT id, login, password FROM $1 WHERE login=$2`, config.GlobalConfig.DBConfig.AccountTable, login)
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
	privilege := roles.Privilege{}
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

	_, err = db.GetDB().Exec(`UPDATE $1 SET token = $2 WHERE login = $3`, config.GlobalConfig.DBConfig.AccountTable, account.Token, account.Login)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "Connection to DB error. Please try again")
	}

	//Формируем ответ
	resp := u.Message(http.StatusOK, "Logged In")
	resp.Obj["login"] = account.Login
	resp.Obj["token"] = tokenString
	return resp
}

////LogOut выход из учетной записи
//func LogOut(mapContx map[string]string) map[string]interface{} {
//	err := roles.GetDB().Exec("update public.accounts set token = ? where login = ?", "", mapContx["login"]).Error
//	if err != nil {
//		return u.Message(false, "Connection to DB error. Please try again")
//	}
//	resp := u.Message(true, "Log out")
//	return resp
//}
//
////Validate проверка аккаунда в бд
//func (account *Account) Validate() (map[string]interface{}, bool) {
//	if account.Login != regexp.QuoteMeta(account.Login) {
//		return u.Message(false, "login contains invalid characters"), false
//	}
//	if account.Password != regexp.QuoteMeta(account.Password) {
//		return u.Message(false, "Password contains invalid characters"), false
//	}
//	if len(account.Password) < 6 {
//		return u.Message(false, "Password is required"), false
//	}
//	//логин аккаунта должен быть уникальным
//	temp := &Account{}
//	err := roles.GetDB().Table("accounts").Where("login = ?", account.Login).First(temp).Error
//	if err != nil && err != gorm.ErrRecordNotFound {
//		return u.Message(false, "Connection error, please try again"), false
//	}
//	if temp.login != "" {
//		return u.Message(false, "login already in use by another user."), false
//	}
//	return u.Message(false, "Requirement passed"), true
//}
//
////Create создание аккаунта для пользователей
//func (account *Account) Create(privilege roles.Privilege) map[string]interface{} {
//	if resp, ok := account.Validate(); !ok {
//		return resp
//	}
//	//Отдаем ключ для yandex map
//	account.YaMapKey = roles.GlobalConfig.YaKey
//	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
//	account.Password = string(hashedPassword)
//	roles.GetDB().Table("accounts").Create(account)
//	if account.ID <= 0 {
//		return u.Message(false, "Failed to create account, connection error.")
//	}
//	roles.RoleInfo.mux.Lock()
//	privilege.Role.Perm = append(privilege.Role.Perm, roles.RoleInfo.MapRoles[privilege.Role.Name]...)
//	roles.RoleInfo.mux.Unlock()
//	if err := privilege.WriteRoleInBD(account.Login); err != nil {
//		return u.Message(false, "Connection to DB error. Please try again")
//	}
//	account.Password = ""
//	resp := u.Message(true, "Account has been created")
//	resp["login"] = account.Login
//	return resp
//}
//
////Update обновление данных аккаунта
//func (account *Account) Update(privilege roles.Privilege) map[string]interface{} {
//	roles.RoleInfo.mux.Lock()
//	privilege.Role.Perm = append(privilege.Role.Perm, roles.RoleInfo.MapRoles[privilege.Role.Name]...)
//	roles.RoleInfo.mux.Unlock()
//	privStr, _ := json.Marshal(privilege)
//	updateStr := fmt.Sprintf("update public.accounts set privilege = '%s',work_time = %d where login = '%s'", string(privStr), account.WorkTime, account.Login)
//	err := roles.GetDB().Exec(updateStr).Error
//	if err != nil {
//		resp := u.Message(false, fmt.Sprintf("Account update error: %s", err.Error()))
//		return resp
//	}
//	resp := u.Message(true, "Account has updated")
//	return resp
//}
//
////Delete удаление аккаунта из БД
//func (account *Account) Delete() map[string]interface{} {
//	sqlStr := fmt.Sprintf("DELETE FROM public.accounts WHERE login = '%s';", account.Login)
//	err := roles.GetDB().Exec(sqlStr).Error
//	if err != nil {
//		resp := u.Message(true, "account deletion error "+err.Error())
//		return resp
//	}
//	resp := u.Message(true, "Account deleted")
//	return resp
//}
//
////ChangePW изменение пароля пользователя
//func (account *Account) ChangePW() map[string]interface{} {
//	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
//	account.Password = string(hashedPassword)
//	sqlStr := fmt.Sprintf("update public.accounts set password = '%s' where login = '%s';UPDATE public.accounts SET token='' WHERE login='%s'", account.Password, account.Login, account.Login)
//	err := roles.GetDB().Exec(sqlStr).Error
//	if err != nil {
//		resp := u.Message(true, "password change error "+err.Error())
//		return resp
//	}
//	resp := u.Message(true, "Password changed")
//	return resp
//}
//
////ParserBoxPointsUser заполняет BoxPoint
//func (account *Account) ParserBoxPointsUser() (err error) {
//	var (
//		boxpoint  = roles.BoxPoint{}
//		privilege = roles.Privilege{}
//	)
//	err = privilege.ReadFromBD(account.Login)
//	if err != nil {
//		return errors.New(fmt.Sprintf("ParserPoints. Privilege error: %s", err.Error()))
//	}
//	var sqlString = `SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross"`
//	if !strings.EqualFold(privilege.Region, "*") {
//		sqlString = sqlString + fmt.Sprintf(" where region = %s;", privilege.Region)
//	}
//	row := roles.GetDB().Raw(sqlString).Row()
//	err = row.Scan(&boxpoint.Point0.Y, &boxpoint.Point0.X, &boxpoint.Point1.Y, &boxpoint.Point1.X)
//	if err != nil {
//		return errors.New(fmt.Sprintf("ParserPoints. Request error: %s", err.Error()))
//	}
//	if boxpoint.Point0.X > 180 {
//		boxpoint.Point0.X -= 360
//	}
//	if boxpoint.Point1.X > 180 {
//		boxpoint.Point1.X -= 360
//	}
//
//	account.BoxPoint = boxpoint
//	return nil
//}
//
////GetInfoForUser собор информацию для пользователя, который авторизировался
//func (account *Account) GetInfoForUser() map[string]interface{} {
//	err := roles.GetDB().Table("accounts").Where("login = ?", account.Login).First(account).Error
//	if err != nil {
//		if err == gorm.ErrRecordNotFound {
//			//logger.Info.Println("Account: Invalid token, log in again, ", account.Login)
//			return u.Message(false, "Invalid token, log in again")
//		}
//		return u.Message(false, "Connection to DB error. Please log in again")
//	}
//	err = account.ParserBoxPointsUser()
//	if err != nil {
//		return u.Message(false, err.Error())
//	}
//	tflight := roles.GetLightsFromBD(account.BoxPoint)
//	resp := u.Message(true, "Loading content for the main page")
//	resp["ya_map"] = account.YaMapKey
//	resp["boxPoint"] = account.BoxPoint
//	resp["tflight"] = tflight
//
//	//собираю в кучу регионы для отображения
//	chosenRegion := make(map[string]string)
//	roles.CacheInfo.mux.Lock()
//	for first, second := range roles.CacheInfo.mapRegion {
//		chosenRegion[first] = second
//	}
//	delete(chosenRegion, "*")
//	resp["regionInfo"] = chosenRegion
//
//	//собираю в кучу районы для отображения
//	chosenArea := make(map[string]map[string]string)
//	for first, second := range roles.CacheInfo.mapArea {
//		chosenArea[first] = make(map[string]string)
//		chosenArea[first] = second
//	}
//	delete(chosenArea, "Все регионы")
//	roles.CacheInfo.mux.Unlock()
//	resp["areaInfo"] = chosenArea
//
//	return resp
//}
//
//SuperCreate создание суперпользователя
func SuperCreate() {
	account := &Account{}
	account.Login = "Super"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 24
	account.Password = "$2a$10$ZCWyIEfEVF3KGj6OUtIeSOQ3WexMjuAZ43VSO6T.QqOndn4HN1J6C"
	//privilege := Privilege{}
	privilege := roles.NewPrivilege("Super", "*", []string{"*"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	//!!!!! Другие пользователи Для ОТЛАДКИ потом УДАЛИТЬ все что ниже
	account = &Account{}
	account.Login = "Moscow"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Sachalin"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("RegAdmin", "3", []string{"1"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Cykotka"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("RegAdmin", "2", []string{"1", "2", "3"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	_ = privilege.WriteRoleInBD(account.Login)
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "All"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 1000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("Admin", "*", []string{"*"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Rura"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("Admin", "*", []string{"*"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Alex_B"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 24
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("Admin", "*", []string{"*"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "MMM"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("Admin", "*", []string{"*"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Admin"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 10000
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("Admin", "*", []string{"*"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "RegAdmin"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("RegAdmin", "1", []string{"1", "2", "3"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "User"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("User", "2", []string{"2"})
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	//roles.GetDB().Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	account = &Account{}
	account.Login = "Viewer"
	//Отдаем ключ для yandex map
	account.YaMapKey = config.GlobalConfig.YaKey
	account.WorkTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	//privilege = Privilege{}
	privilege = roles.NewPrivilege("Viewer", "3", []string{"1"})
	//roles.GetDB().Table("accounts").Create(account)
	_, _ = db.GetDB().NamedExec(`INSERT INTO  public.accounts (id, login, password, work_time, ya_map_key) VALUES ($1, $2, $3, $4, $5)`, account)
	////Записываю координаты в базу!!!
	//GetDB().Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	_ = privilege.WriteRoleInBD(account.Login)

	//!!!!! НЕ забудь удалить все что вышел
	fmt.Println("Super created!")
}
