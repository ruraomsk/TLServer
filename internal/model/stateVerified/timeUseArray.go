package stateVerified

import (
	"fmt"
	valid "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"strconv"
	"strings"
)

// TimeUseVerified проверка недельных карт
func TimeUseVerified(cross *agspudge.Cross) (result StateResult) {
	timeUse := cross.Arrays.SetTimeUse
	result.SumResult = append(result.SumResult, "Проверка: Внешние выходы")
	for _, uses := range timeUse.Uses {
		validRes := valid.ValidateStruct(&uses,
			valid.Field(&uses.Type, valid.Min(0), valid.Max(1)),
			valid.Field(&uses.Tvps, valid.Min(0), valid.Max(9)),
			valid.Field(&uses.Dk, valid.Min(0), valid.Max(1)),
		)
		if validRes != nil {
			if strings.Contains(validRes.Error(), "type") {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): тип стат. должен быть 0 или 1", uses.Name))
			}
			if strings.Contains(validRes.Error(), "tvps") {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): ТВП1,2, МГР, ВПУ должен быть от 0 до 9", uses.Name))
			}
			if strings.Contains(validRes.Error(), "dk") {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): № ДК должен быть 0 или 1", uses.Name))
			}
			result.Err = errors.New("detected")
		}

		if uses.Tvps == 0 {
			//if uses.Type != 0 {
			//	result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Если № ТВП = 0 поле тип стат. должно быть 0", uses.Name))
			//	result.Err = errors.New("detected")
			//}
			if uses.Long != 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v):Если № ТВП = 0 интервала должен быть 0", uses.Name))
				result.Err = errors.New("detected")
			}
			uses.Fazes = strings.TrimSpace(uses.Fazes)
			//if uses.Fazes != "" {
			//	result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Если № ТВП = 0 значение фазы должна быть пустая строка", uses.Name))
			//	result.Err = errors.New("detected")
			//}
		}

		if uses.Tvps > 0 {
			if uses.Long < 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Интервал должен быть ноль или больше", uses.Name))
				result.Err = errors.New("detected")
			} else if uses.Long == 0 {
				//if numUses > 1 {
				//	result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Интервал должен быть больше нуля", uses.Name))
				//	result.Err = errors.New("detected")
				//}
			}

			tempFazes := strings.Split(uses.Fazes, ",")
			for _, faze := range tempFazes {
				faze = strings.TrimSpace(faze)
				_, err := strconv.Atoi(faze)
				if err != nil {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Значение фазы должно быть числом указанным через запятую `,`", uses.Name))
					result.Err = errors.New("detected")
				}
				//if fazeInt <= 0 {
				//	result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Значение фазы должно быть больше нуля", uses.Name))
				//	result.Err = errors.New("detected")
				//}
			}
		}

	}
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Внешние выходы: OK")
	}
	return
}
