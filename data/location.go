package data

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

//Point координаты точки
type Point struct {
	Y, X float64 //Координата Х и Y
}

//BoxPoint координаты для отрисовки квадрата
type BoxPoint struct {
	Point0 Point `json:"point0"` //левая нижняя точка на карте
	Point1 Point `json:"point1"` //правая верхняя точка на карте
}

//GetPoint возврашает значение координат
func (points *Point) GetPoint() (y, x float64) {
	return points.Y, points.X
}

//SetPoint задать значение координат
func (points *Point) SetPoint(y, x float64) {
	points.X, points.Y = x, y
}

// //ToSqlString формирует SQL строку для обновления координат в БД
// func (points *Point) ToSqlString(table, column, login string) string {
// 	return fmt.Sprintf("update %s set %s = '(%f,%f)' where login = '%s'", table, column, points.Y, points.X, login)
// }

//StrToFloat преобразует строку полученную из БД в структуру Point
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
func TakePointFromBD(numRegion, numArea, numID string) (point Point, err error) {
	var dgis string
	sqlStr := fmt.Sprintf("select dgis from %s where region = %v and area = %v and id = %v", os.Getenv("gis_table"), numRegion, numArea, numID)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err = rowsTL.Scan(&dgis)
	if err != nil {
		return point, err
	}
	point.StrToFloat(dgis)
	return point, nil
}
