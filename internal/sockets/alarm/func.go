package alarm

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/logger"
	"time"
)

//getCross запроса состояния перекрестков
func getCross(reg int, db *sqlx.DB) []*CrossInfo {
	var (
		crosses = make([]*CrossInfo, 0)
		sqlStr  = `SELECT region,
 					area, 
 					id,
  					idevice, 
  					describ, 
  					subarea, 
       				state ->'status'
  					FROM public.cross`
	)
	if reg != -1 {
		sqlStr += fmt.Sprintf(` WHERE region = %v`, reg)
	}
	rows, err := db.Query(sqlStr)
	if err != nil {
		logger.Error.Println("|IP: server |Login: server |Resource: /techArm |Message: Error get Cross from BD ", err.Error())
		return make([]*CrossInfo, 0)
	}
	for rows.Next() {
		temp := new(CrossInfo)
		_ = rows.Scan(&temp.Region,
			&temp.Area,
			&temp.ID,
			&temp.Idevice,
			&temp.Describe,
			&temp.Subarea,
			&temp.StatusCode)
		temp.Time = time.Now()
		temp.Status = data.CacheInfo.MapTLSost[temp.StatusCode].Description
		temp.Control = data.CacheInfo.MapTLSost[temp.StatusCode].Control
		crosses = append(crosses, temp)
	}
	return crosses
}
