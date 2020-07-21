package sockets

//PosInfo положение перекрестка
type PosInfo struct {
	Region string `json:"region"` //регион
	Area   string `json:"area"`   //район
	Id     int    `json:"id"`     //ID
}
