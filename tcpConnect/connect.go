package tcpConnect

import (
	"net"
	"os"
	"time"

	"../logger"
)

//StateMessage State
type StateMessage struct {
	User     string
	Info     string
	StateStr string
	Message  string
}

//ArmCommandMessage ARM
type ArmCommandMessage struct {
	User       string
	CommandStr string
	Message    string
}

//StateChan канал для передачи информации связанных со state
var StateChan = make(chan StateMessage)

//ArmCommandChan канал для передачи информации связанных с командами арма
var ArmCommandChan = make(chan ArmCommandMessage)

//TCPClientStart запуск соединений
func TCPClientStart() {
	go TCPForState(os.Getenv("tcpServerAddress") + os.Getenv("portState"))
	go TCPForARM(os.Getenv("tcpServerAddress") + os.Getenv("portArmCommand"))
}

//TCPForState для обмена с сервером State
func TCPForState(IP string) {
	var (
		conn     net.Conn
		err      error
		errCount = 0
	)
	timeTick := time.Tick(time.Second * 5)
	FlagConnect := false
	for {
		select {
		case state := <-StateChan:
			{
				if !FlagConnect {
					state.Message = "TCP Server not responding"
					StateChan <- state
					continue
				}
				state.StateStr += "\n"
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				_, err := conn.Write([]byte(state.StateStr))
				if err != nil {
					if errCount < 5 {
						logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
						FlagConnect = false
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
		case <-timeTick:
			{
				if !FlagConnect {
					conn, err = net.Dial("tcp", IP)
					if err != nil {
						if errCount < 5 {
							logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
						}
						errCount++
						time.Sleep(time.Second * 5)
						continue
					}
					FlagConnect = true
				}
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
				_, err := conn.Write([]byte("0\n"))
				if err != nil {
					FlagConnect = false
				}
			}
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
	timeTick := time.Tick(time.Second * 5)
	FlagConnect := false
	for {
		select {
		case armCommand := <-ArmCommandChan:
			{
				if !FlagConnect {
					armCommand.Message = "TCP Server not responding"
					ArmCommandChan <- armCommand
					continue
				}
				armCommand.CommandStr += "\n"
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				_, err := conn.Write([]byte(armCommand.CommandStr))
				if err != nil {
					if errCount < 5 {
						logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
						FlagConnect = false
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
		case <-timeTick:
			{
				if !FlagConnect {
					conn, err = net.Dial("tcp", IP)
					if err != nil {
						if errCount < 5 {
							logger.Error.Println("|Message: TCP Server " + IP + " not responding: " + err.Error())
						}
						errCount++
						time.Sleep(time.Second * 5)
						continue
					}
					FlagConnect = true
				}
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
				_, err := conn.Write([]byte("0\n"))
				if err != nil {
					FlagConnect = false
				}
			}
		}
	}
}
