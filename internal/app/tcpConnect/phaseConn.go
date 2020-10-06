package tcpConnect

import (
	"bufio"
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/phaseInfo"
	"github.com/JanFant/TLServer/logger"
	"net"
	"time"
)

var ChanChangePhase chan phaseInfo.Phase

func PhaseListen(tcpConfig TCPConfig) {
	ChanChangePhase = make(chan phaseInfo.Phase, 50)
	errCount := 0
	for {
		conn, err := net.Dial("tcp", tcpConfig.getPhaseIP())
		if err != nil {
			if errCount < 5 {
				logger.Error.Println("|Message: TCP Server " + tcpConfig.getPhaseIP() + " not responding: " + err.Error())
			}
			errCount++
			time.Sleep(time.Second * 5)
			continue
		}
		reader := bufio.NewReader(conn)
		for {
			_ = conn.SetReadDeadline(time.Now().Add(time.Second * 20))
			b, err := reader.ReadBytes('\n')
			if err != nil {
				if errCount < 5 {
					logger.Info.Println("|Message: Phase read err " + err.Error())
				}
				errCount++
				time.Sleep(time.Second * 5)
				_ = conn.Close()
				break
			}
			var phase phaseInfo.Phase
			err = json.Unmarshal(b, &phase)
			if err == nil {
				ChanChangePhase <- phase
			}
			errCount = 0
		}
	}
}
