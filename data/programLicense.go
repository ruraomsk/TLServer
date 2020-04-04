package data

import (
	"errors"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	u "github.com/JanFant/TLServer/utils"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
	"sync"
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

//LicenseFields обращение к полям токена
var LicenseFields licenseInfo

//licenseInfo информация о полях лицензии
type licenseInfo struct {
	mux       sync.Mutex
	NumDev    int    //количество устройств
	Yakey     string //ключ яндекса
	TokenPass string //пароль для шифрования токена https запросов
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
	resp["tk"] = tk
	return resp
}

func CheckLicenseKey(tokenSTR string) (*LicenseToken, error) {
	tk := &LicenseToken{}
	token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(GlobalConfig.TokenPassword), nil
	})
	//не правильный токен
	if err != nil {
		return tk, err
	}
	//не истек ли токен?
	if !token.Valid {
		return tk, errors.New("Invalid token")
	}
	return tk, nil
}

func ControlLicenseKey() {
	var aaa = make(chan bool)
	timeTick := time.Tick(time.Hour * 1)
	for {
		select {
		case <-aaa:
			{

			}
		case <-timeTick:
			{
				key, err := readFile()
				if err != nil {
					logger.Error.Println("|Message: license.key file don't read: ", err.Error())
					fmt.Println("license.key file don't read: ", err.Error())
				}
				_, err = CheckLicenseKey(key)
				if err != nil {
					fmt.Print("Wrong License key")
					os.Exit(1)
				}
			}
		}
	}
}

func (licInfo *licenseInfo) ParseFields(token *LicenseToken) {
	licInfo.mux.Lock()
	defer licInfo.mux.Unlock()
	licInfo.TokenPass = token.TokenPass
	licInfo.Yakey = token.YaKey
	licInfo.NumDev = token.NumDevice
}

func LicenseCheck() {
	key, err := readFile()
	if err != nil {
		logger.Error.Println("|Message: license.key file don't read: ", err.Error())
		fmt.Println("license.key file don't read: ", err.Error())
	}
	for {
		tk, err := CheckLicenseKey(key)
		if err != nil {
			fmt.Print("Wrong License key")
			os.Exit(1)
		} else {
			LicenseFields.ParseFields(tk)
			fmt.Printf("Token END time:= %v\n", time.Unix(tk.ExpiresAt, 0))
			break
		}
	}
}

func LicenseInfo() map[string]interface{} {
	keyStr, err := readFile()
	if err != nil {
		logger.Error.Println("|Message: license.key file don't read: ", err.Error())
		fmt.Println("license.key file don't read: ", err.Error())
	}
	tk := &LicenseToken{}
	_, _ = jwt.ParseWithClaims(keyStr, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(GlobalConfig.TokenPassword), nil
	})
	resp := u.Message(true, "License Info")
	resp["tk"] = tk
	resp["Time END"] = time.Unix(tk.ExpiresAt, 0)
	return resp
}

func LicenseNewKey(keyStr string) map[string]interface{} {
	tk, err := CheckLicenseKey(keyStr)
	if err != nil {
		return u.Message(false, "Wrong License key")
	}
	err = writeFile(keyStr)
	if err != nil {
		return u.Message(false, "Error write license.key file")
	}
	LicenseFields.ParseFields(tk)
	resp := u.Message(true, "New key saved")
	return resp
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
