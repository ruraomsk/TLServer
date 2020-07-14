package mapSock

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/locations"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
)

type Mode struct {
	Id          int                `json:"id"`
	Description string             `json:"description"`
	Box         locations.BoxPoint `json:"box"`
	List        []ModeTL           `json:"listTL"`
}

type ModeTL struct {
	Num   int               `json:"num"`
	Phase int               `json:"phase"`
	Point locations.Point   `json:"point"`
	Pos   crossSock.PosInfo `json:"pos"`
}

func (m *Mode) create(db *sqlx.DB) error {
	m.setBox()
	list, _ := json.Marshal(m.List)
	box, _ := json.Marshal(m.Box)
	row := db.QueryRow(`INSERT INTO  public.modes (description, box, listtl) VALUES ($1, $2, $3) RETURNING id`, m.Description, string(box), string(list))
	err := row.Scan(&m.Id)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mode) setBox() {
	var (
		box locations.BoxPoint
	)
	for num, tl := range m.List {
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
	m.Box = box
}

func getAllModes(db *sqlx.DB) interface{} {
	var (
		modes = make([]Mode, 0)
	)
	rows, err := db.Query(`SELECT id, description, box, listtl FROM public.modes`)
	if err != nil {
		logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
		return modes
	}
	for rows.Next() {
		var (
			temp            Mode
			listSrt, boxStr string
		)
		err := rows.Scan(&temp.Id, &temp.Description, &boxStr, &listSrt)
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return modes
		}
		err = json.Unmarshal([]byte(listSrt), &temp.List)
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return modes
		}
		err = json.Unmarshal([]byte(boxStr), &temp.Box)
		if err != nil {
			logger.Error.Printf("|IP: - |Login: - |Resource: /greenStreet |Message: %v", err.Error())
			return modes
		}
		if len(temp.List) == 0 {
			temp.List = make([]ModeTL, 0)
		}
		modes = append(modes, temp)
	}
	return modes
}
