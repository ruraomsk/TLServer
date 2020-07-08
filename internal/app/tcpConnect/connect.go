package tcpConnect

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"net"
	"time"
)

//poolTCPConnect хранилише подключений (ключ ip, значение информация о соединении)
var poolTCPConnect map[string]tcpInfo

//SendMessageToTCPServer канал для приема сообщений для отправки на сервер
var SendMessageToTCPServer chan TCPMessage

//tcpInfo информация о тсп соединения
type tcpInfo struct {
	conn        net.Conn //соединение
	flagConnect bool     //статус соединения
	errCount    int      //счетчик ошибок чтобы не заспамить лог файл
}

//TCPMessage структура данных для обработки и отправки ТСП сообщений
type TCPMessage struct {
	Type string      //тип сообщения
	User string      //пользователь который отправил сообщение
	Id   int         //id устройства на которое отправляется сообщение
	Data interface{} //данные для отправки
}

//SendToTCPServer отправка сообщения на сервер
func (m *TCPMessage) SendToTCPServer() bool {
	SendMessageToTCPServer <- *m
	for {
		tcpResp := <-SendToUserResp
		if tcpResp.User == m.User && tcpResp.Id == m.Id && tcpResp.Type == m.Type {
			return tcpResp.Status
		} else {
			SendRespTCPMess <- tcpResp
		}
	}
}

//dataToString превратить данные в строку и добавить \n для понимания сервера
func (m *TCPMessage) dataToString() string {
	raw, _ := json.Marshal(m.Data)
	return fmt.Sprint(string(raw), "\n")
}

//TCPClientStart запуск соединений
func TCPClientStart(tcpConfig TCPConfig) {
	typeInfo = make(map[string]string)
	typeInfo[TypeDispatch] = tcpConfig.getArmIP()
	typeInfo[TypeState] = tcpConfig.getStateIP()
	go TCPBroadcast(typeInfo)
}

//TCPBroadcast обработка подключений к сервера ТСП и отправка сообщений на него
func TCPBroadcast(typeIP map[string]string) {
	poolTCPConnect = make(map[string]tcpInfo)
	SendMessageToTCPServer = make(chan TCPMessage, 5)
	SendRespTCPMess = make(chan RespTCPMess, 20)
	SendToUserResp = make(chan RespTCPMess, 20)
	go tcpRespBroadcast()

	for _, ip := range typeIP {
		poolTCPConnect[ip] = tcpInfo{conn: nil, errCount: 0, flagConnect: false}
	}
	timeTick := time.NewTicker(time.Second * 5)
	defer timeTick.Stop()

	for {
		select {
		case <-timeTick.C:
			{
				//крутим все подключения смотрим че да как
				for ip, connInfo := range poolTCPConnect {
					if !connInfo.flagConnect {
						conn, err := net.Dial("tcp", ip)
						if err != nil {
							//проверка нужно ли еще писать в лог инфу о неподключении
							if connInfo.errCount < 5 {
								logger.Error.Println("|Message: TCP Server " + ip + " not responding: " + err.Error())
							}
							connInfo.errCount++
							//подключиться не удалось запишим коунт и пойдем дальше
							poolTCPConnect[ip] = connInfo
							time.Sleep(time.Second * 5)
							continue
						}
						//соединение создалось, сохраняем информацию
						connInfo.flagConnect = true
						connInfo.conn = conn
						poolTCPConnect[ip] = connInfo
					}
					//проверка наличия соединения
					_ = connInfo.conn.SetWriteDeadline(time.Now().Add(time.Second))
					_, err := connInfo.conn.Write([]byte("0\n"))
					if err != nil {
						connInfo.flagConnect = false
					}
				}
			}
		case msg := <-SendMessageToTCPServer:
			{
				info, _ := poolTCPConnect[typeIP[msg.Type]]
				var resp = RespTCPMess{Type: msg.Type, User: msg.User, Status: false, Id: msg.Id}
				if !info.flagConnect {
					//соединение почемуто не открыто, пусть попробует в следующий раз
					resp.Status = false
					SendRespTCPMess <- resp
					continue
				}
				//ошибки нет подготовим данные и отправим
				sendStr := msg.dataToString()

				_ = info.conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
				_, err := info.conn.Write([]byte(sendStr))
				if err != nil {
					//почемуто не отправили сообщение
					if info.errCount < 5 {
						logger.Error.Println("|Message: TCP Server " + msg.Type + " not responding: " + err.Error())
						info.flagConnect = false
					}
					info.errCount++
					_ = info.conn.Close()
					//соединение оборвано отключились и нужно отправить
					poolTCPConnect[typeIP[msg.Type]] = info
					resp.Status = false
					SendRespTCPMess <- resp
					continue
				}
				//все нормально отправил ответ об успешности отправки
				info.errCount = 0
				resp.Status = true
				poolTCPConnect[typeIP[msg.Type]] = info
				SendRespTCPMess <- resp
			}
		}
	}
}
