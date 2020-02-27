package stateVerified

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/ruraomsk/ag-server/binding"
)

type StateResult struct {
	SumResult []string
	Err       error
}

//DaySetsVerified проверка суточных карт
func DaySetsVerified(sets *binding.DaySets) (result StateResult) {
	result.SumResult = append(result.SumResult, "Проверка: Суточные карты")
	if len(sets.DaySets) > 12 {
		result.SumResult = append(result.SumResult, "Превышено количество суточных карт")
		result.Err = errors.New("detected")
		return
	}
	for numDay, day := range sets.DaySets {
		if day.Number > 12 || day.Number < 1 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верный номер суточной карты: %v", day.Number))
			result.Err = errors.New("detected")
		}
		lineCount := 0
		flagZero := false
		for numLine, line := range day.Lines {
			if line.Hour > 24 || line.Hour < 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): значение часа = %v", numDay+1, numLine+1, line.Hour))
				result.Err = errors.New("detected")
			}
			if line.Min > 59 || line.Min < 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): значение минуты = %v", numDay+1, numLine+1, line.Min))
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
			//-----------

			if line.PKNom != 0 {
				if numLine != 11 {
					if line.Hour > sets.DaySets[numDay].Lines[numLine+1].Hour && sets.DaySets[numDay].Lines[numLine+1].PKNom != 0 {
						result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): текущее значение времени часа %v больше следующего %v", numDay+1, numLine+1, line.Hour, sets.DaySets[numDay].Lines[numLine+1].Hour))
						result.Err = errors.New("detected")
					}
					if line.Hour == sets.DaySets[numDay].Lines[numLine+1].Hour {
						if line.Min >= sets.DaySets[numDay].Lines[numLine+1].Min {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) стр. № (%v): текущее значение времени минуты %v больше следующего %v", numDay+1, numLine+1, line.Min, sets.DaySets[numDay].Lines[numLine+1].Min))
							result.Err = errors.New("detected")
						}
					}
				} else {
					if sets.DaySets[numDay].Lines[numLine].Hour != 24 && sets.DaySets[numDay].Lines[numLine].Min != 0 {
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
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Суточные карты: OK")
	}
	return
}
