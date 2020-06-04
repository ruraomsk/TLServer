package data

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var ConnectedUsers map[*websocket.Conn]bool
var WriteMap chan mapResponse

func MapReader(conn *websocket.Conn) {
	ConnectedUsers[conn] = true
	login := ""
	{
		resp := mapMessage(typeMapInfo, conn)
		resp.Data = mapOpenInfo()
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
			resp := mapMessage(typeError, conn)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeJump:
			{
				location := &Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := mapMessage(typeJump, conn)
				resp.Data["boxPoint"] = box
				resp.send()
			}
		case typeLogin:
			{
				account := &Account{}
				_ = json.Unmarshal(p, &account)
				resp := mapMessage(typeLogin, conn)
				obj := Login(account.Login, account.Password, conn.RemoteAddr().String())
				if obj.Code == http.StatusOK {
					login = fmt.Sprint(obj.Obj["login"])
					resp.Data["authorizedFlag"] = true
					resp.Data["manageFlag"], _ = AccessCheck(login, fmt.Sprint(obj.Obj["role"]), 1)
					resp.Data["logDeviceFlag"], _ = AccessCheck(login, fmt.Sprint(obj.Obj["role"]), 11)
				}
				resp.Data["login"] = obj
				resp.send()
			}
		case typeLogOut:
			{
				if login != "" {
					resp := mapMessage(typeLogOut, conn)
					resp.Data["logOut"] = LogOut(login)
					resp.send()
				}
			}

		}
	}
}

func Broadcast() {
	ConnectedUsers = make(map[*websocket.Conn]bool)
	WriteMap = make(chan mapResponse)
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
						resp := mapMessage(typeTFlight, nil)
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
