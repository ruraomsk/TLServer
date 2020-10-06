package phaseInfo

import "time"

//phaseInfo инофрмация о фазах
type Phase struct {
	Idevice int       `json:"idevice"` //идентификатор утройства
	Fdk     int       `json:"fdk"`     //фаза
	Tdk     int       `json:"tdk"`     //время обработки
	Time    time.Time `json:"time"`
}
