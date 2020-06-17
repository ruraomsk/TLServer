package data

import (
	u "github.com/JanFant/TLServer/internal/utils"
	"net/http"
)

type CrossDisc struct {
	Arms    []CrossInfo `json:"arms"`
	Crosses []CrossInfo `json:"crosses"`
}

func DisplayCrossEditInfo(mapContx map[string]string) u.Response {
	resp := u.Message(http.StatusOK, "edit info")

	getArmUsers <- true
	arms := <-crArmUsers
	if len(arms) == 0 {
		arms = make([]CrossInfo, 0)
	}

	getCrossUsers <- true
	crosses := <-crossUsers
	if len(crosses) == 0 {
		crosses = make([]CrossInfo, 0)
	}

	if mapContx["region"] != "*" {
		var temp = make([]CrossInfo, 0)
		for _, arm := range arms {
			if arm.Pos.Region == mapContx["region"] {
				temp = append(temp, arm)
			}
		}
		arms = temp

		temp = make([]CrossInfo, 0)
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

func CrossEditFree(disc CrossDisc) u.Response {
	resp := u.Message(http.StatusOK, "free")
	if len(disc.Arms) > 0 {
		discArmUsers <- disc.Arms
	}
	if len(disc.Crosses) > 0 {
		discCrossUsers <- disc.Crosses
	}
	return resp
}
