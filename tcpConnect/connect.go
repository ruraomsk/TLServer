package tcpConnect

import (
	"net"
	"os"
	"time"

	"../logger"
)

//StateMessage
type StateMessage struct {
	User     string
	StateStr string
	Message  string
}

type ArmCommandMessage struct {
	User       string
	CommandStr string
	Message    string
}

var StateChan = make(chan StateMessage)
var ArmCommandChan = make(chan ArmCommandMessage)

func TCPClientStart() {
	go TCPForState(os.Getenv("tcpServerAddress") + os.Getenv("portState"))
	go TCPForARM(os.Getenv("tcpServerAddress") + os.Getenv("portArmCommand"))
}

//TCPForState соединение с сервром для обмена State
func TCPForState(IP string) {
	var (
		conn     net.Conn
		err      error
		errCount = 0
	)
	for {
		conn, err = net.Dial("tcp", IP)
		if err != nil {
			if errCount < 5 {
				logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
			}
			errCount++
			time.Sleep(time.Second * 5)
			continue
		}
		for {
			state := <-StateChan
			state.StateStr += "\n"
			_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			_, err := conn.Write([]byte(state.StateStr))
			if err != nil {
				if errCount < 5 {
					logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
				}
				errCount++
				state.Message = err.Error()
				StateChan <- state
				_ = conn.Close()
				break
			}
			state.Message = "ok"
			errCount = 0
			StateChan <- state
		}
	}
}

//TCPForARM для обмена с сервером команды АРМ
func TCPForARM(IP string) {
	var (
		conn     net.Conn
		err      error
		errCount = 0
	)
	for {
		conn, err = net.Dial("tcp", IP)

		if err != nil {
			if errCount < 5 {
				logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
			}
			errCount++
			time.Sleep(time.Second * 5)
			continue
		}
		for {
			armCommand := <-ArmCommandChan
			armCommand.CommandStr += "\n"
			_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			_, err := conn.Write([]byte(armCommand.CommandStr))
			if err != nil {
				if errCount < 5 {
					logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
				}
				errCount++
				armCommand.Message = err.Error()
				ArmCommandChan <- armCommand
				_ = conn.Close()
				break
			}
			armCommand.Message = "ok"
			errCount = 0
			ArmCommandChan <- armCommand
		}
	}
}
