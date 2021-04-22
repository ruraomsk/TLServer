package dispatchControl

import (
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/device"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/maps"
	"github.com/ruraomsk/TLServer/internal/sockets/maps/mainMap"
	"github.com/ruraomsk/ag-server/comm"
	"time"
)

//HubDispCtrl структура хаба для Диспетчерского управления
type HubDispCtrl struct {
	clients    map[*ClientDC]bool
	broadcast  chan dCResponse
	register   chan *ClientDC
	unregister chan *ClientDC
}

//NewDCHub создание хаба
func NewDCHub() *HubDispCtrl {
	return &HubDispCtrl{
		broadcast:  make(chan dCResponse),
		clients:    make(map[*ClientDC]bool),
		register:   make(chan *ClientDC),
		unregister: make(chan *ClientDC),
	}
}

//Run запуск хаба для DC
func (h *HubDispCtrl) Run() {

	crossReadTick := time.NewTicker(crossPeriod)
	deviceReadTick := time.NewTicker(devicePeriod)
	checkValidityTicker := time.NewTicker(checkTokensValidity)
	defer func() {
		crossReadTick.Stop()
		deviceReadTick.Stop()
		checkValidityTicker.Stop()
	}()

	oldTFs := maps.SelectTL()
	for {
		select {
		case <-deviceReadTick.C:
			//logger.Debug.Printf("deviceReadTick in")
			if len(h.clients) == 0 {
				//logger.Debug.Printf("deviceReadTick empty")
				break
			}
			for c := range h.clients {
				if len(c.devices) == 0 || !c.sendPhases {
					//logger.Debug.Printf("deviceReadTick empty two")
					continue
				}
				if c.sendPhases {
					ph := getPhases(c.devices)
					if len(ph) != 0 {
						resp := newDCMess(typePhases, nil)
						resp.Data[typePhases] = ph
						c.send <- resp
					}
					//logger.Debug.Printf("deviceReadTick send %v",getPhases(c.devices, c.db))
				}
			}
		case <-crossReadTick.C:
			{
				//logger.Debug.Printf("crossReadTick in")
				if len(h.clients) == 0 {
					//logger.Debug.Printf("crossReadTick zerro")
					break
				}
				newTFs := maps.SelectTL()
				if len(newTFs) != len(oldTFs) {
					//logger.Debug.Printf("crossReadTick in len(newTFs) != len(oldTFs)")
					resp := newDCMess(typeRepaint, nil)
					resp.Data["tflight"] = newTFs
					data.CacheArea.Mux.Lock()
					resp.Data["areaZone"] = data.CacheArea.Areas
					data.CacheArea.Mux.Unlock()
					for client := range h.clients {
						client.send <- resp
					}
				} else {
					//logger.Debug.Printf("crossReadTick in NOT len(newTFs) != len(oldTFs)")
					var (
						tempTF   []data.TrafficLights
						flagFill = false
					)
					for _, nTF := range newTFs {
						var flagAdd = true
						for _, oTF := range oldTFs {
							if oTF.Idevice == nTF.Idevice {
								flagAdd = false
								if oTF.Sost.Num != nTF.Sost.Num || oTF.Description != nTF.Description {
									flagAdd = true
								}
								if oTF.Subarea != nTF.Subarea {
									flagAdd = true
									flagFill = true
								}
								break
							}
						}
						if flagAdd {
							tempTF = append(tempTF, nTF)
						}
					}
					//logger.Debug.Printf("crossReadTick in len(tempTF) > 0")
					if len(tempTF) > 0 {
						resp := newDCMess(typeTFlight, nil)
						if flagFill {
							data.FillMapAreaZone()
							data.CacheArea.Mux.Lock()
							resp.Data["areaZone"] = data.CacheArea.Areas
							data.CacheArea.Mux.Unlock()
						}
						resp.Data["tflight"] = tempTF
						for client := range h.clients {
							client.send <- resp
						}
					}
				}
				oldTFs = newTFs

			}
		case client := <-h.register:
			{
				h.clients[client] = true
			}
		case client := <-h.unregister:
			{
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
					_ = client.conn.Close()
				}
				client.sendPhases = false
				if len(client.devices) != 0 {
					arm := comm.CommandARM{Command: 4, Params: 0, User: client.cInfo.Login}
					var mess = tcpConnect.TCPMessage{
						User:        client.cInfo.Login,
						TCPType:     tcpConnect.TypeDispatch,
						From:        tcpConnect.FromCrossSoc,
						CommandType: typeDButton,
						Pos:         sockets.PosInfo{},
					}

					device.GlobalDevEdit.Mux.Lock()
					for _, dev := range client.devices {
						tDev := device.GlobalDevEdit.MapDevices[dev]
						tDev.BusyCount--
						if tDev.BusyCount == 0 && tDev.TurnOnFlag == true {
							arm.ID = dev
							mess.Idevice = arm.ID
							mess.Data = arm
							mess.SendToTCPServer()
							tDev.TurnOnFlag = false
						}
						device.GlobalDevEdit.MapDevices[dev] = tDev
					}
					device.GlobalDevEdit.Mux.Unlock()
				}
			}
		case mess := <-h.broadcast:
			{
				for client := range h.clients {
					select {
					case client.send <- mess:
					default:
						delete(h.clients, client)
						close(client.send)
					}
				}
			}
		case login := <-mainMap.UserLogoutGS:
			{
				resp := newDCMess(typeClose, nil)
				resp.Data["message"] = "пользователь вышел из системы"
				for client := range h.clients {
					if client.cInfo.Login == login {
						client.send <- resp
					}
				}
			}
		case <-checkValidityTicker.C:
			{
				for client := range h.clients {
					if client.cInfo.Valid() != nil {
						msg := newDCMess(typeClose, nil)
						msg.Data["message"] = "вышло время сеанса пользователя"
						client.send <- msg
					}
				}
			}
		case msg := <-tcpConnect.TCPRespDC:
			{
				resp := newDCMess(typeDButton, nil)
				resp.Data["status"] = msg.Status
				if msg.Status {
					resp.Data["command"] = msg.Data
					var message = sockets.DBMessage{Data: resp, Idevice: msg.Idevice}
					sockets.DispatchMessageFromAnotherPlace <- message
				}
				for client := range h.clients {
					if client.cInfo.Login == msg.User {
						client.send <- resp
					}
				}
			}
		}
	}
}
