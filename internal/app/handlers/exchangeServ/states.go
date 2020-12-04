package exchangeServ

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/exchangeData"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"
)

//StatesHandler обработчик запроса фаз
func StatesHandler(c *gin.Context, db *sqlx.DB) {
	iDevicesStr := c.QueryArray("controller_id")
	if len(iDevicesStr) <= 0 {
		resp := u.Message(http.StatusBadRequest, "blank field: controller_id")
		c.JSON(resp.Code, resp.Obj)
		return
	}
	var iDevices []int
	for _, devStr := range iDevicesStr {
		id, err := strconv.Atoi(devStr)
		if err != nil {
			resp := u.Message(http.StatusBadRequest, fmt.Sprintf("controller_id = | %v | must be int", devStr))
			c.JSON(resp.Code, resp.Obj)
			return
		}
		iDevices = append(iDevices, id)
	}
	resp := exchangeData.GetStates(iDevices, db)
	c.JSON(resp.Code, resp.Obj)
}
