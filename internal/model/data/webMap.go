package data

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

var ConnectedUsers map[*websocket.Conn]bool
var WriteMap chan MapSocketResponse

func MapReader(conn *websocket.Conn) {
	ConnectedUsers[conn] = true
	login := ""
	{
		resp := mapMessage(typeMapInfo, conn, mapOpenInfo())
		resp.send()
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			delete(ConnectedUsers, conn)
			//закрытие коннекта
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := mapMessage(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeJump:
			{
				location := &Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := mapMessage(typeJump, conn, nil)
				resp.Data["boxPoint"] = box
				resp.send()
			}
		case typeLogin:
			{
				account := &Account{}
				_ = json.Unmarshal(p, &account)
				resp := Login(account.Login, account.Password, conn.RemoteAddr().String())
				if resp.Type == typeLogin {
					login = fmt.Sprint(resp.Data["login"])
				}
				resp.conn = conn
				resp.send()
			}
		case typeLogOut:
			{
				if login != "" {
					resp := LogOut(login)
					resp.conn = conn
					resp.Data["authorizedFlag"] = true
					resp.send()
				}
			}

		}
	}
}

func Broadcast() {
	ConnectedUsers = make(map[*websocket.Conn]bool)
	WriteMap = make(chan MapSocketResponse)
	crossReadTick := time.Tick(time.Second * 5)
	oldTFs := selectTL()
	for {
		select {
		case <-crossReadTick:
			{
				newTFs := selectTL()
				var tempTF []TrafficLights
				for _, nTF := range newTFs {
					for _, oTF := range oldTFs {
						if oTF.Idevice == nTF.Idevice && oTF.Sost.Num != nTF.Sost.Num {
							tempTF = append(tempTF, nTF)
							break
						}
					}
				}
				oldTFs = newTFs
				if len(ConnectedUsers) > 0 {
					if len(tempTF) > 0 {
						resp := mapMessage(typeTFlight, nil, nil)
						resp.Data["tflight"] = tempTF
						for conn := range ConnectedUsers {
							if err := conn.WriteJSON(resp); err != nil {
								_ = conn.Close()
							}
						}
					}
				}
			}
		case msg := <-WriteMap:
			{
				if err := msg.conn.WriteJSON(msg); err != nil {
					_ = msg.conn.Close()
					return
				}
			}
		}
	}
}

//var (
//	typePage  = "page"
//	typePage1 = "page1"
//	typePage2 = "page2"
//)
//
//case typePage1:
//{
//resp := mapMessage(typePage, conn, nil)
//raw, _ := ioutil.ReadFile("./web/html/screen111.html")
//resp.Data["html"] = string(raw)
//resp.send()
//}
//case typePage2:
//{
//resp := mapMessage(typePage, conn, nil)
//raw, _ := ioutil.ReadFile("./web/html/screen.html")
//resp.Data["html"] = string(raw)
//resp.send()
//}
