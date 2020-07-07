package stateVerified

import (
	"fmt"
	valid "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"strings"
)

//StateResult информация о прорке данных
type StateResult struct {
	SumResult []string //накоплении информации о ходе проверки
	Err       error    //признак ошибки при проверке
}

//DaySetsVerified проверка суточных карт
func DaySetsVerified(cross *agspudge.Cross) (result StateResult) {
	daySets := cross.Arrays.DaySets
	dkSets := cross.Arrays.SetDK
	result.SumResult = append(result.SumResult, "Проверка: Суточные карты")
	if len(daySets.DaySets) > 12 {
		result.SumResult = append(result.SumResult, "Превышено количество суточных карт")
		result.Err = errors.New("detected")
		return
	}
	for numDay, day := range daySets.DaySets {
		if valid.Validate(&day.Number, valid.Min(1), valid.Max(12)) != nil {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верный номер суточной карты: %v", day.Number))
			result.Err = errors.New("detected")
		}
		lineCount := 0
		flagZero := false
		for numLine, line := range day.Lines {

			valRes := valid.ValidateStruct(line,
				valid.Field(&line.Hour, valid.Min(0), valid.Max(24)),
				valid.Field(&line.Min, valid.Min(0), valid.Max(59)),
			)
			if valRes != nil {
				if strings.Contains(valRes.Error(), "hour") {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): значение часа = %v", numDay+1, numLine+1, line.Hour))
				}
				if strings.Contains(valRes.Error(), "min") {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): значение минуты = %v", numDay+1, numLine+1, line.Min))
				}
				result.Err = errors.New("detected")
			}

			if line.Hour == 24 && line.Min != 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): значение минуты = %v должно быть 0", numDay+1, numLine+1, line.Min))
				result.Err = errors.New("detected")
			}
			if line.PKNom == 0 && (line.Hour != 0 || line.Min != 0) {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): №ПК = %v, время должно быть 00:00", numDay+1, numLine+1, line.PKNom))
				result.Err = errors.New("detected")
			}

			if line.PKNom < 0 || line.PKNom > 12 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): №ПК должен быть от 0 до 12", numDay+1, numLine+1))
				result.Err = errors.New("detected")
			} else {
				//-----------
				if line.PKNom != 0 {
					if dkSets.IsEmpty(1, line.PKNom) {
						result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): нельзя искользовать пустой ПК(%v)", numDay+1, numLine+1, line.PKNom))
						result.Err = errors.New("detected")
					}
					if numLine != 11 {
						if line.Hour > daySets.DaySets[numDay].Lines[numLine+1].Hour && daySets.DaySets[numDay].Lines[numLine+1].PKNom != 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): текущее значение времени часа %v больше следующего %v", numDay+1, numLine+1, line.Hour, daySets.DaySets[numDay].Lines[numLine+1].Hour))
							result.Err = errors.New("detected")
						}
						if line.Hour == daySets.DaySets[numDay].Lines[numLine+1].Hour {
							if line.Min >= daySets.DaySets[numDay].Lines[numLine+1].Min {
								result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): текущее значение времени минуты %v больше следующего %v", numDay+1, numLine+1, line.Min, daySets.DaySets[numDay].Lines[numLine+1].Min))
								result.Err = errors.New("detected")
							}
						}
					} else {
						if daySets.DaySets[numDay].Lines[numLine].Hour != 24 || daySets.DaySets[numDay].Lines[numLine].Min != 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): текущее значение последнего времени должно быть 24:00", numDay+1, numLine+1))
							result.Err = errors.New("detected")
						}
					}
					if flagZero {
						if line.Hour != 0 || line.Min != 0 || line.PKNom != 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): значение времени должно быть 00:00 и #ПК 0", numDay+1, numLine+1))
							result.Err = errors.New("detected")
						}
					}
				} else {
					flagZero = true
					if numLine-1 >= 0 {
						if daySets.DaySets[numDay].Lines[numLine-1].PKNom != 0 {
							if daySets.DaySets[numDay].Lines[numLine-1].Hour != 24 || daySets.DaySets[numDay].Lines[numLine-1].Min != 0 {
								result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): текущее значение последнего времени должно быть 24:00", numDay+1, numLine))
								result.Err = errors.New("detected")
							}
						}
					}
				}
			}
			//-----------
			if line.PKNom != 0 {
				lineCount++
			}
		}
		if lineCount != day.Count {
			daySets.DaySets[numDay].Count = lineCount
		}
	}
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Суточные карты: OK")
	}
	return
}
