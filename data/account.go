package data

import (
	"../logger"
	u "../utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

//Token JWT claims struct
type Token struct {
	UserID uint   //Уникальный ID пользователя
	Login  string //Уникальный логин пользователя
	IP     string //IP пользователя
	jwt.StandardClaims
}

//Account struct to user account
type Account struct {
	gorm.Model
	Login     string        `json:"login",sql:"login"` //Имя пользователя
	Password  string        `json:"password"`          //Пароль
	BoxPoint  BoxPoint      `json:"boxpoint",sql:"-"`  //Точки области отображения
	WTime     time.Duration `json:"wtime",sql:"wtime"` //Время работы пользователя в часах
	YaMapKey  string        `json:"ya_key",sql:"-"`    //Ключ доступа к ндекс карте
	Token     string        `json:"token",sql:"-"`     //Токен пользователя
	//Privilege Privilege     `json:"privilege",sql:"-"`
}

//Login in system
func Login(login, password, ip string) map[string]interface{} {
	account := &Account{}
	//privilege := Privilege{}
	//Забираю из базы запись с подходящей почтой

	//sqlStr := fmt.Sprintf("select id, login, password, w_time, ya_map_key from %s ", os.Getenv("gis_table"))



	err := GetDB().Table("accounts").Where("login = ?", login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Info.Println("Account: Login not found: ", login)
			return u.Message(false, "login not found")
		}
		logger.Info.Println("Account: Connection to DB err")
		return u.Message(false, "Connection error. Please try again")
	}


	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		logger.Info.Println("Account: Invalid login credentials. Please try again, ", login)
		return u.Message(false, "Invalid login credentials. Please try again")
	}
	//Залогинились, создаем токен
	account.Password = ""
	tk := &Token{UserID: account.ID, Login: account.Login, IP: ip}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = time.Now().Add(time.Hour * account.WTime).Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)

	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString
	account.ParserPointsUser()
	//сохраняем токен в БД чтобы точно знать что дейтвителен только 1 токен
	GetDB().Exec("update accounts set token = ? where login = ?", account.Token, account.Login)

	//Формируем ответ
	resp := u.Message(true, "Logged In")
	resp["login"] = account.Login
	resp["token"] = tokenString
	return resp

}

//Validate checking for an account in the database
func (account *Account) Validate() (map[string]interface{}, bool) {
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
		return u.Message(false, "login already in use by another user."), false
	}

	return u.Message(false, "Requirement passed"), true

}

//Create создание аккаунта для пользователей
func (account *Account) Create() map[string]interface{} {
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

	db.Exec(account.BoxPoint.Point0.ToSqlString("accounts", "points0", account.Login))
	db.Exec(account.BoxPoint.Point1.ToSqlString("accounts", "points1", account.Login))
	//db.Exec(account.Privilege.ToSqlStrUpdate("accounts", account.Login))

	account.Password = ""
	resp := u.Message(true, "Account has been created")
	resp["login"] = account.Login
	return resp
}

//SuperCreate создание суперпользователя
func (account *Account) SuperCreate() *Account {
	account.Login = "Super"
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	account.WTime = 24
	account.Password = "$2a$10$ZCWyIEfEVF3KGj6OUtIeSOQ3WexMjuAZ43VSO6T.QqOndn4HN1J6C"
	account.BoxPoint.Point0.SetPoint(42.79610884568009, 25.56378846464164)
	account.BoxPoint.Point1.SetPoint(77.13872007901705, -174.12371153535893)
	//account.Privilege.Role.Name = "Super"
	return account
}

//ParserPointsUser заполняет поля Point в аккаунт
func (account *Account) ParserPointsUser() {
	var str string
	row := db.Raw("select points0 from accounts where login = ?", account.Login).Row()
	row.Scan(&str)
	account.BoxPoint.Point0.StrToFloat(str)
	row = db.Raw("select points1 from accounts where login = ?", account.Login).Row()
	row.Scan(&str)
	account.BoxPoint.Point1.StrToFloat(str)
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
	tflight := GetLightsFromBD(account.BoxPoint.Point0, account.BoxPoint.Point1)
	resp := u.Message(true, "Take this DATA")

	resp["ya_map"] = account.YaMapKey
	resp["boxPoint"] = account.PointToMap()
	resp["tflight"] = tflight
	return resp
}

func (account *Account) PointToMap() (PointMap map[string]Point) {
	PointMap = make(map[string]Point, 2)
	PointMap["point0"] = account.BoxPoint.Point0
	PointMap["point1"] = account.BoxPoint.Point1
	return PointMap
}
