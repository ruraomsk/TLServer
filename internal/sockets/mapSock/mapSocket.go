package mapSock

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/internal/sockets/chat"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/internal/sockets/techArm"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

var connectedUsersOnMap map[*websocket.Conn]bool //пулл соединений
var writeMap chan MapSokResponse                 //канал для отправки сообщений

const pingPeriod = time.Second * 30

//MapReader обработчик открытия сокета для карты
func MapReader(conn *websocket.Conn, c *gin.Context, db *sqlx.DB) {
	connectedUsersOnMap[conn] = true
	login := ""
	//подготовка начальной информации
	{
		flag, tk := checkToken(c, db)
		resp := newMapMess(typeMapInfo, conn, mapOpenInfo(db))
		if flag {
			login = tk.Login
			role := tk.Role
			resp.Data["manageFlag"], _ = data.AccessCheck(login, role, 2)
			resp.Data["logDeviceFlag"], _ = data.AccessCheck(login, role, 5)
			resp.Data["techArmFlag"], _ = data.AccessCheck(login, role, 7)
			resp.Data["description"] = tk.Description
			resp.Data["authorizedFlag"] = true
			resp.Data["region"] = tk.Region
			var areaMap = make(map[string]string)
			for _, area := range tk.Area {
				var tempA data.AreaInfo
				tempA.SetAreaInfo(tk.Region, area)
				areaMap[tempA.Num] = tempA.NameArea
			}
			resp.Data["area"] = areaMap
			data.CacheArea.Mux.Lock()
			resp.Data["areaZone"] = data.CacheArea.Areas
			data.CacheArea.Mux.Unlock()
		}
		resp.send()
	}
	crossSock.GetCrossUserForMap <- true
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//закрытие коннекта
			resp := newMapMess(typeClose, conn, nil)
			resp.send()
			return
		}

		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			resp := newMapMess(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeJump: //отправка default
			{
				location := &data.Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := newMapMess(typeJump, conn, nil)
				resp.Data["boxPoint"] = box
				resp.send()
			}
		case typeLogin: //отправка default
			{
				account := &data.Account{}
				_ = json.Unmarshal(p, &account)
				resp := newMapMess(typeLogin, conn, nil)
				resp.Data = logIn(account.Login, account.Password, conn.RemoteAddr().String(), db)
				if _, ok := resp.Data["message"]; !ok {
					login = fmt.Sprint(resp.Data["login"])
				}
				resp.send()
			}
		case typeChangeAccount:
			{
				account := &data.Account{}
				_ = json.Unmarshal(p, &account)
				resp := newMapMess(typeLogin, conn, nil)
				resp.Data = logIn(account.Login, account.Password, conn.RemoteAddr().String(), db)
				if _, ok := resp.Data["message"]; !ok {
					//делаем выход из аккаунта
					respLO := newMapMess(typeLogOut, conn, nil)
					status := logOut(login, db)
					if status {
						respLO.Data["login"] = login //сохраним а потом удалим из отправки чтобы нормально закрыть все сокеты
						login = fmt.Sprint(resp.Data["login"])
					}
					respLO.send()
				}
				resp.send()
			}
		case typeLogOut: //отправка default
			{
				if login != "" {
					resp := newMapMess(typeLogOut, conn, nil)
					status := logOut(login, db)
					if status {
						resp.Data["authorizedFlag"] = false
						resp.Data["login"] = login //сохраним а потом удалим из отправки чтобы нормально закрыть все сокеты
					}
					resp.send()
				}
			}
		case typeCheckConn: //отправка default
			{
				resp := newMapMess(typeCheckConn, conn, nil)
				statusDB := false
				_, err := db.Exec(`SELECT * FROM public.accounts;`)
				if err == nil {
					statusDB = true
				}
				resp.Data["statusBD"] = statusDB
				var tcpPackage = tcpConnect.TCPMessage{
					TCPType:     tcpConnect.TypeState,
					User:        login,
					Idevice:     -1,
					Data:        0,
					From:        tcpConnect.FromMapSoc,
					CommandType: typeDButton,
				}
				tcpPackage.SendToTCPServer()

				resp.send()
			}
		}
	}
}

//MapBroadcast передатчик для карты (map)
func MapBroadcast(db *sqlx.DB) {
	connectedUsersOnMap = make(map[*websocket.Conn]bool)
	writeMap = make(chan MapSokResponse, 50)

	crossReadTick := time.NewTicker(time.Second * 5)
	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		crossReadTick.Stop()
	}()
	oldTFs := selectTL(db)
	for {
		select {
		case <-crossReadTick.C:
			{
				if len(connectedUsersOnMap) > 0 {
					newTFs := selectTL(db)
					var tempTF []data.TrafficLights
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
		case <-crossSock.MapRepaint:
			{
				if len(connectedUsersOnMap) > 0 {
					time.Sleep(time.Second * time.Duration(config.GlobalConfig.DBConfig.DBWait))
					oldTFs = selectTL(db)
					resp := newMapMess(typeRepaint, nil, nil)
					resp.Data["tflight"] = oldTFs
					data.FillMapAreaZone()
					GSRepaint <- true
					data.CacheArea.Mux.Lock()
					resp.Data["areaZone"] = data.CacheArea.Areas
					data.CacheArea.Mux.Unlock()
					for conn := range connectedUsersOnMap {
						_ = conn.WriteJSON(resp)
					}
				}
			}
		case <-pingTicker.C:
			{
				for conn := range connectedUsersOnMap {
					_ = conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		case crossUsers := <-crossSock.CrossUsersForMap:
			{
				resp := newMapMess(typeEditCrossUsers, nil, nil)
				resp.Data["editCrossUsers"] = crossUsers
				for conn := range connectedUsersOnMap {
					_ = conn.WriteJSON(resp)
				}
			}
		case msg := <-tcpConnect.TCPRespMap:
			{
				resp := newMapMess(typeCheckConn, nil, nil)
				resp.Data["statusS"] = msg.Status
				for conn := range connectedUsersOnMap {
					_ = conn.WriteJSON(resp)
				}
			}
		case msg := <-writeMap:
			switch msg.Type {
			case typeLogOut:
				{
					login := fmt.Sprint(msg.Data["login"])
					delete(msg.Data, "login")
					_ = msg.conn.WriteJSON(msg)
					chat.UserLogoutChat <- login
					crossSock.UserLogoutCrControl <- login
					crossSock.UserLogoutCross <- login
					techArm.UserLogoutTech <- login
					userLogout <- login
				}
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
