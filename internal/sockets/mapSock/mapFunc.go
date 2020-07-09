package mapSock

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

//logIn обработчик авторизации пользователя в системе
func logIn(login, password, ip string, db *sqlx.DB) MapSokResponse {
	ipSplit := strings.Split(ip, ":")
	account := &data.Account{}
	//Забираю из базы запись с подходящей почтой
	rows, err := db.Query(`SELECT id, login, password, work_time, description FROM public.accounts WHERE login=$1`, login)
	if rows == nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = fmt.Sprintf("Invalid login credentials. login(%s)", login)
		return resp
	}
	if err != nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = "connection to DB error. Please try again"
		return resp
	}
	for rows.Next() {
		_ = rows.Scan(&account.ID, &account.Login, &account.Password, &account.WorkTime, &account.Description)
	}

	//Авторизировались добираем полномочия
	privilege := data.Privilege{}
	err = privilege.ReadFromBD(account.Login)
	if err != nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = fmt.Sprintf("Invalid login credentials. login(%s)", login)
		return resp
	}

	//Сравниваю хэши полученного пароля и пароля взятого из БД
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = fmt.Sprintf("Invalid login credentials. login(%s)", account.Login)
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
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = "connection to DB error. Please try again"
		return resp
	}

	//Формируем ответ
	resp := newMapMess(typeLogin, nil, nil)
	resp.Data["login"] = account.Login
	resp.Data["token"] = tokenString
	resp.Data["manageFlag"], _ = data.AccessCheck(login, privilege.Role.Name, 2)
	resp.Data["logDeviceFlag"], _ = data.AccessCheck(login, privilege.Role.Name, 5)
	resp.Data["techArmFlag"], _ = data.AccessCheck(login, privilege.Role.Name, 7)
	resp.Data["authorizedFlag"] = true
	resp.Data["description"] = account.Description
	resp.Data["region"] = privilege.Region
	//собрать в районы с их названиями
	var areaMap = make(map[string]string)
	for _, area := range privilege.Area {
		var tempA data.AreaInfo
		tempA.SetAreaInfo(privilege.Region, area)
		areaMap[tempA.Num] = tempA.NameArea
	}
	resp.Data["area"] = areaMap

	data.CacheArea.Mux.Lock()
	resp.Data["areaBox"] = data.CacheArea.Areas
	data.CacheArea.Mux.Unlock()
	return resp
}

//logOut выход из учетной записи
func logOut(login string, db *sqlx.DB) MapSokResponse {
	_, err := db.Exec("UPDATE public.accounts SET token = $1 where login = $2", "", login)
	if err != nil {
		resp := newMapMess(typeError, nil, nil)
		resp.Data["message"] = "connection to DB error. Please try again"
		return resp
	}
	return newMapMess(typeLogOut, nil, nil)
}

//selectTL возвращает массив в котором содержатся светофоры, которые попали в указанную область
func selectTL(db *sqlx.DB) (tfdata []data.TrafficLights) {
	var dgis string
	temp := &data.TrafficLights{}
	rowsTL, err := db.Query(`SELECT region, area, subarea, id, idevice, dgis, describ, status FROM public.cross`)
	if err != nil {
		logger.Error.Println("|Message: db not respond", err.Error())
		return nil
	}
	for rowsTL.Next() {
		err := rowsTL.Scan(&temp.Region.Num, &temp.Area.Num, &temp.Subarea, &temp.ID, &temp.Idevice, &dgis, &temp.Description, &temp.Sost.Num)
		if err != nil {
			logger.Error.Println("|Message: No result at these points", err.Error())
			return nil
		}
		temp.Points.StrToFloat(dgis)
		data.CacheInfo.Mux.Lock()
		temp.Region.NameRegion = data.CacheInfo.MapRegion[temp.Region.Num]
		temp.Area.NameArea = data.CacheInfo.MapArea[temp.Region.NameRegion][temp.Area.Num]
		temp.Sost.Description = data.CacheInfo.MapTLSost[temp.Sost.Num]
		data.CacheInfo.Mux.Unlock()
		tfdata = append(tfdata, *temp)
	}

	return tfdata
}

//mapOpenInfo сбор всех данных для отображения информации на карте
func mapOpenInfo(db *sqlx.DB) (obj map[string]interface{}) {
	obj = make(map[string]interface{})

	location := &data.Locations{}
	box, _ := location.MakeBoxPoint()
	obj["boxPoint"] = &box
	obj["tflight"] = selectTL(db)
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

//checkConnect проверка соединения с БД и Сервером
func checkConnect(db *sqlx.DB) interface{} {
	var tempStatus = struct {
		StatusBD bool `json:"statusBD"`
		StatusS  bool `json:"statusS"`
	}{
		StatusBD: false,
		StatusS:  false,
	}

	_, err := db.Exec(`SELECT * FROM public.accounts;`)
	if err == nil {
		tempStatus.StatusBD = true
	}

	var tcpPackage = tcpConnect.TCPMessage{Type: tcpConnect.TypeState, User: "TestConn", Id: -1, Data: 0}
	tempStatus.StatusS = tcpPackage.SendToTCPServer()
	return tempStatus
}
