package tcpConnect

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/sockets"
)

//SendRespTCPMess канал приема сообщений от сервера
var SendRespTCPMess chan TCPMessage

//SendMessageToTCPServer канал для приема сообщений для отправки на сервер
var SendMessageToTCPServer chan TCPMessage

var (
	TCPRespCrossSoc     chan TCPMessage //канал для отправки ответа в сокет Cross
	TCPRespCrControlSoc chan TCPMessage //канал для отправки ответа в сокет CrossControl
	TCPRespGS           chan TCPMessage //канал для отправки ответа в сокет GreenStreet
	TCPRespMap          chan TCPMessage //канал для отправки ответа в сокет Map
	TCPRespTArm         chan TCPMessage //канал для отправки ответа в сокет TechArm

	TypeDispatch       = "dispatch"       //команды арм
	TypeState          = "state"          //команды стате
	TypeChangeProtocol = "changeProtocol" //команды gps protocol
	typeInfo           map[string]string  //ключ тип сервера, значение ip сервера
)

var (
	FromGsSoc        = "gsSoc"        //обозначение сокета GreenStreet
	FromCrossSoc     = "crossSoc"     //обозначение сокета Cross
	FromCrControlSoc = "crControlSoc" //обозначение сокета CrossControl
	FromMapSoc       = "mapSoc"       //обозначение сокета Map
	FromTechArmSoc   = "techArmSoc"   //обозначение сокета TechArm
	FromServer       = "server"       //обозначение самого сервера
)

//TCPMessage структура данных для обработки и отправки ТСП сообщений
type TCPMessage struct {
	TCPType string      //тип тсп сообщения указывается из messageHandle
	User    string      //пользователь который отправил сообщение
	From    string      //указания в какой сокет вернуть сообщение
	Data    interface{} //данные для отправки

	CommandType string          //тип команды указывается на "месте"
	Pos         sockets.PosInfo //информация о перекрестке
	Idevice     int             //id устройства на которое отправляется сообщение
	Status      bool            //статус выполнения команды
}

//SendToTCPServer отправка сообщения на сервер
func (m *TCPMessage) SendToTCPServer() {
	SendMessageToTCPServer <- *m
}

//dataToString превратить данные в строку и добавить \n для понимания сервера
func (m *TCPMessage) dataToString() string {
	raw, _ := json.Marshal(m.Data)
	return fmt.Sprint(string(raw), "\n")
}

//tcpRespBroadcast рассылка сообщении
func tcpRespBroadcast() {
	for {
		select {
		case msg := <-SendRespTCPMess:
			{
				switch msg.From {
				case FromCrossSoc:
					{
						TCPRespCrossSoc <- msg
					}
				case FromCrControlSoc:
					{
						TCPRespCrControlSoc <- msg
					}
				case FromGsSoc:
					{
						TCPRespGS <- msg
					}
				case FromMapSoc:
					{
						TCPRespMap <- msg
					}
				case FromTechArmSoc:
					{
						TCPRespTArm <- msg
					}
				case FromServer:
					{

					}
				default:
					{

					}
				}
			}
		}
	}
}
