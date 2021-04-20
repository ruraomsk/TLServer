package data

import (
	"sync"
)

//AreaOnMap струтура для массива районов
type AreaOnMap struct {
	Mux   sync.Mutex
	Areas []AreaZone `json:"areas"` //массив зон
}

//AreaBox информация об регионе, районе, и входяших подрайонах
type AreaZone struct {
	Region string        `json:"region"` //регион
	Area   string        `json:"area"`   //район
	Zone   Points        `json:"zone"`   //массив точек района
	Sub    []SybAreaZone `json:"sub"`    //массив подрайонов
}

//SybAreaBox информация о подрайонах
type SybAreaZone struct {
	SubArea int    `json:"subArea"` //подрайон номер
	Zone    Points `json:"zone"`    //массив точек подрайора
}
