package tcpConnect

import (
	"github.com/JanFant/TLServer/internal/model/device"
	"net"
	"time"

	"github.com/JanFant/TLServer/logger"
)

//poolTCPConnect хранилише подключений (ключ ip, значение информация о соединении)
var poolTCPConnect map[string]tcpInfo

//tcpInfo информация о тсп соединения
type tcpInfo struct {
	conn        net.Conn //соединение
	flagConnect bool     //статус соединения
	errCount    int      //счетчик ошибок чтобы не заспамить лог файл
}

//TCPClientStart запуск соединений
func TCPClientStart(tcpConfig TCPConfig) {
	typeInfo = make(map[string]string)
	typeInfo[TypeDispatch] = tcpConfig.getArmIP()
	typeInfo[TypeState] = tcpConfig.getStateIP()
	typeInfo[TypeChangeProtocol] = tcpConfig.getChangeProtocolIP()
	go DkListen(tcpConfig)
	go TCPBroadcast(typeInfo)
}

//TCPBroadcast обработка подключений к сервера ТСП и отправка сообщений на него
func TCPBroadcast(typeIP map[string]string) {
	startSys := true
	poolTCPConnect = make(map[string]tcpInfo)
	SendMessageToTCPServer = make(chan TCPMessage, 20)
	SendRespTCPMess = make(chan TCPMessage, 20)

	TCPRespCrossSoc = make(chan TCPMessage, 5)
	TCPRespCrControlSoc = make(chan TCPMessage, 5)
	TCPRespGS = make(chan TCPMessage, 5)
	TCPRespMap = make(chan TCPMessage, 5)
	TCPRespTArm = make(chan TCPMessage, 5)

	go tcpRespBroadcast()

	//заполняем пулл соединений по TCP
	for _, ip := range typeIP {
		poolTCPConnect[ip] = tcpInfo{conn: nil, errCount: 0, flagConnect: false}
	}
	//таймер времени опроса соединений
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
						poolTCPConnect[ip] = connInfo
					}
				}

				//отправим на страте всем устройствам подключенным к системе команду 4.0
				if startSys {
					devicesId := make([]int, 0)
					device.GlobalDevices.Mux.Lock()
					for key := range device.GlobalDevices.MapDevices {
						devicesId = append(devicesId, key)
					}
					device.GlobalDevices.Mux.Unlock()
					//arm := comm.CommandARM{User: "Server", Command: 4, Params: 0}
					//for _, devId := range devicesId {
					//	arm.ID = devId
					//	var mess = TCPMessage{
					//		TCPType:     TypeDispatch,
					//		User:        arm.User,
					//		From:        FromServer,
					//		Data:        arm,
					//		CommandType: "dispatch",
					//		Pos:         sockets.PosInfo{},
					//		Idevice:     arm.ID,
					//	}
					//	mess.SendToTCPServer()
					//}
					startSys = false
				}
			}
		case msg := <-SendMessageToTCPServer:
			{
				//определим кокое соединение нужно взять для отправки сообщения
				info, _ := poolTCPConnect[typeIP[msg.TCPType]]
				//сохраним экземпляр в ответное сообщение
				var resp = msg
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
						logger.Error.Println("|Message: TCP Server " + msg.TCPType + " not responding: " + err.Error())
						info.flagConnect = false
					}
					info.errCount++
					_ = info.conn.Close()
					//соединение оборвано отключились и нужно отправить
					poolTCPConnect[typeIP[msg.TCPType]] = info
					resp.Status = false
					SendRespTCPMess <- resp
					continue
				}
				//все нормально отправил ответ об успешности отправки
				info.errCount = 0
				resp.Status = true
				poolTCPConnect[typeIP[msg.TCPType]] = info
				SendRespTCPMess <- resp
			}
		}
	}
}
