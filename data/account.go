package data

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"../logger"
	u "../utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

//Token JWT claims struct
type Token struct {
	UserID uint   //Уникальный ID пользователя
	Login  string //Уникальный логин пользователя
	IP     string //IP пользователя
	Role   string //Роль
	Region int    //Регион пользователя
	jwt.StandardClaims
}

//Account struct to user account
type Account struct {
	gorm.Model
	Login    string        `json:"login",sql:"login"` //Имя пользователя
	Password string        `json:"password"`          //Пароль
	BoxPoint BoxPoint      `json:"boxpoint",sql:"-"`  //Точки области отображения
	WTime    time.Duration `json:"wtime",sql:"wtime"` //Время работы пользователя в часах
	YaMapKey string        `json:"ya_key",sql:"-"`    //Ключ доступа к ндекс карте
	Token    string        `json:"token",sql:"-"`     //Токен пользователя
}

//Login in system
func Login(login, password, ip string) map[string]interface{} {
	account := &Account{}
	//Забираю из базы запись с подходящей почтой
	err := GetDB().Table("accounts").Where("login = ?", login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info.Println("Account: Login not found: ", login)
			return u.Message(false, "login not found")
		}
		logger.Info.Println("Account: Connection to DB err")
		return u.Message(false, "Connection error. Please try again")
	}

	//Авторизировались добираем полномочия
	privilege := Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		logger.Info.Println("Account: Bad privilege")
		return u.Message(false, "Account: Bad privilege")
	}

	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		logger.Info.Println("Account: Invalid login credentials. Please try again, ", login)
		return u.Message(false, "Invalid login credentials. Please try again")
	}
	//Залогинились, создаем токен
	account.Password = ""
	tk := &Token{UserID: account.ID, Login: account.Login, IP: ip, Role: privilege.Role, Region: privilege.Region}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = time.Now().Add(time.Hour * account.WTime).Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)

	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString
	//сохраняем токен в БД чтобы точно знать что дейтвителен только 1 токен
	GetDB().Exec("update public.accounts set token = ? where login = ?", account.Token, account.Login)

	//Формируем ответ
	resp := u.Message(true, "Logged In")
	resp["role"] = privilege.Role
	resp["login"] = account.Login
	resp["token"] = tokenString
	return resp

}

//Validate checking for an account in the database
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
	//login mast be unique
	temp := &Account{}
	//check for error and duplicate login
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
	account.YaMapKey = os.Getenv("ya_key")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	GetDB().Table("accounts").Create(account)
	if account.ID <= 0 {
		return u.Message(false, "Failed to create account, connection error.")
	}
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))
	account.Password = ""
	resp := u.Message(true, "Account has been created")
	resp["login"] = account.Login
	return resp
}

func (account *Account) Update(privilege Privilege) map[string]interface{} {
	privStr, _ := json.Marshal(privilege)
	updateStr := fmt.Sprintf("update public.accounts set privilege = '%s',w_time = %d where login = '%s'", string(privStr), account.WTime, account.Login)
	err := db.Exec(updateStr).Error
	if err != nil {
		resp := u.Message(true, "account update error "+err.Error())
		return resp
	}
	resp := u.Message(true, "Account has updated")
	return resp
}

func (account *Account) Delete() map[string]interface{} {
	sqlStr := fmt.Sprintf("DELETE FROM public.accounts WHERE login = '%s';", account.Login)
	err := db.Exec(sqlStr).Error
	if err != nil {
		resp := u.Message(true, "account deletion error "+err.Error())
		return resp
	}
	resp := u.Message(true, "Account deleted")
	return resp
}

func (account *Account) ChangePW() map[string]interface{} {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	sqlStr := fmt.Sprintf("update public.accounts set password = '%s' where login = '%s';UPDATE public.accounts SET token='' WHERE login='%s'", account.Password, account.Login, account.Login)
	err := db.Exec(sqlStr).Error
	if err != nil {
		resp := u.Message(true, "password change error "+err.Error())
		return resp
	}
	resp := u.Message(true, "Password changed")
	return resp
}

func (account *Account) DisplayInfoForAdmin(mapContx map[string]string) map[string]interface{} {
	var (
		privilege = Privilege{}
		sqlStr    string
		shortAcc  []ShortAccount
	)
	err := privilege.ReadFromBD(mapContx["login"])
	if err != nil {
		logger.Info.Println("DisplayInfoForAdmin: Не смог считать привилегии пользователя", err)
		return nil
	}
	sqlStr = fmt.Sprintf("select login, w_time, privilege from public.accounts where login != '%s'", mapContx["login"])
	if privilege.Region > 0 {
		sqlStr += fmt.Sprintf("and privilege::jsonb->'region' = '%d'", privilege.Region)
	}

	rowsTL, _ := GetDB().Raw(sqlStr).Rows()
	for rowsTL.Next() {
		var tempSA = ShortAccount{}
		err := rowsTL.Scan(&tempSA.Login, &tempSA.Wtime, &tempSA.Privilege)
		if err != nil {
			logger.Info.Println("DisplayInfoForAdmin: Что-то не так с запросом", err)
			return nil
		}
		var tempPrivilege = Privilege{}
		err = tempPrivilege.ConvertToJson(tempSA.Privilege)
		if err != nil {
			logger.Info.Println("DisplayInfoForAdmin: Что-то не так со строкой привилегий", err)
			return nil
		}
		tempSA.Role = tempPrivilege.Role
		tempSA.Region.SetRegionInfo(tempPrivilege.Region)
		for _, num := range tempPrivilege.Area {
			tempArea := AreaInfo{}
			tempArea.SetAreaInfo(tempSA.Region.Num, num)
			tempSA.Area = append(tempSA.Area, tempArea)
		}

		shortAcc = append(shortAcc, tempSA)
	}

	resp := u.Message(true, "DisplayInfoForAdmin")
	resp["accInfo"] = shortAcc
	resp["regionInfo"] = CacheInfo.mapRegion
	resp["areaInfo"] = CacheInfo.mapArea
	return resp
}

