package data

import (
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
	jwt.StandardClaims
}

//Account struct to user account
type Account struct {
	gorm.Model
	Login    string        `json:"login",sql:"login"` //Имя пользователя
	Password string        `json:"password"`          //Пароль
	Point0   Point         `json:"point0",sql:"-"`    //Первая точка области
	Point1   Point         `json:"point1",sql:"-"`    //Вторая точка области
	WTime    time.Duration `json:"wtime",sql:"wtime"` //Время работы пользователя в часах
	YaMapKey string        `json:"ya_key",sql:"-"`    //Ключ доступа к ндекс карте
	Token    string        `json:"token",sql:"-"'`    //Токен пользователя
}

//Login in system
func Login(login, password string) map[string]interface{} {
	account := &Account{}
	//Забираю из базы запись с подходящей почтой
	err := GetDB().Table("accounts").Where("login = ?", login).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "login not found")
		}
		return u.Message(false, "Connection error. Please try again")
	}
	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(false, "Invalid login credentials. Please try again")
	}
	//Залогинились, создаем токен
	account.Password = ""
	tk := &Token{UserID: account.ID, Login: account.Login}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = time.Now().Add(time.Hour * account.WTime).Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)

	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString
	account.ParserPointsUser()
	trlight := GetLightsFromBD(account.Point0, account.Point1)
	//Записываем координаты подложки
	account.ParserPointsUser()
	//Формируем ответ
	resp := u.Message(true, "Logged In")
	resp["login"] = account.Login
	resp["trlight"] = trlight
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

	db.Exec(account.Point0.ToSqlString("accounts", "points0", account.Login))
	db.Exec(account.Point1.ToSqlString("accounts", "points1", account.Login))

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
	account.Point0.SetPoint(55.00000121541251, 36.000000154512121)
	account.Point1.SetPoint(56.3, 36.5)
	return account
}

//ParserPointsUser заполняет поля Point в аккаунт
func (account *Account) ParserPointsUser() {
	var str string
	row := db.Raw("select points0 from accounts where login = ?", account.Login).Row()
	row.Scan(&str)
	account.Point0.StrToFloat(str)
	row = db.Raw("select points1 from accounts where login = ?", account.Login).Row()
	row.Scan(&str)
	account.Point1.StrToFloat(str)
}
