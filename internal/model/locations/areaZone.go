package locations

import (
	"sync"
)

//AreaOnMap струтура для массива районов
type AreaOnMap struct {
	Mux   sync.Mutex
	Areas []AreaZone `json:"areas"`
}

//AreaBox информация об регионе, районе, и входяших подрайонах
type AreaZone struct {
	Region string        `json:"region"`
	Area   string        `json:"area"`
	Zone   Points        `json:"zone"`
	Sub    []SybAreaZone `json:"sub"`
}

//SybAreaBox информация о подрайонах
type SybAreaZone struct {
	SubArea int    `json:"subArea"`
	Zone    Points `json:"zone"`
}
