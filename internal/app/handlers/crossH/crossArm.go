package crossH

import (
	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/sockets/crossSock"
	u "github.com/ruraomsk/TLServer/internal/utils"
)

//ControlTestState обработчик проверки State
var ControlTestState = func(c *gin.Context) {
	accTK, _ := c.Get("tk")
	accInfo, _ := accTK.(*accToken.Token)
	resp := crossSock.TestCrossStateData(accInfo, data.GetDB())
	u.SendRespond(c, resp)
}
