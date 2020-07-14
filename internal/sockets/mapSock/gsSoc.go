package mapSock

import (
	"encoding/json"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

var connectOnGS map[*websocket.Conn]string
var writeGS chan GSSokResponse
var GSRepaint chan bool
var userLogout chan string

//GSReader обработчик открытия сокета для режима зеленой улицы
func GSReader(conn *websocket.Conn, mapContx map[string]string, db *sqlx.DB) {
	connectOnGS[conn] = mapContx["login"]
	{
		resp := newGSMess(typeMapInfo, conn, mapOpenInfo(db))
		resp.Data["modes"] = getAllModes(db)
		data.CacheArea.Mux.Lock()
		resp.Data["areaBox"] = data.CacheArea.Areas
		data.CacheArea.Mux.Unlock()
		resp.send()
	}
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//закрытие коннекта
			resp := newGSMess(typeClose, conn, nil)
			resp.send()
			return
		}

		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			resp := newGSMess(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeCreateMode:
			{
				temp := Mode{}
				_ = json.Unmarshal(p, &temp)
				resp := newGSMess(typeCreateMode, conn, nil)
				err := temp.create(db)
				if err != nil {
					resp.Data[typeError] = errCantWriteInBD
				} else {
					resp.Data["mode"] = temp
				}
				resp.send()
			}
		case typeJump: //отправка default
			{
				location := &data.Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := newGSMess(typeJump, conn, nil)
				resp.Data["boxPoint"] = box
				resp.send()
			}
		}
	}
}

//GSBroadcast передатчик для ЗУ
func GSBroadcast(db *sqlx.DB) {
	connectOnGS = make(map[*websocket.Conn]string)
	writeGS = make(chan GSSokResponse)
	userLogout = make(chan string)
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
				if len(connectOnGS) > 0 {
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
						resp := newGSMess(typeTFlight, nil, nil)
						resp.Data["tflight"] = tempTF
						for conn := range connectOnGS {
							_ = conn.WriteJSON(resp)
						}
					}
				}
			}
		case <-GSRepaint:
			{
				time.Sleep(time.Second * time.Duration(config.GlobalConfig.DBConfig.DBWait))
				oldTFs = selectTL(db)
				resp := newGSMess(typeRepaint, nil, nil)
				resp.Data["tflight"] = oldTFs
				data.FillMapAreaBox()
				data.CacheArea.Mux.Lock()
				resp.Data["areaBox"] = data.CacheArea.Areas
				data.CacheArea.Mux.Unlock()
				for conn := range connectOnGS {
					_ = conn.WriteJSON(resp)
				}
			}
		case <-pingTicker.C:
			{
				for conn := range connectOnGS {
					_ = conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		case login := <-userLogout:
			{
				for conn, user := range connectOnGS {
					if user == login {
						_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "пользователь вышел из системы"))
					}
				}
			}
		case msg := <-writeGS:
			switch msg.Type {
			case typeClose:
				{
					delete(connectOnGS, msg.conn)
				}
			case typeCreateMode:
				{
					if _, ok := msg.Data[typeError]; !ok {
						for conn := range connectOnGS {
							_ = conn.WriteJSON(msg)
						}
					} else {
						_ = msg.conn.WriteJSON(msg)
					}
				}
			default:
				{
					_ = msg.conn.WriteJSON(msg)
				}
			}
		}
	}
}
