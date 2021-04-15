package handlers

import (
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/deviceLog"
	u "github.com/ruraomsk/TLServer/internal/utils"
)

//DisplayDeviceLogFile обработчик отображения файлов лога устройства
var DisplayDeviceLogFile = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	db, id := data.GetDB()

	resp := deviceLog.DisplayDeviceLog(accInfo, db)
	data.FreeDB(id)
	data.CacheInfo.Mux.Lock()
	resp.Obj["regionInfo"] = data.CacheInfo.MapRegion
	resp.Obj["areaInfo"] = data.CacheInfo.MapArea
	data.CacheInfo.Mux.Unlock()
	u.SendRespond(c, resp)
}

//LogDeviceInfo обработчик запроса на выгрузку информации логов устройства за определенный период
var LogDeviceInfo = func(c *gin.Context) {
	arm := &deviceLog.LogDeviceInfo{}
	if err := c.ShouldBindJSON(&arm); err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "Invalid request"))
		return
	}
	db, id := data.GetDB()
	resp := deviceLog.DisplayDeviceLogInfo(*arm, db)
	data.FreeDB(id)
	u.SendRespond(c, resp)
}
