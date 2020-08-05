package mapSock

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/binding"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

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

//SelectTL возвращает массив в котором содержатся светофоры, которые попали в указанную область
func SelectTL(db *sqlx.DB) (tfdata []data.TrafficLights) {
	var dgis string
	rowsTL, err := db.Query(`SELECT region, area, subarea, id, idevice, dgis, describ, status, state->'arrays'->'SetDK' FROM public.cross`)
	if err != nil {
		logger.Error.Println("|Message: db not respond", err.Error())
		return nil
	}
	for rowsTL.Next() {
		var (
			temp      = data.TrafficLights{}
			tempSetDK binding.SetDK
			dkStr     string
		)
		err := rowsTL.Scan(&temp.Region.Num, &temp.Area.Num, &temp.Subarea, &temp.ID, &temp.Idevice, &dgis, &temp.Description, &temp.Sost.Num, &dkStr)
		if err != nil {
			logger.Error.Println("|Message: No result at these points", err.Error())
			return nil
		}
		_ = json.Unmarshal([]byte(dkStr), &tempSetDK)
		temp.Phases = tempSetDK.GetPhases()
		temp.Points.StrToFloat(dgis)
		data.CacheInfo.Mux.Lock()
		temp.Region.NameRegion = data.CacheInfo.MapRegion[temp.Region.Num]
		temp.Area.NameArea = data.CacheInfo.MapArea[temp.Region.NameRegion][temp.Area.Num]
		temp.Sost.Description = data.CacheInfo.MapTLSost[temp.Sost.Num]
		data.CacheInfo.Mux.Unlock()
		tfdata = append(tfdata, temp)
	}
	return tfdata
}

//MapOpenInfo сбор всех данных для отображения информации на карте
func MapOpenInfo(db *sqlx.DB) (obj map[string]interface{}) {
	obj = make(map[string]interface{})

	location := &data.Locations{}
	box, _ := location.MakeBoxPoint()
	obj["boxPoint"] = &box
	obj["tflight"] = SelectTL(db)
	obj["authorizedFlag"] = false

	//собираю в кучу регионы для отображения
	chosenRegion := make(map[string]string)
	data.CacheInfo.Mux.Lock()
	for first, second := range data.CacheInfo.MapRegion {
		chosenRegion[first] = second
	}
	delete(chosenRegion, "*")
	obj["regionInfo"] = chosenRegion

	//собираю в кучу районы для отображения
	chosenArea := make(map[string]map[string]string)
	for first, second := range data.CacheInfo.MapArea {
		chosenArea[first] = make(map[string]string)
		chosenArea[first] = second
	}
	delete(chosenArea, "Все регионы")
	data.CacheInfo.Mux.Unlock()
	obj["areaInfo"] = chosenArea
	return
}