//ParserPointsUser заполняет поля Point в аккаунт
func (account *Account) ParserPointsUser() (err error) {
	var (
		boxpoint  = BoxPoint{}
		privilege = Privilege{}
	)
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		logger.Info.Println("ParserPointsUser: Не смог считать привилегии пользователя", err)
		return err
	}
	if privilege.Region == 0 {
		boxpoint.Point0.SetPoint(42.7961, 25.5637)
		boxpoint.Point1.SetPoint(77.1387, -174.1237)
	} else {
		row := db.Raw(`SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross" where region = ?;`, privilege.Region).Row()
		err := row.Scan(&boxpoint.Point0.Y, &boxpoint.Point0.X, &boxpoint.Point1.Y, &boxpoint.Point1.X)
		if err != nil {
			logger.Info.Println("ParserPointsUser: Что-то не так с запросом", err)
			return err
		}
		if boxpoint.Point0.X > 180 {
			boxpoint.Point0.X -= 360
		}
		if boxpoint.Point1.X > 180 {
			boxpoint.Point0.X -= 360
		}
	}
	account.BoxPoint = boxpoint
	return nil
}

//GetInfoForUser собираю информацию для пользователя который только что залогинился
func (account *Account) GetInfoForUser() map[string]interface{} {
	err := GetDB().Table("accounts").Where("login = ?", account.Login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info.Println("Account: Invalid token, log in again, ", account.Login)
			return u.Message(false, "Invalid token, log in again")
		}
		return u.Message(false, "Connection error. Please log in again")
	}
	account.ParserPointsUser()
	tflight := GetLightsFromBD(account.BoxPoint)
	resp := u.Message(true, "Take this DATA")
	resp["ya_map"] = account.YaMapKey
	resp["boxPoint"] = account.BoxPoint
	resp["tflight"] = tflight
	return resp
}

//SuperCreate создание суперпользователя
func SuperCreate() (err error) {
	account := &Account{}
	account.Login = "Super"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 24
	account.Password = "$2a$10$ZCWyIEfEVF3KGj6OUtIeSOQ3WexMjuAZ43VSO6T.QqOndn4HN1J6C"
	privilege := Privilege{}
	privilege.Role = "Super"
	privilege.Region = 1
	privilege.Area = append(privilege.Area, 1, 2, 3)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	//!!!!! Другие пользователи Для ОТЛАДКИ потом УДАЛИТЬ все что ниже
	account = &Account{}
	account.Login = "Moscow"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "RegAdmin"
	privilege.Region = 1
	privilege.Area = append(privilege.Area, 1, 2, 3)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Sachalin"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "RegAdmin"
	privilege.Region = 3
	privilege.Area = append(privilege.Area, 1)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Cykotka"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "RegAdmin"
	privilege.Region = 2
	privilege.Area = append(privilege.Area, 1, 2)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "All"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "Admin"
	privilege.Region = 0
	privilege.Area = append(privilege.Area, 0)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "MMM"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "Admin"
	privilege.Region = 0
	privilege.Area = append(privilege.Area, 0)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Admin"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "Admin"
	privilege.Region = 0
	privilege.Area = append(privilege.Area, 0)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "RegAdmin"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "RegAdmin"
	privilege.Region = 1
	privilege.Area = append(privilege.Area, 0)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "User"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "User"
	privilege.Region = 2
	privilege.Area = append(privilege.Area, 2)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	account = &Account{}
	account.Login = "Viewer"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 12
	account.Password = "$2a$10$BPvHSsc5VO5zuuZqUFltJeln93d28So27gt81zE0MyAAjnrv8OfaW"
	privilege = Privilege{}
	privilege.Role = "Viewer"
	privilege.Region = 3
	privilege.Area = append(privilege.Area, 1)
	db.Table("accounts").Create(account)
	////Записываю координаты в базу!!!
	db.Exec(privilege.ToSqlStrUpdate("accounts", account.Login))

	//!!!!! НЕ забудь удалить все что вышел
	logger.Info.Println("Super created!")
	fmt.Println("Super created!")
	return err
}
