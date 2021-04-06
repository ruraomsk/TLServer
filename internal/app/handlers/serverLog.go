package handlers

import (
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/serverLog"
	"github.com/ruraomsk/TLServer/logger"

	u "github.com/ruraomsk/TLServer/internal/utils"
)

//DisplayServerLogFile обработчик отображения файлов лога сервера
var DisplayServerLogFile = func(c *gin.Context) {
	resp := serverLog.DisplayServerLogFiles(logger.LogGlobalConf.LogPath)
	u.SendRespond(c, resp)
}

//DisplayServerLogInfo обработчик выгрузки содержимого лог файла сервера
var DisplayServerLogInfo = func(c *gin.Context) {
	fileName := c.Query("fileName")
	if fileName == "" {
		u.SendRespond(c, u.Message(http.StatusBadRequest, "Blank field: fileName"))
		return
	}

	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := serverLog.DisplayServerFileLog(fileName, logger.LogGlobalConf.LogPath, accInfo, data.GetDB("DisplayServerLogInfo"))
	data.FreeDB("DisplayServerLogInfo")
	u.SendRespond(c, resp)
}
