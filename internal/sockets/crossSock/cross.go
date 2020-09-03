package crossSock

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/accToken"
	"github.com/JanFant/TLServer/internal/model/deviceLog"
	"github.com/JanFant/TLServer/internal/model/stateVerified"
	"github.com/JanFant/TLServer/internal/sockets"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/JanFant/TLServer/logger"
	"github.com/jmoiron/sqlx"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"net/http"
	"strconv"
)

//crossInfo информация о перекрестке для которого открыт сокет
type CrossInfo struct {
	Edit        bool            `json:"edit"`        //признак редактирования
	Idevice     int             `json:"idevice"`     //идентификатор утройства
	Pos         sockets.PosInfo `json:"pos"`         //расположение перекрестка
	Description string          `json:"description"` //описание
	Login       string          `json:"login"`
	AccInfo     *accToken.Token `json:"-"`
}

var GetCrossUsersForDisplay chan bool
var GetArmUsersForDisplay chan bool
var CrArmUsersForDisplay chan []CrossInfo
var CrossUsersForDisplay chan []CrossInfo
var DiscCrossUsers chan []CrossInfo
var DiscArmUsers chan []CrossInfo

//GetNewState получение обновленного state
func GetNewState(pos sockets.PosInfo, db *sqlx.DB) (agspudge.Cross, error) {
	var stateStr string
	rowsTL := db.QueryRow(`SELECT state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	_ = rowsTL.Scan(&stateStr)
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		return agspudge.Cross{}, err
	}
	return rState, nil
}

//ConvertStateStrToStruct разбор данных (Cross) полученных из БД в нужную структуру
func ConvertStateStrToStruct(str string) (rState agspudge.Cross, err error) {
	if err := json.Unmarshal([]byte(str), &rState); err != nil {
		return rState, err
	}
	return rState, nil
}

//TestCrossStateData проверить все стрейты на наличие ошибок
func TestCrossStateData(accInfo *accToken.Token, db *sqlx.DB) u.Response {
	var (
		stateSql  string
		stateInfo []deviceLog.BusyArm
		state     deviceLog.BusyArm
	)
	sqlStr := fmt.Sprintf(`SELECT state FROM public.cross `)
	if accInfo.Region != "*" {
		sqlStr += fmt.Sprintf(`WHERE region = %v `, accInfo.Region)
	}
	sqlStr += "order by describ"
	rows, err := db.Query(sqlStr)
	if err != nil {
		resp := u.Message(http.StatusInternalServerError, "DB not respond")
		return resp
	}
	for rows.Next() {
		_ = rows.Scan(&stateSql)
		testState, err := ConvertStateStrToStruct(stateSql)
		if err != nil {
			logger.Error.Println("|Message: Failed to parse cross information: ", err.Error())
			return u.Message(http.StatusInternalServerError, "failed to parse cross information")
		}
		var verif stateVerified.StateResult
		VerifiedState(&testState, &verif, db)
		if verif.Err != nil {
			state.ID = testState.ID
			state.Region = strconv.Itoa(testState.Region)
			state.Area = strconv.Itoa(testState.Area)
			state.Description = testState.Name
			stateInfo = append(stateInfo, state)
		}
	}
	resp := u.Message(http.StatusOK, "state data")
	resp.Obj["arms"] = stateInfo
	return resp
}

//VerifiedState набор проверкок для стейта
func VerifiedState(cross *agspudge.Cross, result *stateVerified.StateResult, db *sqlx.DB) {
	resultMainWind := stateVerified.MainWindVerified(cross, db)
	appendResult(result, resultMainWind)
	resultDay := stateVerified.DaySetsVerified(cross)
	appendResult(result, resultDay)
	resultWeek, empty := stateVerified.WeekSetsVerified(cross)
	appendResult(result, resultWeek)
	resultMouth := stateVerified.MouthSetsVerified(cross, empty)
	appendResult(result, resultMouth)
	resultTimeUse := stateVerified.TimeUseVerified(cross)
	appendResult(result, resultTimeUse)
	resultCtrl := stateVerified.CtrlVerified(cross)
	appendResult(result, resultCtrl)
	return
}

//appendResult накапливание результатов верификации
func appendResult(mainRes *stateVerified.StateResult, addResult stateVerified.StateResult) {
	mainRes.SumResult = append(mainRes.SumResult, addResult.SumResult...)
	if addResult.Err != nil {
		mainRes.Err = addResult.Err
	}
}
