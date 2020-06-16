package data

import (
	u "github.com/JanFant/TLServer/internal/utils"
	"net/http"
)

func DisplayCrossEditInfo(mapContx map[string]string) u.Response {
	resp := u.Message(http.StatusOK, "Edit info")

	getArmUsers <- true
	arms := <-crArmUsers
	if len(arms) == 0 {
		arms = make([]crossInfo, 0)
	}

	getCrossUsers <- true
	crosses := <-crossUsers
	if len(crosses) == 0 {
		crosses = make([]crossInfo, 0)
	}

	if mapContx["region"] != "*" {
		var temp []crossInfo
		for _, arm := range arms {
			if arm.Pos.Region == mapContx["region"] {
				temp = append(temp, arm)
			}
		}
		arms = temp

		temp = make([]crossInfo, 0)
		for _, cross := range crosses {
			if cross.Pos.Region == mapContx["region"] {
				temp = append(temp, cross)
			}
		}
		crosses = temp
	}

	resp.Obj["arms"] = arms
	resp.Obj["crosses"] = crosses
	return resp
}
