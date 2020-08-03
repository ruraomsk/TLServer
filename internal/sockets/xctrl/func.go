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

//writeXctrl запис массива state в базу
func writeXctrl(states []xcontrol.State, db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, state := range states {
		strState, _ := json.Marshal(state)
		_, err := tx.Exec(`UPDATE public.xctrl SET state = $1 WHERE region = $2 AND area = $3 AND subarea = $4`, string(strState), state.Region, state.Area, state.SubArea)
		if err != nil {
			err = tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
