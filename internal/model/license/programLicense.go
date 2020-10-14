package license

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
	"github.com/dgrijalva/jwt-go"
)

// key ключ шифрования токена должен совпадать с ключем на сервере создания ключа
var key = "yreRmn6JKVv1md1Yh1PptBIjtGrL8pRjo8sAp5ZPlR6zK8xjxnzt6mGi6mtjWPJ6lz1HbhgNBxfSReuqP9ijLQ4JiWLQ4ADHefWVgtTzeI35pqB6hsFjOWufdAW8UEdK9ajm3T76uQlucUP2g4rUV8B9gTMoLtkn5Pxk6G83YZrvAIR7ddsd5PreTwGDoLrS6bdsbJ7u"

//LicenseToken токен лицензии клиента
type LicenseToken struct {
	NumDevice int      //количество устройств
	YaKey     string   //ключ яндекса
	TokenPass string   //пароль для шифрования токена https запросов
	NumAcc    int      //колическво аккаунтов
	Name      string   //название фирмы
	Address   string   //расположение фирмы
	Phone     string   //телефон фирмы
	Id        int      //уникальный номер сервера
	TechEmail []string //почта для отправки сообщений в тех поддержку
	Email     string   //почта фирмы
	jwt.StandardClaims
}

//LicenseFields обращение к полям токена
var LicenseFields licenseInfo
var LogOutAllFromLicense chan bool //канал для закрытия сокетов, если введена новая лицензия

//licenseInfo информация о полях лицензии
type licenseInfo struct {
	Mux         sync.Mutex
	NumDev      int      //количество устройств
	NumAcc      int      //колическво аккаунтов
	YaKey       string   //ключ яндекса
	Address     string   //расположение фирмы
	Id          int      //уникальный номер сервера
	CompanyName string   //название фирмы
	TechEmail   []string //почта для отправки сообщений в тех поддержку
	TokenPass   string   //пароль для шифрования токена https запросов
}

//CheckLicenseKey проверка токена лицензии
func CheckLicenseKey(tokenSTR string) (*LicenseToken, error) {
	tk := &LicenseToken{}
	token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	//не правильный токен
	if err != nil {
		return tk, err
	}
	//не истек ли токен?
	if !token.Valid {
		return tk, errors.New("invalid token")
	}
	return tk, nil
}

//ControlLicenseKey процесс проверки токена раз в час
func ControlLicenseKey() {
	timeTick := time.NewTicker(time.Hour * 1)
	defer timeTick.Stop()
	for {
		select {
		case <-timeTick.C:
			{
				key, err := readFile()
				if err != nil {
					logger.Error.Println("|Message: license.key file don't read: ", err.Error())
					fmt.Println("license.key file don't read: ", err.Error(), " server stop")
					os.Exit(1)
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

//ParseFields разбираем поля полученного токена в структуру для работы
func (licInfo *licenseInfo) ParseFields(token *LicenseToken) {
	licInfo.Mux.Lock()
	defer licInfo.Mux.Unlock()
	licInfo.TokenPass = token.TokenPass
	licInfo.YaKey = token.YaKey
	licInfo.NumAcc = token.NumAcc
	licInfo.NumDev = token.NumDevice
	licInfo.Id = token.Id
	licInfo.CompanyName = token.Name
	licInfo.Address = token.Address
	licInfo.TechEmail = token.TechEmail

}

//LicenseCheck проверка лицензи на старте
func LicenseCheck() {
	key, err := readFile()
	if err != nil {
		logger.Error.Println("|Message: license.key file don't read: ", err.Error())
		fmt.Println("license.key file don't read: ", err.Error())
	}
	tk, err := CheckLicenseKey(key)
	if err != nil {
		fmt.Print("Wrong License key")
		os.Exit(1)
	} else {
		LicenseFields.ParseFields(tk)
		fmt.Printf("Token END time:= %v\n", time.Unix(tk.ExpiresAt, 0))
		go ControlLicenseKey()
	}
}

//LicenseInfo запрос информации о лицензии
func LicenseInfo() u.Response {
	keyStr, err := readFile()
	if err != nil {
		logger.Error.Println("|Message: license.key file don't read: ", err.Error())
		fmt.Println("license.key file don't read: ", err.Error())
	}
	tk := &LicenseToken{}
	_, _ = jwt.ParseWithClaims(keyStr, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	resp := u.Message(http.StatusOK, "license Info")
	resp.Obj["numDev"] = tk.NumDevice
	resp.Obj["numAcc"] = tk.NumAcc
	resp.Obj["name"] = tk.Name
	resp.Obj["address"] = tk.Address
	resp.Obj["phone"] = tk.Phone
	resp.Obj["timeEnd"] = time.Unix(tk.ExpiresAt, 0)
	return resp
}

//LicenseNewKey запись нового ключа лицензии
func LicenseNewKey(keyStr string) u.Response {
	tk, err := CheckLicenseKey(keyStr)
	if err != nil {
		return u.Message(http.StatusBadRequest, "wrong License key")
	}
	err = writeFile(keyStr)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "error write license.key file")
	}
	LicenseFields.ParseFields(tk)
	LogOutAllFromLicense <- true
	resp := u.Message(http.StatusOK, "new key saved")
	return resp
}

//readFile прочитать файл лицензии
func readFile() (string, error) {
	byteFile, err := ioutil.ReadFile("configs/license.key")
	if err != nil {
		logger.Error.Println("|Message: Error reading license.key file")
		return "", err
	}
	return string(byteFile), nil
}

//writeFile записать файл лицензии
func writeFile(tokenStr string) error {
	err := ioutil.WriteFile("./configs/license.key", []byte(tokenStr), os.ModePerm)
	if err != nil {
		logger.Error.Println("|Message: Error write license.key file")
		return err
	}
	return nil
}
