package controlCross

import "github.com/jmoiron/sqlx"

//get запрос фазы из базы
func (p *phaseInfo) get(db *sqlx.DB) error {
	err := db.QueryRow(`SELECT fdk, tdk, pdk FROM public.devices WHERE id = $1`, p.idevice).Scan(&p.Fdk, &p.Tdk, &p.Pdk)
	if err != nil {
		return err
	}
	return nil
}

////formCrossUser сформировать пользователей которые редактируеют кросы
//func formCrossUser() []CrossInfo {
//	var temp = make([]CrossInfo, 0)
//	for _, info := range crossConnect {
//		if info.Edit {
//			temp = append(temp, info)
//		}
//	}
//	return temp
//}
