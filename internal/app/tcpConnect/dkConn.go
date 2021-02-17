package tcpConnect

import (
	"bufio"
	"encoding/json"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"net"
	"time"
)

var ChanChangeDk chan DkInfo

type DkInfo struct {
	Idevice int      `json:"idevice"`
	DK      pudge.DK `json:"dk"`
}

func DkListen(tcpConfig TCPConfig) {
	ChanChangeDk = make(chan DkInfo, 50)
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
					logger.Info.Println("|Message: Dk read err " + err.Error())
				}
				errCount++
				time.Sleep(time.Second * 5)
				_ = conn.Close()
				break
			}
			var dk DkInfo
			err = json.Unmarshal(b, &dk)
			if err == nil {
				ChanChangeDk <- dk
			}
			errCount = 0
		}
	}
}
