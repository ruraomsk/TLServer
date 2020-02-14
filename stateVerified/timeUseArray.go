package stateVerified

import (
	"fmt"
	"github.com/pkg/errors"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
	"strconv"
	"strings"
)

// TimeUseVerified проверка недельных карт
func TimeUseVerified(cross *agS_pudge.Cross) (result StateResult) {
	timeUse := cross.Arrays.SetTimeUse
	result.SumResult = append(result.SumResult, "Проверка: Внешние выходы")
	for numUses, uses := range timeUse.Uses {
		if uses.Type > 1 || uses.Type < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): тип стат. должен быть 0 или 1", uses.Name))
			result.Err = errors.New("detected")
		}
		if uses.Tvps > 6 || uses.Tvps < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): ТВП1,2, МГР, ВПУ должен быть от 0 до 6", uses.Name))
			result.Err = errors.New("detected")
		}
		if uses.Dk > 1 || uses.Dk < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): № ДК должен быть 0 или 1", uses.Name))
			result.Err = errors.New("detected")
		}
		if uses.Dk == 0 {
			if uses.Type != 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Если № ДК = 0 поле тип стат. должно быть 0", uses.Name))
				result.Err = errors.New("detected")
			}
			if uses.Type != 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Если № ДК = 0 ТВП1,2, МГР, ВПУ должно быть 0", uses.Name))
				result.Err = errors.New("detected")
			}
			if uses.Long != 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v):Если № ДК = 0 интервала должен быть 0", uses.Name))
				result.Err = errors.New("detected")
			}
			uses.Fazes = strings.TrimSpace(uses.Fazes)
			if uses.Fazes != "" {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Если № ДК = 0 значение фазы должна быть пустая строка", uses.Name))
				result.Err = errors.New("detected")
			}
		}
		if uses.Dk == 1 {
			if uses.Long < 0 {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Интервал должен быть ноль или больше", uses.Name))
				result.Err = errors.New("detected")
			} else if uses.Long == 0 {
				if numUses > 1 {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Интервал должен быть больше нуля", uses.Name))
					result.Err = errors.New("detected")
				}
			}

			tempFazes := strings.Split(uses.Fazes, ",")
			for _, faze := range tempFazes {
				faze = strings.TrimSpace(faze)
				fazeInt, err := strconv.Atoi(faze)
				if err != nil {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Значение фазы должно быть числом указанным через запятую `,`", uses.Name))
					result.Err = errors.New("detected")
				}
				if fazeInt <= 0 {
					result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Значение фазы должно быть больше нуля", uses.Name))
					result.Err = errors.New("detected")
				}
			}
		}

	}
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Внешние выходы: OK")
	}
	return
}
