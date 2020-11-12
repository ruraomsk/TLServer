package exchangeData

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/JanFant/TLServer/internal/sockets"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"net/http"
	"time"
)

type phasesInfo struct {
	Idevice      int       `json:"id"`                //id устройства
	Created      time.Time `json:"created"`           //когда создан план
	Updated      time.Time `json:"updated"`           //когда изменен план
	HwTacts      hwTact    `json:"hw_tacts"`          //нет информации
	IsAdaptive   bool      `json:"is_adaptive"`       //нет информации
	SecureTime   int       `json:"secure_time"`       //нет информации
	PhaseNum     int       `json:"phase_number"`      //номер ???
	ControllerId int       `json:"controller_id"`     //нет информации
	AdaptiveMax  int       `json:"adaptive_max_time"` //нет информации
	AdaptiveMin  int       `json:"adaptive_min_time"` //нет информации
}

type hwTact struct {
	Duration int `json:"duration"`
}

type dataPhase struct {
	Str string `xml:"dataphase"`
}

func GetPhases(iDevice []int, db *sqlx.DB) u.Response {
	var (
		PhasesList = make([]phasesInfo, 0)
	)

	query, args, err := sqlx.In(`SELECT idevice, region, area, id FROM public.cross WHERE idevice IN (?)`, iDevice)
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
			tempPhase phasesInfo
			idevice   int
			pos       sockets.PosInfo
		)
		err = rows.Scan(&idevice, &pos.Region, &pos.Area, &pos.Id)
		if err != nil {
			return u.Message(http.StatusInternalServerError, "error convert cross info")
		}

		//тут разбираем
		path := config.GlobalConfig.StaticPath + "//cross"
		file, _ := ioutil.ReadFile(path + fmt.Sprintf("//%v//%v//%v//cross.svg", pos.Region, pos.Area, pos.Id))
		var svgFile dataPhase
		_ = xml.Unmarshal(file, &svgFile)
		_ = json.Unmarshal([]byte(svgFile.Str), &tempPhase)
		tempPhase.Idevice = idevice

		PhasesList = append(PhasesList, tempPhase)
	}

	//хотят чтобы было уложено так через data/list (что поделать)
	globalData := make(map[string]interface{})
	globalData["data"] = PhasesList
	resp := u.Response{Code: http.StatusOK, Obj: map[string]interface{}{}}
	resp.Obj["list"] = globalData
	return resp
}
