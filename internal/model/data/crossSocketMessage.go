package data

import (
	"github.com/JanFant/TLServer/logger"
)

type CrossSokResponse struct {
	Type   string                 `json:"type"`
	Data   map[string]interface{} `json:"data"`
	crConn CrossConn              `json:"-"`
}

func crossSokMessage(mType string, crConn CrossConn, data map[string]interface{}) CrossSokResponse {
	var resp CrossSokResponse
	resp.Type = mType
	resp.crConn = crConn
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

func (m *CrossSokResponse) send() {
	if m.Type == typeError {
		go func() {
			logger.Warning.Printf("|IP: %s |Login: %s |Resource: %s |Message: %v", m.crConn.Conn.RemoteAddr(), "map socket", "/map", m.Data["message"])
		}()
	}
	WriteCrossMessage <- *m
}

var (
	typeCrossBuild         = "cross build"
	errDoubleOpeningDevice = "double opening device"
)
