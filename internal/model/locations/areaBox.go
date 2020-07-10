package locations

import (
	"sync"
)

//AreaOnMap струтура для массива районов
type AreaOnMap struct {
	Mux   sync.Mutex
	Areas []AreaBox
}

//AreaBox информация об регионе, районе, и входяших подрайонах
type AreaBox struct {
	Region string
	Area   string
	Box    BoxPoint
	Sub    []SybAreaBox
}

//SybAreaBox информация о подрайонах
type SybAreaBox struct {
	SubArea int
	Box     BoxPoint
}
