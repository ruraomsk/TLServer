package alarm

import (
	"fmt"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"time"
)

var (
	typeClose     = "close"
	typeAlarmInfo = "alarmInfo"
	typeRingData  = "alarm"
)

//armResponse структура для отправки сообщений (map)
type alarmResponse struct {
	Type string                 `json:"type"` //тип сообщения
	Data map[string]interface{} `json:"data"` //данные
}

//newMapMess создание нового сообщения
func newAlarmMess(mType string, data map[string]interface{}) alarmResponse {
	var resp alarmResponse
	resp.Type = mType
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

//AlarmInfo информация о запрашиваемом арме
type Info struct {
	Region int `json:"region"` //регион запрошенные пользователем
	//Area    []string `json:"area"`   //район запрошенные пользователем
	AccInfo *accToken.Token
}
type CrossRing struct {
	Ring      bool `json:"ring"`
	CrossInfo map[string]*CrossInfo
}
type CrossResponse struct {
	Ring      bool `json:"ring"`
	CrossInfo []*CrossInfo
}

//CrossInfo информация о перекрестке для техАРМ
type CrossInfo struct {
	Time       time.Time `json:"time"`     //время первого появления
	Region     int       `json:"region"`   //регион
	Area       int       `json:"area"`     //район
	ID         int       `json:"id"`       //id
	Subarea    int       `json:"subarea"`  //подрайон
	Idevice    int       `json:"idevice"`  //идентификатор устройства
	Describe   string    `json:"describe"` //описание
	StatusCode int       `json:"-"`        //статус кодом
	Status     string    `json:"status"`   //статус строкой
	Control    bool      `json:"-"`        //true если можно управлять false в противном случае
}

func key(region, area, id int) string {
	return fmt.Sprintf("%d:%d:%d", region, area, id)
}
