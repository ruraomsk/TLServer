package crossH

import (
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//ControlTestState обработчик проверки State
var ControlTestState = func(c *gin.Context) {
	mapContx := u.ParserInterface(c.Value("info"))
	resp := crossSock.TestCrossStateData(mapContx, data.GetDB())
	u.SendRespond(c, resp)
}
