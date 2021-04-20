package routeGS

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/sockets"
)

//Route маршрут движения
type Route struct {
	Region      string        `json:"region"`      //регион
	Description string        `json:"description"` //описание маршрута
	Box         data.BoxPoint `json:"box"`         //координаты на которые перемещаться при выборе маршрута
	List        []RouteTL     `json:"listTL"`      //список светофоров входящих в маршрут
}

//RouteTL информация о перекрестке
type RouteTL struct {
	Num         int             `json:"num"`         //порятковый новер светофора в маршруте
	Phase       int             `json:"phase"`       //фаза заданная для данного перекрестка
	Description string          `json:"description"` //описание светофора
	Point       data.Point      `json:"point"`       //координаты перекретска
	Pos         sockets.PosInfo `json:"pos"`         //информация о перекрестка (где находится)
}

var (
	errCantWriteInBD      = "запись в БД не удалась"
	errCantDeleteFromBD   = "удаление из БД не удалось"
	errEntryAlreadyExists = "заданное название маршрута уже существует"
)

//Create создание/запись маршрута в БД
func (r *Route) Create() error {
	db, id := data.GetDB()
	defer data.FreeDB(id)
	r.setBox()
	list, _ := json.Marshal(r.List)
	box, _ := json.Marshal(r.Box)
	for numR, route := range r.List {
		rowRoute := db.QueryRow(`SELECT describ FROM public.cross WHERE region = $1 AND area = $2 AND id = $3`, route.Pos.Region, route.Pos.Area, route.Pos.Id)
		_ = rowRoute.Scan(&r.List[numR].Description)
	}

	var temp string
	row := db.QueryRow(`SELECT description FROM public.routes WHERE description = $1 AND region = $2`, r.Description, r.Region)
	err := row.Scan(&temp)
	if err != sql.ErrNoRows {
		return errors.New(errEntryAlreadyExists)
	}

	_, err = db.Exec(`INSERT INTO  public.routes (description, box, listtl, region) VALUES ($1, $2, $3, $4)`, r.Description, string(box), string(list), r.Region)
	if err != nil {
		return errors.New(errCantWriteInBD)
	}
	return nil
}

//Update обновление маршрута в БД
func (r *Route) Update() error {
	db, id := data.GetDB()
	defer data.FreeDB(id)

	r.setBox()
	list, _ := json.Marshal(r.List)
	box, _ := json.Marshal(r.Box)
	for numR, route := range r.List {
		rowRoute := db.QueryRow(`SELECT describ FROM public.cross WHERE region = $1 AND area = $2 AND id = $3`, route.Pos.Region, route.Pos.Area, route.Pos.Id)
		_ = rowRoute.Scan(&r.List[numR].Description)
	}
	_, err := db.Exec(`UPDATE public.routes SET box = $1, listtl = $2 WHERE description = $3  AND region = $4`, string(box), string(list), r.Description, r.Region)
	if err != nil {
		return errors.New(errCantWriteInBD)
	}
	return nil
}

//Delete удаление маршрута из БД
func (r *Route) Delete() error {
	db, id := data.GetDB()
	defer data.FreeDB(id)
	_, err := db.Exec(`DELETE FROM public.routes WHERE description = $1 AND region = $2`, r.Description, r.Region)
	if err != nil {
		return errors.New(errCantDeleteFromBD)
	}
	return nil
}

//setBox создать область в которую входять все перекрестки
func (r *Route) setBox() {
	var (
		box data.BoxPoint
	)
	for num, tl := range r.List {
		//чтобы не мешалось отприцательное значение сделаем положительным
		if tl.Point.X < 0 {
			tl.Point.X += 360.0
		}
		if tl.Point.Y < 0 {
			tl.Point.Y += 180.0
		}
		//если первая запись не разбираясь записываем
		if num == 0 {
			box.Point0 = tl.Point
			box.Point1 = tl.Point
			continue
		}
		if tl.Point.X < box.Point0.X {
			box.Point0.X = tl.Point.X
		}
		if tl.Point.Y < box.Point0.Y {
			box.Point0.Y = tl.Point.Y
		}
		if tl.Point.X > box.Point1.X {
			box.Point1.X = tl.Point.X
		}
		if tl.Point.Y > box.Point1.Y {
			box.Point1.Y = tl.Point.Y
		}
	}
	if box.Point0.X > 180 {
		box.Point0.X -= 360.0
	}
	if box.Point1.X > 180 {
		box.Point1.X -= 360.0
	}
	if box.Point0.Y > 90 {
		box.Point0.Y -= 180.0
	}
	if box.Point1.Y > 90 {
		box.Point1.Y -= 180.0
	}
	r.Box = box
}
