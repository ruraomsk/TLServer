package data

import (
	"github.com/JanFant/TLServer/logger"
	u "github.com/JanFant/TLServer/utils"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
	"time"
)

//LicenseToken токен лицензии клиента
type LicenseToken struct {
	NumDevice int    //количество устройств
	YaKey     string //ключ яндекса
	TokenPass string //пароль для шифрования токена https запросов
	Name      string //название фирмы
	Phone     string //телефон фирмы
	Email     string //почта фирмы
	jwt.StandardClaims
}

//License информация о лицензии клиента (БД?)
type License struct {
	NumDevice     int       `json:"numDev"`    //количество устройств
	NameClient    string    `json:"name"`      //название фирмы
	AddressClient string    `json:"address"`   //адресс фирмы
	PhoneClient   string    `json:"phone"`     //телефон фирмы
	EmailClient   string    `json:"email"`     //емайл фирмы
	YaKey         string    `json:"yaKey"`     //ключ яндекса
	TokenPass     string    `json:"tokenPass"` //пароль для шифрования токена https запросов
	EndTime       time.Time `json:"time"`      //время окончания лицензии
}

func CreateLicenseToken(license License) map[string]interface{} {
	//создаем токен
	tk := &LicenseToken{Name: license.NameClient, YaKey: license.YaKey, Email: license.EmailClient, NumDevice: license.NumDevice, Phone: license.PhoneClient, TokenPass: license.TokenPass}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = license.EndTime.Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(GlobalConfig.TokenPassword))

	//сохраняем токен в БД
	//GetDB().Exec("update public.accounts set token = ? where login = ?", account.Token, account.Login)

	//Формируем ответ
	resp := u.Message(true, "LicenseToken")
	resp["token"] = tokenString
	resp["license"] = license
	return resp
}

func CheckLicenseKey(tokenSTR string) {
	tk := &LicenseToken{}
	token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(GlobalConfig.TokenPassword), nil
	})

	//не правильный токен
	if err != nil {
	}
	//не истек ли токен?
	if !token.Valid {

	}
}

func ControlLicenseKey() {
}

func readFile() (string, error) {
	byteFile, err := ioutil.ReadFile("license.key")
	if err != nil {
		logger.Error.Println("|Message: Error reading license.key file")
		return "", err
	}
	return string(byteFile), nil
}

func writeFile(tokenStr string) error {
	err := ioutil.WriteFile("license.key", []byte(tokenStr), os.ModePerm)
	if err != nil {
		logger.Error.Println("|Message: Error write license.key file")
		return err
	}
	return nil
}
