package tcpConnect

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/license"
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
	ServerAddr      string `toml:"tcpServerAddress"` //адресс сервера
	PortState       string `toml:"portState"`        //порт для обмена Стате
	PortArmComm     string `toml:"portArmCommand"`   //порт для обмена арм командами
	PortMessageInfo string `toml:"portMessageInfo"`  //порт для запроса сохранения информации о состоянии сервера
}

//MessageInfo информация о устройствах (архив)
type MessageInfo struct {
	Id   int    //номер сервера
	Text string //информация о запросе (время)
}

//messageInfoMarshal преобразовать структуру в строку
func (mInfo *MessageInfo) messageInfoMarshal() (str string, err error) {
	newByte, err := json.Marshal(mInfo)
	if err != nil {
		return "", err
	}
	return string(newByte), err
}

//fillInfo Заполнить поле id из ключа
func (mInfo *MessageInfo) fillInfo() {
	license.LicenseFields.Mux.Lock()
	defer license.LicenseFields.Mux.Unlock()
	mInfo.Id = license.LicenseFields.Id
	mInfo.Text = time.Now().String()
}

//getStateIP возвращает ip+port для State соединения
func (tcpConfig *TCPConfig) getStateIP() string {
	return tcpConfig.ServerAddr + tcpConfig.PortState
}

//getArmIP возвращает ip+port для ArmCommand соединения
func (tcpConfig *TCPConfig) getArmIP() string {
	return tcpConfig.ServerAddr + tcpConfig.PortArmComm
}

//getMessageIP возвращает ip+port для Message соединения
func (tcpConfig *TCPConfig) getMessageIP() string {
	return tcpConfig.ServerAddr + tcpConfig.PortMessageInfo
}

//StateChan канал для передачи информации связанной со state
var StateChan = make(chan StateMessage)

//ArmCommandChan канал для передачи информации связанной с командами арма
var ArmCommandChan = make(chan ArmCommandMessage)

//TCPClientStart запуск соединений
func TCPClientStart(tcpConfig TCPConfig) {
	go TCPForState(tcpConfig.getStateIP())
	go TCPForARM(tcpConfig.getArmIP())
	go TCPForMessage(tcpConfig.getMessageIP())
}

//TCPForMessage обмен с сервером  о устройствах (архив)
func TCPForMessage(IP string) {
	var (
		conn    net.Conn
		message MessageInfo
		err     error
	)
	timeTick := time.Tick(time.Hour)
	FlagConnect := false
	for {
		select {
		case <-timeTick:
			{
				for {
					if !FlagConnect {
						conn, err = net.Dial("tcp", IP)
						if err != nil {
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
					break
				}
				message.fillInfo()
				messageStr, _ := message.messageInfoMarshal()
				messageStr += "\n"
				fmt.Println(messageStr)
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				_, err := conn.Write([]byte(messageStr))
				if err != nil {
					FlagConnect = false
					_ = conn.Close()
					break
				}
			}
		}
	}
}

//TCPForState обмен с сервером данными State
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

//TCPForARM обмен с сервером командами для АРМ
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
