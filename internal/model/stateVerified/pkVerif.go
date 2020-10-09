package stateVerified

import (
	"errors"
	"fmt"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

// PkVerified проверка ПК
func PkVerified(cross *agspudge.Cross) (result StateResult) {
	DKs := cross.Arrays.SetDK.DK
	result.SumResult = append(result.SumResult, "Проверка: ПК")
	var pkNums = make(map[int]bool)
	for _, dk := range DKs {
		if _, ok := pkNums[dk.Pk]; !ok {
			pkNums[dk.Pk] = true
		} else {
			result.SumResult = append(result.SumResult, fmt.Sprintf("Нарушение нумерации ПК"))
			result.Err = errors.New("detected")
		}

	}
	if result.Err == nil {
		result.SumResult = append(result.SumResult, "ПК: OK")
	}
	return
}
