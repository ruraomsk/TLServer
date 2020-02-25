package data

import (
	"fmt"
	u "github.com/JanFant/TLServer/utils"
	"os"
	"sync"
	"time"
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
}

//EditCrossInfo информация о пользователе занявшем перекресток на изменение
type EditCrossInfo struct {
	Login    string `json:"login"`    //логин пользователя
	EditFlag bool   `json:"editFlag"` //флаг разрешения на редактирование перекрестка
	Kick     bool   `json:"kick"`     //флаг закрытия арма у данного пользователя
	time     time.Time
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
			sqlStr := fmt.Sprintf("select describ from %v where region = %v and area = %v and id = %v", os.Getenv("gis_table"), arm.Region, arm.Area, arm.ID)
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
	resp := u.Message(true, "Release task were sent")
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
