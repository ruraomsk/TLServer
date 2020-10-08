package crossSock

import (
	"github.com/JanFant/TLServer/internal/model/accToken"
	u "github.com/JanFant/TLServer/internal/utils"
	"net/http"
)

//CrossDisc информация о занятых на редактирования страницах
type CrossDisc struct {
	Arms    []CrossInfo `json:"arms"`    //список занятых АРМов управления
	Crosses []CrossInfo `json:"crosses"` //список занятых АРМов редактирования привязки
}

//DisplayCrossEditInfo сбор информации для отображения информации о редактируемых страницах
func DisplayCrossEditInfo(accInfo *accToken.Token) u.Response {
	resp := u.Message(http.StatusOK, "edit info")

	GetArmUsersForDisplay <- true
	arms := <-CrArmUsersForDisplay
	if len(arms) == 0 {
		arms = make([]CrossInfo, 0)
	}

	GetCrossUsersForDisplay <- true
	crosses := <-CrossUsersForDisplay
	if len(crosses) == 0 {
		crosses = make([]CrossInfo, 0)
	}

	if accInfo.Region != "*" {
		var temp = make([]CrossInfo, 0)
		for _, arm := range arms {
			if arm.Pos.Region == accInfo.Region {
				temp = append(temp, arm)
			}
		}
		arms = temp

		temp = make([]CrossInfo, 0)
		for _, cross := range crosses {
			if cross.Pos.Region == accInfo.Region {
				temp = append(temp, cross)
			}
		}
		crosses = temp
	}

	resp.Obj["arms"] = arms
	resp.Obj["crosses"] = crosses
	return resp
}

//CrossEditFree сброс редактирования занях армов
func CrossEditFree(disc CrossDisc) u.Response {
	resp := u.Message(http.StatusOK, "free")
	if len(disc.Arms) > 0 {
		DiscArmUsers <- disc.Arms
	}
	if len(disc.Crosses) > 0 {
		DiscCrossUsers <- disc.Crosses
	}
	return resp
}
