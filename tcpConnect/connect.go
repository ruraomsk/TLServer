package tcpConnect

import (
	"../logger"
	"net"
	"os"
	"time"
)

type StateMessage struct {
	User     string
	StateStr string
	Message  string
}

var StateChan chan StateMessage = make(chan StateMessage)

func TCPClientStart() {
	var (
		conn net.Conn
		err  error
	)
	for {
		conn, err = net.Dial("tcp", os.Getenv("tcpServerAddress"))
		if err != nil {
			logger.Error.Println("|Message: TCP Server not responding: " + err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		for {
			state := <-StateChan
			state.StateStr += "\n"
			_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			_, err := conn.Write([]byte(state.StateStr))
			if err != nil {
				logger.Error.Println("|Message: TCP Server not responding: " + err.Error())
				state.Message = err.Error()
				StateChan <- state
				_ = conn.Close()
				break
			}
			state.Message = "ok"
			StateChan <- state
		}
	}
}
