package stateVerified

import (
	"fmt"
	"github.com/pkg/errors"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
)

type IsEmpty struct {
	Week map[int]bool
}

// WeekSetsVerified проверка недельных карт
func WeekSetsVerified(cross *agS_pudge.Cross) (result StateResult, Empty IsEmpty) {
	Empty.Week = make(map[int]bool)
	weekSets := cross.Arrays.WeekSets
	daySets := cross.Arrays.DaySets
	result.SumResult = append(result.SumResult, "Проверка: Недельные карты")
	if len(weekSets.WeekSets) > 12 {
		result.SumResult = append(result.SumResult, "Превышено количество недельных карт")
		result.Err = errors.New("detected")
		return
	}
	for _, week := range weekSets.WeekSets {
		if week.Number > 12 || week.Number < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты неделя № (%v): № недельной карты должен быть от 1 до 12", week.Number))
			result.Err = errors.New("detected")
		}
		flagFill := false
		for numWDay, wDay := range week.Days {
			//проверка целостности одной недельной карты
			if numWDay == 0 {
				if wDay == 0 {
					for _, zeroDay := range week.Days {
						if zeroDay != 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты неделя № (%v): в неделе есть 0 позиция (%v)", week.Number, numWDay+1))
							result.Err = errors.New("detected")
							break
						}
					}
				}
			}
			if wDay != 0 {
				flagFill = true
			}
			if flagFill {
				if wDay == 0 {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты неделя № (%v): в неделе есть 0 позиция (%v)", week.Number, numWDay+1))
					result.Err = errors.New("detected")
				}
				for numDay, day := range daySets.DaySets {
					if day.Number == wDay {
						if day.Count == 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты неделя № (%v) позиция (%v): дневная карта %v не заполнена", week.Number, numWDay+1, day.Number))
							result.Err = errors.New("detected")
						}
						break
					}
					if numDay+1 == len(daySets.DaySets) {
						if wDay != 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты неделя № (%v) позиция (%v): значение (%v) в дневных картах не найдено", week.Number, numWDay+1, wDay))
							result.Err = errors.New("detected")
						}
					}
				}
			}
		}
		Empty.Week[week.Number] = flagFill
	}

	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Недельные карты: OK")
	}
	return
}
