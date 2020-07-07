package tcpConnect

import "time"

var SendRespTCPMess chan RespTCPMess
var SendToUserResp chan RespTCPMess

var (
	TypeDispatch = "dispatch"
	TypeState    = "state"
	typeInfo     map[string]string
)

type RespTCPMess struct {
	User    string `json:"user"`
	Id      int    `json:"id"`
	Type    string `json:"-"`
	Status  bool   `json:"-"`
	Message string `json:"message"`
}

//tcpRespBroadcast рассылаю сообщения
func tcpRespBroadcast() {
	for {
		select {
		case msg := <-SendRespTCPMess:
			{
				SendToUserResp <- msg
				time.Sleep(time.Millisecond * 100)
			}
		}
	}
}
