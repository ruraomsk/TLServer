package exchangeData

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/model/config"
	"github.com/ruraomsk/TLServer/internal/model/data"
	u "github.com/ruraomsk/TLServer/internal/utils"
	"io/ioutil"
	"net/http"
)

type svgData struct {
	Idevice  int    `json:"idevice"`  //idevice устройтсва
	SvgBytes []byte `json:"svgBytes"` //массив байт svg
}

//GetSvgs обработчик запроса svg
func GetSvgs(iDevice []int) u.Response {
	var (
		SvgsList = make([]svgData, 0)
	)
	db, id := data.GetDB()
	defer data.FreeDB(id)
	query, args, err := sqlx.In(`SELECT region, area, id, idevice FROM public.cross WHERE idevice IN (?)`, iDevice)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "error formatting IN query")
	}
	query = db.Rebind(query)
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return u.Message(http.StatusInternalServerError, "db server not response")
	}

	for rows.Next() {
		var (
			tempSvgData      svgData
			tReg, tArea, tId string
		)
		err = rows.Scan(&tReg, &tArea, &tId, &tempSvgData.Idevice)
		if err != nil {
			return u.Message(http.StatusInternalServerError, "error convert cross info")
		}

		path := config.GlobalConfig.StaticPath + fmt.Sprintf("//cross//%v//%v//%v//cross.svg", tReg, tArea, tId)
		rawByte, err := ioutil.ReadFile(path)
		if err != nil {
			tempSvgData.SvgBytes = make([]byte, 0)
		} else {
			tempSvgData.SvgBytes = rawByte
		}

		SvgsList = append(SvgsList, tempSvgData)
	}
	flag := false
	for _, idev := range iDevice {
		flag = false
		for _, svgInfo := range SvgsList {
			if svgInfo.Idevice == idev {
				flag = true
				break
			}
		}
		if !flag {
			SvgsList = append(SvgsList, svgData{Idevice: idev, SvgBytes: make([]byte, 0)})
		}
	}

	resp := u.Response{Code: http.StatusOK, Obj: map[string]interface{}{}}
	resp.Obj["data"] = SvgsList
	return resp
}
