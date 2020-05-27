package crossEdit

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"net/http"
	"sync"
	"time"

	u "github.com/JanFant/newTLServer/internal/utils"
)

//BusyArmInfo глобальная переменная для управления
var BusyArmInfo mainBusyArm

//mainBusyArm общее хранилище информации для мапы
type mainBusyArm struct {
	Mux        sync.Mutex
	MapBusyArm map[BusyArm]EditCrossInfo //занятые рабочие станции
}

//BusyArms массив занятых перекрестков (обмен)
type BusyArms struct {
	BusyArms []BusyArm `json:"busyArms"`
}

//BusyArm информация о занятом перекрестке
type BusyArm struct {
	Region      string `json:"region"`      //регион
	Area        string `json:"area"`        //район
	ID          int    `json:"ID"`          //ID
	Description string `json:"description"` //описание
	structStr   string //строка для запроса в бд
}

//EditCrossInfo информация о пользователе, занявшем перекресток на изменение
type EditCrossInfo struct {
	Login    string    `json:"login"`    //логин пользователя
	EditFlag bool      `json:"editFlag"` //флаг разрешения на редактирование перекрестка
	Kick     bool      `json:"kick"`     //флаг закрытия арма у данного пользователя
	Time     time.Time //метка времени
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
func DisplayCrossEditInfo(mapContx map[string]string, db *sqlx.DB) u.Response {
	busyArm := make(map[string][]BusyArm)
	resp := u.Message(http.StatusOK, "display information about changing crosses")
	BusyArmInfo.Mux.Lock()
	for busy, edit := range BusyArmInfo.MapBusyArm {
		if edit.Time.Add(time.Second * 10).Before(time.Now()) {
			delete(BusyArmInfo.MapBusyArm, busy)
			continue
		}
		if mapContx["region"] == "*" {
			busyArm[edit.Login] = append(busyArm[edit.Login], busy)
		} else if busy.Region == mapContx["region"] {
			busyArm[edit.Login] = append(busyArm[edit.Login], busy)
		}
	}
	BusyArmInfo.Mux.Unlock()
	for nameUser, arms := range busyArm {
		for numArm, arm := range arms {
			_ = db.QueryRow(`SELECT describ FROM public.cross WHERE region = $1 AND area = $2 AND id = $3`, arm.Region, arm.Area, arm.ID).Scan(&busyArm[nameUser][numArm].Description)
		}
	}

	resp.Obj["CrossEditInfo"] = busyArm
	return resp
}

//FreeCrossEdit освобождение занятых перекрестков
func FreeCrossEdit(busyArm BusyArms) u.Response {
	BusyArmInfo.Mux.Lock()
	defer BusyArmInfo.Mux.Unlock()
	for _, arm := range busyArm.BusyArms {
		arm.Description = ""
		edit := BusyArmInfo.MapBusyArm[arm]
		edit.Kick = true
		BusyArmInfo.MapBusyArm[arm] = edit
	}
	resp := u.Message(http.StatusOK, "release task was sent")
	return resp
}

//CleanMapBusyArm очистка хранилища от просроченных записей
func CleanMapBusyArm() {
	BusyArmInfo.Mux.Lock()
	defer BusyArmInfo.Mux.Unlock()
	for busyArm, editCross := range BusyArmInfo.MapBusyArm {
		if editCross.Time.Add(time.Second * 10).Before(time.Now()) {
			delete(BusyArmInfo.MapBusyArm, busyArm)
		}
	}
}

//BusyArmDelete удаление из хранилища освобожденного перекрестка
func BusyArmDelete(arm BusyArm) u.Response {
	resp := u.Message(http.StatusOK, "delete busyArm")
	BusyArmInfo.Mux.Lock()
	delete(BusyArmInfo.MapBusyArm, arm)
	BusyArmInfo.Mux.Unlock()
	resp.Obj["ArmDelete"] = arm
	return resp
}
