package locations

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"sync"
)

type AreaOnMap struct {
	Mux   sync.Mutex
	Areas []AreaBox
}

type AreaBox struct {
	Region string
	Area   string
	Box    BoxPoint
}

func (a *AreaBox) FillBox(db *sqlx.DB, region, area string) error {
	row := db.QueryRow(`SELECT Min(dgis[0]) as "Y0", Min(convTo360(dgis[1])) as "X0", Max(dgis[0]) as "Y1", Max(convTo360(dgis[1])) as "X1"  FROM public."cross" WHERE region = $1 AND area = $2`, region, area)
	err := row.Scan(&a.Box.Point0.Y, &a.Box.Point0.X, &a.Box.Point1.Y, &a.Box.Point1.X)
	if err != nil {
		return errors.New(fmt.Sprintf("parserPoints. Request error: %s", err.Error()))
	}
	if a.Box.Point0.X > 180 {
		a.Box.Point0.X -= 360
	}
	if a.Box.Point1.X > 180 {
		a.Box.Point1.X -= 360
	}
	return nil
}
