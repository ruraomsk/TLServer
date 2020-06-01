package handlers

import (
	"net/http"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/serverLog"
	"github.com/JanFant/TLServer/logger"
	"github.com/gin-gonic/gin"

	u "github.com/JanFant/TLServer/internal/utils"
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

	mapContx := u.ParserInterface(c.Value("info"))
	resp := serverLog.DisplayServerFileLog(fileName, logger.LogGlobalConf.LogPath, mapContx, data.GetDB())
	u.SendRespond(c, resp)
}
