package data

import (
	"fmt"
	"strconv"
	"strings"
)

type Point struct {
	X, Y float64    //Координата Х и Y
}

//GetPoint возврашает значение координат
func (points *Point) GetPoint() (x, y float64) {
	return points.X, points.Y
}

//SetPoint задать значение координат
func (points *Point) SetPoint(x, y float64) {
	points.X, points.Y = x, y
}

//ToSqlString формирует SQL строку для обновления координат в БД
func (points *Point) ToSqlString(table, column, email string) string {
	return fmt.Sprintf("update %s set %s = '(%f,%f)' where email = '%s'", table, column, points.X, points.Y, email)
}

//StrToFloat преобразует строку полученную из бз в структуру Point
func (points *Point) StrToFloat(str string) {
	str = strings.TrimPrefix(str, "(")
	str = strings.TrimSuffix(str, ")")
	temp := strings.Split(str, ",")
	points.X, _ = strconv.ParseFloat(temp[0], 64)
	points.Y, _ = strconv.ParseFloat(temp[1], 64)
}
