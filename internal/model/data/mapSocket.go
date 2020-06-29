package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/chat"
	"time"

	"github.com/JanFant/TLServer/internal/model/config"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var connectedUsersOnMap map[*websocket.Conn]bool
var writeMap chan MapSokResponse
var mapRepaint chan bool
var getCrossUserForMap chan bool

const pingPeriod = time.Second * 30

//MapReader обработчик открытия сокета для карты
func MapReader(conn *websocket.Conn, c *gin.Context) {
	connectedUsersOnMap[conn] = true
	login := ""
	flag, mapContx := checkToken(c)
	{
		resp := newMapMess(typeMapInfo, conn, MapOpenInfo())
		if flag {
			login = fmt.Sprint(mapContx["login"])
			role := fmt.Sprint(mapContx["role"])
			resp.Data["manageFlag"], _ = AccessCheck(login, role, 2)
			resp.Data["logDeviceFlag"], _ = AccessCheck(login, role, 5)
			resp.Data["techArmFlag"], _ = AccessCheck(login, role, 7)
			resp.Data["description"] = mapContx["description"]
			resp.Data["authorizedFlag"] = true
			resp.Data["region"] = mapContx["region"]
			resp.Data["area"] = mapContx["area"]
			CacheArea.Mux.Lock()
			resp.Data["areaBox"] = CacheArea.Areas
			CacheArea.Mux.Unlock()
		}
		resp.send()
	}
	getCrossUserForMap <- true
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
				resp := newMapMess(typeLogin, conn, nil)
				resp = Login(account.Login, account.Password, conn.RemoteAddr().String())
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
					chat.UserLogoutChat <- login
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
	getCrossUserForMap = make(chan bool)

	crossReadTick := time.NewTicker(time.Second * 5)
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		crossReadTick.Stop()
	}()
	oldTFs := SelectTL()
	for {
		select {
		case <-crossReadTick.C:
			{
				if len(connectedUsersOnMap) > 0 {
					newTFs := SelectTL()
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
				time.Sleep(time.Second * time.Duration(config.GlobalConfig.DBConfig.DBWait))
				oldTFs = SelectTL()
				resp := newMapMess(typeRepaint, nil, nil)
				resp.Data["tflight"] = oldTFs
				FillMapAreaBox()
				CacheArea.Mux.Lock()
				resp.Data["areaBox"] = CacheArea.Areas
				CacheArea.Mux.Unlock()
				for conn := range connectedUsersOnMap {
					_ = conn.WriteJSON(resp)
				}

			}
		case <-pingTicker.C:
			{
				for conn := range connectedUsersOnMap {
					_ = conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		case crossUsers := <-crossUsersForMap:
			{
				resp := newMapMess(typeEditCrossUsers, nil, nil)
				resp.Data["editCrossUsers"] = crossUsers
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
					_ = msg.conn.WriteJSON(msg)
				}
			}
		}
	}
}
