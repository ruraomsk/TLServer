package crossH

import (
	"github.com/JanFant/TLServer/internal/model/accToken"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
)

//ControlTestState обработчик проверки State
var ControlTestState = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := crossSock.TestCrossStateData(accInfo, data.GetDB())
	u.SendRespond(c, resp)
}
