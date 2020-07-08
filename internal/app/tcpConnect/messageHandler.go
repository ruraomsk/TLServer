package tcpConnect

import "time"

//SendRespTCPMess канал приема сообщений от сервера
var SendRespTCPMess chan RespTCPMess

//SendToUserResp канал отправки сообщения пользователям
var SendToUserResp chan RespTCPMess

var (
	TypeDispatch = "dispatch"      //команды арм
	TypeState    = "state"         //команды стате
	typeInfo     map[string]string //ключ тип сервера, значение ip сервера
)

//RespTCPMess ответ сервера со статусом
type RespTCPMess struct {
	User   string `json:"user"` //пользователь который отправил сообщение
	Id     int    `json:"id"`   //id устройства
	Type   string `json:"-"`    //тип команды отправки
	Status bool   `json:"-"`    //статус выполнения команды
}

//tcpRespBroadcast рассылка сообщении
func tcpRespBroadcast() {
	for {
		select {
		case msg := <-SendRespTCPMess:
			{
				SendToUserResp <- msg
				time.Sleep(time.Millisecond * 10)
			}
		}
	}
}
