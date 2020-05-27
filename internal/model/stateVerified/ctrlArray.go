package stateVerified

import (
	"errors"
	"fmt"
	agS_pudge "github.com/ruraomsk/ag-server/pudge"
)

// CtrlVerified проверка недельных карт
func CtrlVerified(cross *agS_pudge.Cross) (result StateResult) {
	Ctrl := cross.Arrays.SetCtrl
	result.SumResult = append(result.SumResult, "Проверка: Контроль входов")
	for _, stage := range Ctrl.Stage {
		if stage.End.Hour > 24 || stage.End.Hour < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): час должен быть от 0 до 24", stage.Nline))
			result.Err = errors.New("detected")

		}
		if stage.End.Minute > 59 || stage.End.Minute < 0 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): минуты должны быть от 0 до 59", stage.Nline))
			result.Err = errors.New("detected")
		}
		if stage.MGRLen < 0 || stage.MGRLen > 255 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Т конт МГР должно быть от 0 до 255", stage.Nline))
			result.Err = errors.New("detected")
		}
		if stage.TVPLen < 0 || stage.TVPLen > 255 {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Т конт ТВП должно быть от 0 до 255", stage.Nline))
			result.Err = errors.New("detected")
		}
	}
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Контроль входов: OK")
	}
	return
}
