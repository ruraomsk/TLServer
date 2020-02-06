package stateVerified

import (
	"fmt"
	"github.com/pkg/errors"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
)

// TimeUseVerified проверка недельных карт
func TimeUseVerified(cross *agS_pudge.Cross) (result StateResult) {
	timeUse := cross.Arrays.SetTimeUse
	result.SumResult = append(result.SumResult, "Проверка: Внешние выходы")
	for _, uses := range timeUse.Uses {
		if uses.Type > 1 || uses.Type < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты внешние выходы (%v): тип стат. должен быть 0 или 1", uses.Name))
			result.Err = errors.New("detected")
		}
		if uses.Tvps > 4 || uses.Tvps < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты внешние выходы (%v): ТВП1,2, МГР, ВПУ должен быть от 0 до 4", uses.Name))
			result.Err = errors.New("detected")
		}
		if uses.Dk > 1 || uses.Dk < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты внешние выходы (%v): № ДК должен быть 0 или 1", uses.Name))
			result.Err = errors.New("detected")

		}
		if uses.Dk == 0 && (uses.Type != 0 || uses.Tvps != 0) {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верно указано значение карты внешние выходы (%v): № ДК должен быть 1 если заполнены предыдушие поля", uses.Name))
			result.Err = errors.New("detected")
		}

	}
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Внешние выходы: OK")
	}
	return
}
