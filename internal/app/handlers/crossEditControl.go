package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/sockets/crossSock"
	u "github.com/ruraomsk/TLServer/internal/utils"
	"net/http"
)

//CrossEditInfo сбор информации о занятых перекрестках
var CrossEditInfo = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := crossSock.DisplayCrossEditInfo(accInfo)
	data.CacheInfo.Mux.Lock()
	resp.Obj["regionInfo"] = data.CacheInfo.MapRegion
	resp.Obj["areaInfo"] = data.CacheInfo.MapArea
	data.CacheInfo.Mux.Unlock()
	u.SendRespond(c, resp)
}

//CrossEditFree освобождение перекрестков
var CrossEditFree = func(c *gin.Context) {
	var discDev crossSock.CrossDisc
	if err := c.ShouldBindJSON(&discDev); err != nil {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "invalid request"))
		return
	}
	resp := crossSock.CrossEditFree(discDev)
	u.SendRespond(c, resp)
}
