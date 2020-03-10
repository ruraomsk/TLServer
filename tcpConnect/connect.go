package tcpConnect

import (
	"net"
	"time"

	"github.com/JanFant/TLServer/logger"
)

//StateMessage state информация для отправки на сервер
type StateMessage struct {
	User     string //пользователь отправляющий данные (логин)
	Info     string //короткая информация о state
	StateStr string //данные подготовленные к отправке
	Message  string //информация о результате передачи данных
}

//ArmCommandMessage ARM информация для отправки на сервер
type ArmCommandMessage struct {
	User       string //пользователь отправляющий данные (логин)
	CommandStr string //данные подготовленные к отправке
	Message    string //информация о результате передачи данных
}

//TCPConfig настройки для тсп соединения
type TCPConfig struct {
	ServerAddr  string `toml:"tcpServerAddress"`
	PortState   string `toml:"portState"`
	PortArmComm string `toml:"portArmCommand"`
}

//getStateIP возвращает ip+port для State соединения
func (tcpConfig *TCPConfig) getStateIP() string {
	return tcpConfig.ServerAddr + tcpConfig.PortState
}

//getArmIP возвращает ip+port для ArmCommand соединения
func (tcpConfig *TCPConfig) getArmIP() string {
	return tcpConfig.ServerAddr + tcpConfig.PortArmComm
}

//StateChan канал для передачи информации связанных со state
var StateChan = make(chan StateMessage)

//ArmCommandChan канал для передачи информации связанных с командами арма
var ArmCommandChan = make(chan ArmCommandMessage)

//TCPClientStart запуск соединений
func TCPClientStart(tcpConfig TCPConfig) {
	go TCPForState(tcpConfig.getStateIP())
	go TCPForARM(tcpConfig.getArmIP())
}

//TCPForState для обмена с сервером данные State
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
