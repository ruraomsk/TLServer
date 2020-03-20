package data

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	u "github.com/JanFant/TLServer/utils"
)

//BusyArmInfo глобальная переменная для управления
var BusyArmInfo mainBusyArm

//mainBusyArm общее хранилище информации для мапы
type mainBusyArm struct {
	mux        sync.Mutex
	mapBusyArm map[BusyArm]EditCrossInfo //занятые рабочие станции
}

//BusyArms массив занятых перекрестков (обмен)
type BusyArms struct {
	BusyArms []BusyArm `json:"busyArms"`
}

//BusyArm информация о занятом перекрестке
type BusyArm struct {
	Region      string `json:"region"`      //регион устройства
	Area        string `json:"area"`        //район устройства
	ID          int    `json:"ID"`          //ID устройства
	Description string `json:"description"` //описание устройства
	structStr   string //строка для запроса в бд
}

//EditCrossInfo информация о пользователе занявшем перекресток на изменение
type EditCrossInfo struct {
	Login    string    `json:"login"`    //логин пользователя
	EditFlag bool      `json:"editFlag"` //флаг разрешения на редактирование перекрестка
	Kick     bool      `json:"kick"`     //флаг закрытия арма у данного пользователя
	time     time.Time //метка времени
}

//toStr конвертировать в строку
func (busyArm *BusyArm) toStr() (str string, err error) {
	newByte, err := json.Marshal(busyArm)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}

//toStr конвертировать в структуру
func (busyArm *BusyArm) toStruct(str string) (err error) {
	err = json.Unmarshal([]byte(str), busyArm)
	if err != nil {
		return err
	}
	return nil
}

//DisplayCrossEditInfo сборка информации о занятых перекрестках
func DisplayCrossEditInfo(mapContx map[string]string) map[string]interface{} {
	busyArm := make(map[string][]BusyArm)
	resp := u.Message(true, "Display information about changing crosses")
	BusyArmInfo.mux.Lock()
	for busy, edit := range BusyArmInfo.mapBusyArm {
		if edit.time.Add(time.Second * 10).Before(time.Now()) {
			delete(BusyArmInfo.mapBusyArm, busy)
			continue
		}
		if mapContx["region"] == "*" {
			busyArm[edit.Login] = append(busyArm[edit.Login], busy)
		} else if busy.Region == mapContx["region"] {
			busyArm[edit.Login] = append(busyArm[edit.Login], busy)
		}
	}
	BusyArmInfo.mux.Unlock()
	for nameUser, arms := range busyArm {
		for numArm, arm := range arms {
			sqlStr := fmt.Sprintf("select describ from %v where region = %v and area = %v and id = %v", GlobalConfig.DBConfig.CrossTable, arm.Region, arm.Area, arm.ID)
			rowsTL := GetDB().Raw(sqlStr).Row()
			_ = rowsTL.Scan(&busyArm[nameUser][numArm].Description)
		}
	}
	CacheInfo.mux.Lock()
	resp["regionInfo"] = CacheInfo.mapRegion
	resp["areaInfo"] = CacheInfo.mapArea
	CacheInfo.mux.Unlock()
	resp["CrossEditInfo"] = busyArm
	return resp
}

//FreeCrossEdit освобождение занятых перекрестков
func FreeCrossEdit(busyArm BusyArms) map[string]interface{} {
	BusyArmInfo.mux.Lock()
	defer BusyArmInfo.mux.Unlock()
	for _, arm := range busyArm.BusyArms {
		arm.Description = ""
		edit := BusyArmInfo.mapBusyArm[arm]
		edit.Kick = true
		BusyArmInfo.mapBusyArm[arm] = edit
	}
	resp := u.Message(true, "Release task was sent")
	return resp
}

//CleanMapBusyArm очистка хранилища от просроченных записей
func CleanMapBusyArm() {
	BusyArmInfo.mux.Lock()
	defer BusyArmInfo.mux.Unlock()
	for busyArm, editCross := range BusyArmInfo.mapBusyArm {
		if editCross.time.Add(time.Second * 10).Before(time.Now()) {
			delete(BusyArmInfo.mapBusyArm, busyArm)
		}
	}
}

//BusyArmDelete удаление из хранилища оспобожденного перекрестка
func BusyArmDelete(arm BusyArm) map[string]interface{} {
	resp := make(map[string]interface{})
	BusyArmInfo.mux.Lock()
	delete(BusyArmInfo.mapBusyArm, arm)
	BusyArmInfo.mux.Unlock()
	resp["ArmDelete"] = arm
	return resp
}
