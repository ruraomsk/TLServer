package exchangeServ

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ruraomsk/TLServer/internal/model/exchangeData"
	u "github.com/ruraomsk/TLServer/internal/utils"
	"net/http"
	"strconv"
)

//SvgsHandler обработчик запроса svg
func SvgsHandler(c *gin.Context) {
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
	resp := exchangeData.GetSvgs(iDevices)
	c.JSON(resp.Code, resp.Obj)
}
