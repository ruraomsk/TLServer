package handlers

import (
	"github.com/JanFant/TLServer/internal/model/accToken"
	"net/http"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/deviceLog"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//DisplayDeviceLogFile обработчик отображения файлов лога устройства
var DisplayDeviceLogFile = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := deviceLog.DisplayDeviceLog(accInfo, data.GetDB())
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

	resp := deviceLog.DisplayDeviceLogInfo(*arm, data.GetDB())
	u.SendRespond(c, resp)
}
