package locations

import (
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
)

//Point координаты точки
type Point struct {
	Y float64
	X float64
}

//BoxPoint координаты для отрисовки зоны работы пользователя
type BoxPoint struct {
	Point0 Point `json:"point0"` //левая нижняя точка на карте
	Point1 Point `json:"point1"` //правая верхняя точка на карте
}

//StrToFloat преобразует строку, полученную из БД в структуру Point
func (points *Point) StrToFloat(str string) {
	str = strings.TrimPrefix(str, "(")
	str = strings.TrimSuffix(str, ")")
	temp := strings.Split(str, ",")
	if len(temp) != 2 {
		points.Y, points.X = 0, 0
		return
	}
	for num, part := range temp {
		temp[num] = strings.TrimSpace(part)
	}
	points.Y, _ = strconv.ParseFloat(temp[0], 64)
	points.X, _ = strconv.ParseFloat(temp[1], 64)
}

//TakePointFromBD запрос координат перекрестка из БД
func TakePointFromBD(numRegion, numArea, numID string, db *sqlx.DB) (point Point, err error) {
	var dgis string
	rowsTL := db.QueryRow(`SELECT dgis FROM public.cross WHERE region = $1 and area = $2 and id = $3`, numRegion, numArea, numID)
	err = rowsTL.Scan(&dgis)
	if err != nil {
		return point, err
	}
	point.StrToFloat(dgis)
	return point, nil
}
