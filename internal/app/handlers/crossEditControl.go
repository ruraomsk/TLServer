package handlers

import (
	"github.com/JanFant/TLServer/internal/model/data"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

//CrossEditInfo сбор информации о занятых перекрестках
var CrossEditInfo = func(c *gin.Context) {
	mapContx := u.ParserInterface(c.Value("info"))
	resp := data.DisplayCrossEditInfo(mapContx)
	data.CacheInfo.Mux.Lock()
	resp.Obj["regionInfo"] = data.CacheInfo.MapRegion
	resp.Obj["areaInfo"] = data.CacheInfo.MapArea
	data.CacheInfo.Mux.Unlock()
	u.SendRespond(c, resp)
}

//CrossEditFree освобождение перекрестков
var CrossEditFree = func(c *gin.Context) {
	var discDev data.CrossDisc
	if err := c.ShouldBindJSON(&discDev); err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	resp := data.CrossEditFree(discDev)
	u.SendRespond(c, resp)
}
