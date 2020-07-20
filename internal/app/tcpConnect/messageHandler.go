package tcpConnect

//SendRespTCPMess канал приема сообщений от сервера
var SendRespTCPMess chan TCPMessage

//SendToUserResp канал отправки сообщения пользователям
//var SendToUserResp chan RespTCPMess

var (
	TypeDispatch       = "dispatch" //команды арм
	TypeState          = "state"    //команды стате
	TypeChangeProtocol = "changeProtocol"
	typeInfo           map[string]string //ключ тип сервера, значение ip сервера
)

//RespTCPMess ответ сервера со статусом
//type RespTCPMess struct {
//	User   string      `json:"user"` //пользователь который отправил сообщение
//	From   string      `json:"-"`    //от какого соекета пришел запрос
//	Id     int         `json:"id"`   //id устройства
//	Data   interface{} `json:"-"`    //отправленные данные
//	Type   string      `json:"-"`    //тип команды отправки
//	Status bool        `json:"-"`    //статус выполнения команды
//}

var CrossSocGetTCPResp chan TCPMessage
var CrControlSocGetTCPResp chan TCPMessage
var GSGetTCPResp chan TCPMessage
var MapGetTCPResp chan TCPMessage
var TArmGetTCPResp chan TCPMessage
var (
	GsSoc        = "gsSoc"
	CrossSoc     = "crossSoc"
	CrControlSoc = "crControlSoc"
	MapSoc       = "mapSoc"
	TechArmSoc   = "techArmSoc"
)

//tcpRespBroadcast рассылка сообщении
func tcpRespBroadcast() {
	for {
		select {
		case msg := <-SendRespTCPMess:
			{
				switch msg.From {
				case CrossSoc:
					{
						CrossSocGetTCPResp <- msg
					}
				case CrControlSoc:
					{
						CrControlSocGetTCPResp <- msg
					}
				case GsSoc:
					{
						GSGetTCPResp <- msg
					}
				case MapSoc:
					{
						MapGetTCPResp <- msg
					}
				case TechArmSoc:
					{
						TArmGetTCPResp <- msg
					}
				}
			}
		}
	}
}
