package data

import (
	u "../utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
)

//Token JWT claims struct
type Token struct {
	UserID uint     		//Уникальный ID пользователя
	jwt.StandardClaims
}

//Account struct to user account
type Account struct {
	gorm.Model
	Email    string `json:"email"`          //Почта пользователя
	Password string `json:"password"` 		//Пароль
	Point0   Point  `json:"point0",sql:"-"` //Первая точка области
	Point1   Point  `json:"point1",sql:"-"` //Вторая точка области
	YaMapKey string `json:"ya_key",sql:"-"` //Ключ доступа к ндекс карте
	Token    string `json:"token",sql:"-"'` //Токен пользователя
}

//Login in system
func Login(email, password string) map[string]interface{} {
	account := &Account{}
	//Забираю из базы запись с подходящей почтой
	err := GetDB().Table("accounts").Where("email = ?", email).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return u.Message(false, "Email address not found")
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
	tk := &Token{UserID: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString
	//Записываем координаты подложки
	account.ParserPointsUser()
	//Отдаем ключ для yandex map
	account.YaMapKey = os.Getenv("ya_key")
	//Получить информацию о сфетофорах которые входят в начальную зану
	tflight := GetLightsFromBD(account.Point0, account.Point1)
	//Формируем ответ
	resp := u.Message(true, "Logged In")
	resp["account"] = account
	resp["tflight"] = tflight
	return resp

}

//Validate checking for an account in the database
func (account *Account) Validate() (map[string]interface{}, bool) {
	if !strings.Contains(account.Email, "@") {
		return u.Message(false, "Email address is required"), false
	}

	if len(account.Password) < 6 {
		return u.Message(false, "Password is required"), false
	}

	//Email mast be unique
	temp := &Account{}

	//check for error and duplicate emails
	err := GetDB().Table("accounts").Where("email = ?", account.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error, please try again"), false
	}
	if temp.Email != "" {
		return u.Message(false, "Email address already in use by another user."), false
	}

	return u.Message(false, "Requirement passed"), true

}

//Create создание аккаунта для пользователей
func (account *Account) Create() map[string]interface{} {
	if resp, ok := account.Validate(); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	GetDB().Table("accounts").Create(account)
	if account.ID <= 0 {
		return u.Message(false, "Failed to create account< connection error.")
	}

	//создаем токен для аккаунта
	tk := &Token{UserID: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString

	account.Password = ""
	resp := u.Message(true, "Account has been created")
	resp["account"] = account
	return resp
}

//SuperCreate создание суперпользователя
func (account *Account) SuperCreate() *Account {
	account.Email = "super@super"
	account.Password = "$2a$10$ZCWyIEfEVF3KGj6OUtIeSOQ3WexMjuAZ43VSO6T.QqOndn4HN1J6C"
	account.Point0.SetPoint(55.00000121541251, 36.000000154512121)
	account.Point1.SetPoint(56.3, 36.5)
	return account
}

//ParserPointsUser заполняет поля Point в аккаунт
func (account *Account) ParserPointsUser() {
	var str string
	row := db.Raw("select points0 from accounts where email = ?", account.Email).Row()
	row.Scan(&str)
	account.Point0.StrToFloat(str)
	row = db.Raw("select points1 from accounts where email = ?", account.Email).Row()
	row.Scan(&str)
	account.Point1.StrToFloat(str)
}
