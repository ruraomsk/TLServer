package techArm

import (
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"reflect"
	"strconv"
	"time"
)

//HubTechArm структура хаба для techArm
type HubTechArm struct {
	clients    map[*ClientTechArm]bool
	broadcast  chan armResponse
	register   chan *ClientTechArm
	unregister chan *ClientTechArm
}

//NewTechArmHub создание хаба
func NewTechArmHub() *HubTechArm {
	return &HubTechArm{
		broadcast:  make(chan armResponse),
		clients:    make(map[*ClientTechArm]bool),
		register:   make(chan *ClientTechArm),
		unregister: make(chan *ClientTechArm),
	}
}

//Run запуск хаба для techArm
func (h *HubTechArm) Run() {
	UserLogoutTech = make(chan string)
	sockets.DispatchMessageFromAnotherPlace = make(chan sockets.DBMessage, 50)

	readDeviceTick := time.NewTicker(devUpdate)
	readCrossTick := time.NewTicker(devUpdate)
	checkValidityTicker := time.NewTicker(checkTokensValidity)
	defer func() {
		readDeviceTick.Stop()
		readCrossTick.Stop()
		checkValidityTicker.Stop()
	}()

	var (
		oldDevice = getDevice()
		oldCross  = getCross(-1)
	)
	for {
		select {
		case <-readDeviceTick.C:
			{
				if len(h.clients) > 0 {
					newDevice := getDevice()
					var (
						tempDev []DevInfo
					)
					for _, nDev := range newDevice {
						flagNew := true
						for _, oDev := range oldDevice {
							if oDev.Idevice == nDev.Idevice {
								flagNew = false
								if oDev.Device.LastOperation != nDev.Device.LastOperation ||
									!reflect.DeepEqual(oDev.Device.GPS, nDev.Device.GPS) ||
									!reflect.DeepEqual(oDev.Device.Error, nDev.Device.Error) ||
									!reflect.DeepEqual(oDev.Device.Status, nDev.Device.Status) ||
									!reflect.DeepEqual(oDev.Device.StatusCommandDU, nDev.Device.StatusCommandDU) ||
									oDev.Device.TechMode != nDev.Device.TechMode ||
									oDev.Device.PK != nDev.Device.PK ||
									oDev.Device.CK != nDev.Device.CK ||
									oDev.Device.NK != nDev.Device.NK ||
									!reflect.DeepEqual(oDev.Device.DK, nDev.Device.DK) {
									tempDev = append(tempDev, nDev)
									break
								}
							}
						}
						if flagNew {
							tempDev = append(tempDev, nDev)
						}
					}
					oldDevice = newDevice
					if len(tempDev) > 0 {
						for client := range h.clients {
							var tDev []DevInfo
							for _, area := range client.armInfo.Area {
								tArea, _ := strconv.Atoi(area)
								for _, dev := range tempDev {
									if dev.Area == tArea && dev.Region == client.armInfo.Region {
										tDev = append(tDev, dev)
									}
								}
							}
							if len(tDev) > 0 {
								resp := newArmMess(typeDevices, nil)
								resp.Data[typeDevices] = tDev
								client.send <- resp
							}
						}
					}
				}
			}
		case <-readCrossTick.C:
			{
				if len(h.clients) > 0 {
					newCross := getCross(-1)
					if len(oldCross) != len(newCross) {
						for client := range h.clients {
							var tempCrosses []CrossInfo
							for _, area := range client.armInfo.Area {
								tArea, _ := strconv.Atoi(area)
								for _, cross := range newCross {
									if cross.Region == client.armInfo.Region && cross.Area == tArea {
										tempCrosses = append(tempCrosses, cross)
									}
								}
							}
							resp := newArmMess(typeCrosses, nil)
							resp.Data[typeCrosses] = tempCrosses
							client.send <- resp
						}
					} else {
						flagNew := false
						for _, nCr := range newCross {
							for _, oCr := range oldCross {
								if nCr.Idevice == oCr.Idevice {
									if !reflect.DeepEqual(nCr, oCr) {
										flagNew = true
										break
									}
								}
							}
							if flagNew {
								break
							}
						}
						if flagNew {
							for client := range h.clients {
								var tCross []CrossInfo
								for _, area := range client.armInfo.Area {
									tArea, _ := strconv.Atoi(area)
									for _, cross := range newCross {
										if cross.Area == tArea && cross.Region == client.armInfo.Region {
											tCross = append(tCross, cross)
										}
									}
								}
								if len(tCross) > 0 {
									resp := newArmMess(typeCrosses, nil)
									resp.Data[typeCrosses] = tCross
									client.send <- resp
								}
							}
						}
					}
					oldCross = newCross
				}
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
		case login := <-UserLogoutTech:
			{
				resp := newArmMess(typeClose, nil)
				resp.Data["message"] = "пользователь вышел из системы"
				for client := range h.clients {
					if client.armInfo.AccInfo.Login == login {
						client.send <- resp
					}
				}
			}
		case <-checkValidityTicker.C:
			{
				for client := range h.clients {
					if client.armInfo.AccInfo.Valid() != nil {
						msg := newArmMess(typeClose, nil)
						msg.Data["message"] = "вышло время сеанса пользователя"
						client.send <- msg
					}
				}
			}
		case msg := <-tcpConnect.TCPRespTArm:
			{
				resp := newArmMess("", nil)
				switch msg.CommandType {
				case typeDButton:
					{
						resp.Type = typeDButton
						resp.Data["status"] = msg.Status
						if msg.Status {
							resp.Data["command"] = msg.Data
						}
						var message = sockets.DBMessage{Data: resp, Idevice: msg.Idevice}
						sockets.DispatchMessageFromAnotherPlace <- message
					}
				case typeGPRS:
					{
						resp.Type = typeGPRS
						resp.Data["status"] = msg.Status
					}
				}

				for client := range h.clients {
					if client.armInfo.AccInfo.Login == msg.User {
						client.send <- resp
					}
				}
			}
		}
	}
}
