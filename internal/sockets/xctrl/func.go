package xctrl

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/xcontrol"
)

//getXctrl формирует массив записей из таблицы xctrl
func getXctrl(db *sqlx.DB) ([]xcontrol.State, error) {
	rows, err := db.Query(`SELECT state FROM public.xctrl`)
	if err != nil {
		return nil, err
	}
	var allXctrl []xcontrol.State
	for rows.Next() {
		var (
			strXctrl string
			temp     xcontrol.State
		)
		err := rows.Scan(&strXctrl)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(strXctrl), &temp)
		allXctrl = append(allXctrl, temp)
	}
	return allXctrl, nil
}
