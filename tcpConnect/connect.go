package tcpConnect

import (
	"net"
	"os"
	"time"

	"../logger"
)

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

var StateChan chan StateMessage = make(chan StateMessage)
var ArmCommandChan chan ArmCommandMessage = make(chan ArmCommandMessage)

func TCPClientStart() {
	go TCPForState(os.Getenv("tcpServerAddress") + os.Getenv("portState"))
	go TCPForARM(os.Getenv("tcpServerAddress") + os.Getenv("portArmCommand"))
}

func TCPForState(IP string) {
	var (
		conn net.Conn
		err  error
	)
	for {
		conn, err = net.Dial("tcp", IP)
		if err != nil {
			logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		for {
			state := <-StateChan
			state.StateStr += "\n"
			_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			_, err := conn.Write([]byte(state.StateStr))
			if err != nil {
				logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
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

func TCPForARM(IP string) {
	var (
		conn net.Conn
		err  error
	)
	for {
		conn, err = net.Dial("tcp", IP)

		if err != nil {
			logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
			time.Sleep(time.Second * 5)
			continue
		}
		for {
			armCommand := <-ArmCommandChan
			armCommand.CommandStr += "\n"
			_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
			_, err := conn.Write([]byte(armCommand.CommandStr))
			if err != nil {
				logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
				armCommand.Message = err.Error()
				ArmCommandChan <- armCommand
				_ = conn.Close()
				break
			}
			armCommand.Message = "ok"
			ArmCommandChan <- armCommand
		}
	}
}
