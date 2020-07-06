package stateVerified

import (
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"strings"
)

// CtrlVerified проверка контроль входов
func CtrlVerified(cross *agspudge.Cross) (result StateResult) {
	Ctrl := cross.Arrays.SetCtrl
	result.SumResult = append(result.SumResult, "Проверка: Контроль входов")
	for _, stage := range Ctrl.Stage {
		validRes := validation.ValidateStruct(&stage.End,
			validation.Field(&stage.End.Hour, validation.Min(0), validation.Max(24)),
			validation.Field(&stage.End.Minute, validation.Min(0), validation.Max(59)),
		)
		if validRes != nil {
			if strings.Contains(validRes.Error(), "hour") {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): час должен быть от 0 до 24", stage.Nline))
			}
			if strings.Contains(validRes.Error(), "min") {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): минуты должны быть от 0 до 59", stage.Nline))
			}
			result.Err = errors.New("detected")
		}
		validRes = validation.ValidateStruct(&stage,
			validation.Field(&stage.MGRLen, validation.Min(0), validation.Max(255)),
			validation.Field(&stage.TVPLen, validation.Min(0), validation.Max(255)),
		)
		if validRes != nil {
			if strings.Contains(validRes.Error(), "lenMGR") {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Т конт МГР должно быть от 0 до 255", stage.Nline))
			}
			if strings.Contains(validRes.Error(), "lenTVP") {
				result.SumResult = append(result.SumResult, fmt.Sprintf("Поле (%v): Т конт ТВП должно быть от 0 до 255", stage.Nline))
			}
			result.Err = errors.New("detected")
		}
	}
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Контроль входов: OK")
	}
	return
}
