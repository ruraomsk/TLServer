package routeGS

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/locations"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/jmoiron/sqlx"
)

//Route маршрут движения
type Route struct {
	Id          int                `json:"id"`          //уникальный номер в бд
	Description string             `json:"description"` //описание маршрута
	Box         locations.BoxPoint `json:"box"`         //координаты на которые перемещаться при выборе маршрута
	List        []RouteTL          `json:"listTL"`      //список светофоров входящих в маршрут
}

//RouteTL информация о перекрестке
type RouteTL struct {
	Num   int               `json:"num"`   //порятковый новер светофора в маршруте
	Phase int               `json:"phase"` //фаза заданная для данного перекрестка
	Point locations.Point   `json:"point"` //координаты перекретска
	Pos   crossSock.PosInfo `json:"pos"`   //информация о перекрестка (где находится)
}

//Create создание/запись маршрута в БД
func (r *Route) Create(db *sqlx.DB) error {
	r.setBox()
	list, _ := json.Marshal(r.List)
	box, _ := json.Marshal(r.Box)
	row := db.QueryRow(`INSERT INTO  public.modes (description, box, listtl) VALUES ($1, $2, $3) RETURNING id`, r.Description, string(box), string(list))
	err := row.Scan(&r.Id)
	if err != nil {
		return err
	}
	return nil
}

//Update обновление маршрута в БД
func (r *Route) Update(db *sqlx.DB) error {
	r.setBox()
	list, _ := json.Marshal(r.List)
	box, _ := json.Marshal(r.Box)
	_, err := db.Exec(`UPDATE public.modes SET description = $1, box = $2, listtl = $3 WHERE id = $4`, r.Description, string(box), string(list), r.Id)
	if err != nil {
		return err
	}
	return nil
}

//Delete удаление маршрута из БД
func (r *Route) Delete(db *sqlx.DB) error {
	_, err := db.Exec(`DELETE FROM public.modes WHERE id = $1`, r.Id)
	if err != nil {
		return err
	}
	return nil
}

//setBox создать область в которую входять все перекрестки
func (r *Route) setBox() {
	var (
		box locations.BoxPoint
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
