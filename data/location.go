package data

import (
	"fmt"
	"strconv"
	"strings"
)

//Point координаты точки
type Point struct {
	Y, X float64 //Координата Х и Y
}

//BoxPoint координаты для отрисовки квадрата
type BoxPoint struct {
	Point0 Point
	Point1 Point
}

//GetPoint возврашает значение координат
func (points *Point) GetPoint() (y, x float64) {
	return points.X, points.Y
}

//SetPoint задать значение координат
func (points *Point) SetPoint(y, x float64) {
	points.X, points.Y = x, y
}

//ToSqlString формирует SQL строку для обновления координат в БД
func (points *Point) ToSqlString(table, column, login string) string {
	return fmt.Sprintf("update %s set %s = '(%f,%f)' where login = '%s'", table, column, points.Y, points.X, login)
}

//StrToFloat преобразует строку полученную из бз в структуру Point
func (points *Point) StrToFloat(str string) {
	str = strings.TrimPrefix(str, "(")
	str = strings.TrimSuffix(str, ")")
	temp := strings.Split(str, ",")
	points.Y, _ = strconv.ParseFloat(temp[0], 64)
	points.X, _ = strconv.ParseFloat(temp[1], 64)
}
