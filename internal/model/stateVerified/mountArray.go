package stateVerified

import (
	"fmt"
	valid "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

// MouthSetsVerified проверка месячных(годовых) карт
func MouthSetsVerified(cross *agspudge.Cross, empty IsEmpty) (result StateResult) {
	mouthSets := cross.Arrays.MonthSets
	weekSets := cross.Arrays.WeekSets
	result.SumResult = append(result.SumResult, "Проверка: Годовых карты")
	if len(mouthSets.MonthSets) > 12 {
		result.SumResult = append(result.SumResult, "Превышено количество годовых карт")
		result.Err = errors.New("detected")
		return
	}

	for _, mouth := range mouthSets.MonthSets {

		if valid.Validate(&mouth.Number, valid.Min(0), valid.Max(12)) != nil {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v): № недельной карты должен быть от 1 до 12", mouth.Number))
			result.Err = errors.New("detected")

		}
		flagFill := false

		for numMDay, mDay := range mouth.Days {
			//проверка целостности одной месячной карты
			if numMDay == 0 {
				if mDay == 0 {
					for _, zeroDay := range mouth.Days {
						if zeroDay != 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v): есть 0 позиция (%v)", mouth.Number, numMDay+1))
							result.Err = errors.New("detected")
							break
						}
					}
				}
			}
			if mDay != 0 {
				flagFill = true
			}
			if flagFill {
				if mDay == 0 {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v): есть 0 позиция (%v)", mouth.Number, numMDay+1))
					result.Err = errors.New("detected")
				}
				for numWeek, week := range weekSets.WeekSets {
					if week.Number == mDay {
						if !empty.Week[numWeek+1] {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) позиция (%v): недельная карта %v не заполнена", mouth.Number, numMDay+1, week.Number))
							result.Err = errors.New("detected")
						}
						break
					}
					if numWeek+1 == len(weekSets.WeekSets) {
						if mDay != 0 {
							result.SumResult = append(result.SumResult, fmt.Sprintf("Карта № (%v) позиция (%v): значение (%v) в недельных картах не найдено", mouth.Number, numMDay+1, mDay))
							result.Err = errors.New("detected")
						}
					}
				}
			}
		}
	}

	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Годовые карты: OK")
	}
	return
}
