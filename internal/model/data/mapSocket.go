package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var connectedUsersOnMap map[*websocket.Conn]bool
var writeMap chan MapSokResponse
var mapRepaint chan bool

//MapReader обработчик открытия сокета для карты
func MapReader(conn *websocket.Conn, c *gin.Context) {
	connectedUsersOnMap[conn] = true
	login := ""
	flag, mapContx := checkToken(c)

	{
		resp := newMapMess(typeMapInfo, conn, mapOpenInfo())
		if flag {
			login = mapContx["login"]
			resp.Data["manageFlag"], _ = AccessCheck(login, mapContx["role"], 1)
			resp.Data["logDeviceFlag"], _ = AccessCheck(login, mapContx["role"], 11)
			resp.Data["authorizedFlag"] = true
		}
		resp.send()
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//закрытие коннекта
			resp := newMapMess(typeClose, conn, nil)
			resp.send()
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := newMapMess(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeJump:
			{
				location := &Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := newMapMess(typeJump, conn, nil)
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

//MapBroadcast передатчик для карты (map)
func MapBroadcast() {
	connectedUsersOnMap = make(map[*websocket.Conn]bool)
	writeMap = make(chan MapSokResponse)
	mapRepaint = make(chan bool)

	crossReadTick := time.Tick(time.Second * 5)

	oldTFs := selectTL()
	for {
		select {
		case <-crossReadTick:
			{
				if len(connectedUsersOnMap) > 0 {
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
					if len(tempTF) > 0 {
						resp := newMapMess(typeTFlight, nil, nil)
						resp.Data["tflight"] = tempTF
						for conn := range connectedUsersOnMap {
							_ = conn.WriteJSON(resp)
						}
					}
				}
			}
		case <-mapRepaint:
			{
				time.Sleep(time.Second * time.Duration(config.GlobalConfig.DBWait))
				oldTFs = selectTL()
				resp := newMapMess(typeRepaint, nil, nil)
				resp.Data["tflight"] = oldTFs
				for conn := range connectedUsersOnMap {
					_ = conn.WriteJSON(resp)
				}
			}
		case msg := <-writeMap:
			switch msg.Type {
			case typeClose:
				{
					delete(connectedUsersOnMap, msg.conn)
				}
			default:
				{
					if err := msg.conn.WriteJSON(msg); err != nil {
						delete(connectedUsersOnMap, msg.conn)
						_ = msg.conn.Close()
					}
				}
			}
		}
	}
}
