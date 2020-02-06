package data

import (
	"../logger"
	"../tcpConnect"
	u "../utils"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/ruraomsk/ag-server/binding"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"os"
	"strings"
)

type stateResult struct {
	SumResult []string
	err       error
}

//ControlGetCrossInfo сбор информации для пользователя в расширенном варианте
func ControlGetCrossInfo(TLignt TrafficLights) map[string]interface{} {
	var (
		dgis     string
		sqlStr   string
		StateStr string
	)
	sqlStr = fmt.Sprintf("select area, subarea, idevice, dgis, describ, state from %v where region = %v and id = %v and area = %v", os.Getenv("gis_table"), TLignt.Region.Num, TLignt.ID, TLignt.Area.Num)
	rowsTL := GetDB().Raw(sqlStr).Row()
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &StateStr)
	if err != nil {
		logger.Error.Println("|Message: No result at these points", err.Error())
		return u.Message(false, "No result at these points")
	}
	TLignt.Points.StrToFloat(dgis)
	TLignt.Region.NameRegion = CacheInfo.mapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.mapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(StateStr)
	if err != nil {
		logger.Error.Println("|Message: Failed to parse cross information: ", err.Error())
		return u.Message(false, "Failed to parse cross information")
	}
	TLignt.Sost.Num = rState.StatusDevice
	TLignt.Sost.Description = CacheInfo.mapTLSost[TLignt.Sost.Num]
	resp := u.Message(true, "Cross information")
	resp["cross"] = TLignt
	resp["state"] = rState

	resp["areaMap"] = CacheInfo.mapArea[TLignt.Region.NameRegion]

	return resp
}

//SendCrossData получение данных от пользователя проверка и отправка серверу(устройств)
func SendCrossData(state agS_pudge.Cross, mapContx map[string]string) map[string]interface{} {
	var (
		stateMessage tcpConnect.StateMessage
		err          error
	)
	verif := verifiedState(&state)
	if verif.err != nil {
		resp := u.Message(false, "Data didn't pass verification")
		resp["result"] = verif.SumResult
		return resp
	}
	stateMessage.StateStr, err = stateMarshal(state)
	if err != nil {
		logger.Error.Println("|Message: Failed to Marshal state information: ", err.Error())
		return u.Message(false, "Failed to Marshal state information")
	}
	stateMessage.User = mapContx["login"]
	tcpConnect.StateChan <- stateMessage
	for {
		chanRespond := <-tcpConnect.StateChan
		if strings.Contains(chanRespond.User, mapContx["login"]) {
			if chanRespond.Message == "ok" {
				return u.Message(true, "Cross data send to server")
			} else {
				return u.Message(false, "TCP Server not responding")
			}
		}
	}
}

//CheckCrossData проверка полученных данных
func CheckCrossData(state agS_pudge.Cross) map[string]interface{} {
	verif := verifiedState(&state)
	if verif.err != nil {
		resp := u.Message(false, "Data didn't pass verification")
		resp["result"] = verif.SumResult
		return resp
	}
	resp := u.Message(true, "Data is correct")
	resp["result"] = verif.SumResult
	return resp
}

//stateMarshal преобразовать структуру в строку
func stateMarshal(cross agS_pudge.Cross) (str string, err error) {
	newByte, err := json.Marshal(cross)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}

//verifiedState набор проверкок для стейта
func verifiedState(cross *agS_pudge.Cross) (result stateResult) {
	result = daySetsVerified(&cross.Arrays.DaySets)
	return
}

//daySetsVerified проверка суточных карт
func daySetsVerified(sets *binding.DaySets) (result stateResult) {
	result.SumResult = append(result.SumResult, "Проверка: Суточные карты")
	if len(sets.DaySets) > 12 {
		result.SumResult = append(result.SumResult, "Превышено количество суточных карт")
		result.err = errors.New("detected")
		return
	}
	for numDay, day := range sets.DaySets {
		if day.Number > 12 || day.Number < 1 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верный номер суточной карты: %v", day.Number))
			result.err = errors.New("detected")
		}
		lineCount := 0
		for numLine, line := range day.Lines {
			if line.Hour > 24 || line.Hour < 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты № (%v) стр. № (%v): значение часа = %v", numDay+1, numLine+1, line.Hour))
				result.err = errors.New("detected")
			}
			if line.Min > 59 || line.Min < 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты № (%v) стр. № (%v): значение минуты = %v", numDay+1, numLine+1, line.Min))
				result.err = errors.New("detected")
			}
			if line.PKNom == 0 && (line.Hour != 0 || line.Min != 0) {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно заполнено значение карты № (%v) стр. № (%v): №ПК = %v, время должно быть 00:00", numDay+1, numLine+1, line.PKNom))
				result.err = errors.New("detected")
			}
			//-----------
			if line.PKNom != 0 {
				if numLine != 12 {
					if line.Hour > sets.DaySets[numDay].Lines[numLine+1].Hour && sets.DaySets[numDay].Lines[numLine+1].PKNom != 0 {
						result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно заполнено значение карты № (%v) стр. № (%v): текущее значение времени часа %v больше следующего %v", numDay+1, numLine+1, line.Hour, sets.DaySets[numDay].Lines[numLine+1].Hour))
						result.err = errors.New("detected")
					}
					if line.Hour == sets.DaySets[numDay].Lines[numLine+1].Hour {
						if line.Min >= sets.DaySets[numDay].Lines[numLine+1].Min {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно заполнено значение карты № (%v) стр. № (%v): текущее значение времени минуты %v больше следующего %v", numDay+1, numLine+1, line.Min, sets.DaySets[numDay].Lines[numLine+1].Min))
							result.err = errors.New("detected")
						}
					}
				} else {
					if sets.DaySets[numDay].Lines[numLine-1].Hour != 24 && sets.DaySets[numDay].Lines[numLine-1].Min != 0 {
						result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно заполнено значение карты № (%v) стр. № (%v): текущее значение последнего премени должно быть 24:00", numDay+1, numLine+1))
						result.err = errors.New("detected")
					}
				}
			}
			//-----------
			if line.PKNom != 0 {
				lineCount++
			}
		}
		if lineCount != day.Count {
			sets.DaySets[numDay].Count = lineCount
		}
	}
	if result.err == nil {
		result.SumResult = append(result.SumResult, "Суточные карты: OK")
	}
	return
}
