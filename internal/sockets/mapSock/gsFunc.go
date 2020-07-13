package mapSock

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/jmoiron/sqlx"
)

type Mode struct {
	Id          int      `json:"id"`
	Description string   `json:"description"`
	List        []ModeTL `json:"listTL"`
}

type ModeTL struct {
	Num   int               `json:"num"`
	Phase int               `json:"phase"`
	Pos   crossSock.PosInfo `json:"pos"`
}

func (m *Mode) create(db *sqlx.DB) {
	list, _ := json.Marshal(m.List)
	row := db.QueryRow(`INSERT INTO  public.modes (description, listtl) VALUES ($1, $2) RETURNING id`, m.Description, string(list))
	_ = row.Scan(&m.Id)
}

func getAllModes(db *sqlx.DB) interface{} {
	var (
		modes []Mode
	)
	rows, _ := db.Query(`SELECT id, description, listtl FROM public.modes`)
	for rows.Next() {
		var (
			temp    Mode
			listSrt string
		)
		_ = rows.Scan(&temp.Id, &temp.Description, &listSrt)
		_ = json.Unmarshal([]byte(listSrt), &temp.List)
		if len(temp.List) == 0 {
			temp.List = make([]ModeTL, 0)
		}
		modes = append(modes, temp)
	}
	if len(modes) == 0 {
		modes = make([]Mode, 0)
	}
	return modes
}
