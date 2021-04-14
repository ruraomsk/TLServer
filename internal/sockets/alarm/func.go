package alarm

import (
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

//getCross запроса состояния перекрестков
func getCross(reg int) []*CrossInfo {
	var crosses = make([]*CrossInfo, 0)
	db := data.GetDB("getCrossAlarm")
	defer data.FreeDB("getCrossAlarm")
	w := fmt.Sprintf("SELECT device,state FROM public.devices LEFT JOIN public.\"cross\" ON public.devices.id = public.\"cross\".idevice WHERE public.\"cross\".region=%d;", reg)
	rows, err := db.Query(w)
	if err != nil {
		logger.Error.Println("|IP: server |Login: server |Resource: /alarm |Message: Error get data from BD ", err.Error())
		return make([]*CrossInfo, 0)
	}
	var dev []byte
	var state []byte
	var ctrl pudge.Controller
	var cross pudge.Cross

	for rows.Next() {
		temp := new(CrossInfo)
		_ = rows.Scan(&dev, &state)
		err = json.Unmarshal(dev, &ctrl)
		if err != nil {
			logger.Error.Println("|IP: server |Login: server |Resource: /alarm  |Message: Error get device from BD ", err.Error())
			return make([]*CrossInfo, 0)
		}
		err = json.Unmarshal(state, &cross)
		if err != nil {
			logger.Error.Println("|IP: server |Login: server |Resource: /techArm |Message: Error get cross from BD ", err.Error())
			return make([]*CrossInfo, 0)
		}
		temp.Region = cross.Region
		temp.Area = cross.Area
		temp.Subarea = cross.SubArea
		temp.ID = cross.ID
		temp.Idevice = cross.IDevice
		temp.StatusCode = cross.StatusDevice
		temp.Time = ctrl.LastOperation
		temp.Describe = cross.Name
		temp.Status = data.CacheInfo.MapTLSost[temp.StatusCode].Description
		temp.Control = data.CacheInfo.MapTLSost[temp.StatusCode].Control
		//logger.Debug.Printf("temp %v",temp)
		crosses = append(crosses, temp)
	}
	return crosses
}
