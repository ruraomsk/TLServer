package stateVerified

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	agspudge "github.com/ruraomsk/ag-server/pudge"
)

// MainWindVerified проверка основной страницы
func MainWindVerified(cross *agspudge.Cross, db *sqlx.DB) (result StateResult) {
	model := cross.Model
	result.SumResult = append(result.SumResult, "Проверка: Основной вкладки")

	if model.VPBSL <= 0 || model.VPCPDL <= 0 || model.VPCPDR < 0 || model.VPBSR < 0 {
		result.SumResult = append(result.SumResult, fmt.Sprintf("Не верная версия ПО"))
		result.Err = errors.New("detected")
	}

	//row := db.QueryRow(`SELECT device->'Model' as model FROM public.devices WHERE id = $1`, cross.IDevice)
	//var (
	//	tempModel agspudge.Model
	//	strModel  []byte
	//)
	//
	//err := row.Scan(&strModel)
	//switch err {
	//case sql.ErrNoRows:
	//	{
	//		if model.VPBSL <= 0 || model.VPCPDL <= 0 || model.VPCPDR < 0 || model.VPBSR < 0 {
	//			result.SumResult = append(result.SumResult, fmt.Sprintf("Не верная версия ПО"))
	//			result.Err = errors.New("detected")
	//		}
	//	}
	//case nil:
	//	{
	//		err = json.Unmarshal(strModel, &tempModel)
	//		if err != nil {
	//			result.SumResult = append(result.SumResult, fmt.Sprintf("Ошибка разбора данных из БД, обратитесь к администратору"))
	//			result.Err = errors.New("detected")
	//		}
	//
	//		if tempModel.VPBSL != model.VPBSL || tempModel.VPBSR != model.VPBSR || tempModel.VPCPDL != model.VPCPDL || tempModel.VPCPDR != model.VPCPDR {
	//			result.SumResult = append(result.SumResult, fmt.Sprintf("Несовпадение версии ПО"))
	//			result.Err = errors.New("detected")
	//		}
	//	}
	//default:
	//	{
	//		result.SumResult = append(result.SumResult, fmt.Sprintf("Потеряна связь с БД"))
	//		result.Err = errors.New("detected")
	//	}
	//}

	if result.Err == nil {
		result.SumResult = append(result.SumResult, "Основнаяй вкладка: OK")
	}
	return
}
