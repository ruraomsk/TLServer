package locations

import (
	"sync"
)

//AreaOnMap струтура для массива районов
type AreaOnMap struct {
	Mux   sync.Mutex
	Areas []AreaBox `json:"areas"`
}

//AreaBox информация об регионе, районе, и входяших подрайонах
type AreaBox struct {
	Region string       `json:"region"`
	Area   string       `json:"area"`
	Box    BoxPoint     `json:"box"`
	Sub    []SybAreaBox `json:"sub"`
}

//SybAreaBox информация о подрайонах
type SybAreaBox struct {
	SubArea int      `json:"subArea"`
	Box     BoxPoint `json:"box"`
}
