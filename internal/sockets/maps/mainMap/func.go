package mainMap

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/internal/sockets/chat"
	"github.com/JanFant/TLServer/internal/sockets/crossSock/mainCross"
	"github.com/JanFant/TLServer/internal/sockets/techArm"
	"github.com/JanFant/TLServer/internal/sockets/xctrl"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

//checkToken проверка токена для вебсокета
func checkToken(c *gin.Context, db *sqlx.DB) (flag bool, t *data.Token) {
	var tokenString string
	cookie, err := c.Cookie("Authorization")
	//Проверка куков получили ли их вообще
	if err != nil {
		return false, nil
	}
	tokenString = cookie

	ip := strings.Split(c.Request.RemoteAddr, ":")
	//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
	if tokenString == "" {
		return false, nil
	}
	//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
	splitted := strings.Split(tokenString, " ")
	if len(splitted) != 2 {
		return false, nil
	}

	//берем часть где хранится токен
	tokenSTR := splitted[1]
	tk := &data.Token{}

	token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(license.LicenseFields.TokenPass), nil
	})

	//не правильный токен возвращаем ошибку с кодом 403
	if err != nil {
		return false, nil
	}

	//Проверка на уникальность токена
	var (
		userPrivilege  data.Privilege
		tokenStrFromBd string
	)
	rows, err := db.Query(`SELECT token, privilege FROM public.accounts WHERE login = $1`, tk.Login)
	if err != nil {
		return false, nil
	}
	for rows.Next() {
		_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
	}

	if tokenSTR != tokenStrFromBd || tk.IP != ip[0] || !token.Valid {
		return false, nil
	}

	//проверка токен пришел от правильного URL

	//проверка правильности роли для указанного пользователя
	_ = userPrivilege.ConvertToJson()
	if userPrivilege.Role.Name != tk.Role {
		return false, nil
	}

	return true, tk
}

//logIn обработчик авторизации пользователя в системе
func logIn(login, password, ip string, db *sqlx.DB) map[string]interface{} {
	resp := make(map[string]interface{})
	ipSplit := strings.Split(ip, ":")
	account := &data.Account{}
	//Забираю из базы запись с подходящей почтой
	rows, err := db.Query(`SELECT id, login, password, work_time, description FROM public.accounts WHERE login=$1`, login)
	if rows == nil {
		resp["status"] = false
		resp["message"] = fmt.Sprintf("Неверно указан логин или пароль")
		return resp
	}
	if err != nil {
		resp["status"] = false
		resp["message"] = "Потеряно соединение с сервером БД"
		return resp
	}
	for rows.Next() {
		_ = rows.Scan(&account.ID, &account.Login, &account.Password, &account.WorkTime, &account.Description)
	}

	//Авторизировались добираем полномочия
	privilege := data.Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		resp["status"] = false
		resp["message"] = fmt.Sprintf("Неверно указан логин или пароль")
		return resp
	}

	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		resp["status"] = false
		resp["message"] = fmt.Sprintf("Неверно указан логин или пароль")
		return resp
	}
	//Залогинились, создаем токен
	account.Password = ""
	tk := &data.Token{
		UserID:      account.ID,
		Login:       account.Login,
		IP:          ipSplit[0],
		Role:        privilege.Role.Name,
		Region:      privilege.Region,
		Area:        privilege.Area,
		Permission:  privilege.Role.Perm,
		Description: account.Description,
	}
	//врямя выдачи токена
	tk.IssuedAt = time.Now().Unix()
	//время когда закончится действие токена
	tk.ExpiresAt = time.Now().Add(time.Hour * account.WorkTime).Unix()

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(license.LicenseFields.TokenPass))
	account.Token = tokenString
	//сохраняем токен в БД чтобы точно знать что дейтвителен только 1 токен

	_, err = db.Exec(`UPDATE public.accounts SET token = $1 WHERE login = $2`, account.Token, account.Login)
	if err != nil {
		resp["status"] = false
		resp["message"] = "Потеряно соединение с сервером БД"
		return resp
	}

	//Формируем ответ
	resp["status"] = true
	resp["login"] = account.Login
	resp["token"] = tokenString
	resp["role"] = privilege.Role.Name
	resp["manageFlag"], _ = data.AccessCheck(login, privilege.Role.Name, 2)
	resp["logDeviceFlag"], _ = data.AccessCheck(login, privilege.Role.Name, 5)
	resp["techArmFlag"], _ = data.AccessCheck(login, privilege.Role.Name, 7)
	resp["gsFlag"], _ = data.AccessCheck(login, privilege.Role.Name, 8)
	resp["authorizedFlag"] = true
	resp["description"] = account.Description
	resp["region"] = privilege.Region
	//собрать в районы с их названиями
	var areaMap = make(map[string]string)
	for _, area := range privilege.Area {
		var tempA data.AreaInfo
		tempA.SetAreaInfo(privilege.Region, area)
		areaMap[tempA.Num] = tempA.NameArea
	}
	resp["area"] = areaMap

	data.CacheArea.Mux.Lock()
	resp["areaZone"] = data.CacheArea.Areas
	data.CacheArea.Mux.Unlock()
	return resp
}

//logOut выход из учетной записи
func logOut(login string, db *sqlx.DB) bool {
	_, err := db.Exec("UPDATE public.accounts SET token = $1 where login = $2", "", login)
	if err != nil {
		return false
	}
	return true
}

//logOutSockets закрытие всех сокетов по действию logout на основном экране
func logOutSockets(login string) {
	chat.UserLogoutChat <- login
	//mainCross.UserLogoutCrControl <- login
	mainCross.UserLogoutCross <- login
	techArm.UserLogoutTech <- login
	xctrl.UserLogoutXctrl <- login
	UserLogoutGS <- login
}
