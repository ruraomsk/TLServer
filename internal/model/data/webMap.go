package data

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/internal/model/locations"
	u "github.com/JanFant/TLServer/internal/utils"
	"github.com/gorilla/websocket"
)

func ReaderStrong(conn *websocket.Conn) {
	ch := make(chan mapMessage)
	go broadcast(conn, ch)

	var message mapMessage
	message.send(typeMapInfo, mapOpenInfo(), ch)

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			message.send(closeSocket, nil, ch)
			//закрытие коннекта
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			//var myError = ErrorMessage{Error: errUnregisteredMessageType}
			//message.send(typeError, myError.toString(), ch)
		}
		switch typeSelect {
		case typeNewBox:
			{
				obj := make(map[string]interface{})
				obj["box"] = string(p)
				message.send(typeNewBox, obj, ch)
			}

		case typeUpdate:
			{

			}
		case typeStatus:
			{

			}
		case typeJump:
			{

			}
		}
	}
}

func broadcast(conn *websocket.Conn, write chan mapMessage) {
	var box locations.BoxPoint
	{
		location := &Locations{}
		box, _ = location.MakeBoxPoint()
	}

	oldTFs := GetLightsFromBD(box)

	for {
		select {
		case msg := <-write:
			{
				switch msg.Type {
				case closeSocket:
					{
						return
					}
				case typeNewBox:
					{
						data := u.ParserInterface(msg.Data)
						fmt.Println(data)
						box.ToStruct(data["box"])
						fmt.Println(box)
						newTFs := GetLightsFromBD(box)
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
							var message mapMessage
							message.Data = make(map[string]interface{})
							message.Type = typeTFlight
							message.Data["tflight"] = tempTF
							_ = conn.WriteJSON(message)
						}

					}
				default:
					if err := conn.WriteJSON(msg); err != nil {
						_ = conn.Close()
						return
					}

				}
			}
		}
	}
}

func mapOpenInfo() (obj map[string]interface{}) {
	obj = make(map[string]interface{})

	location := &Locations{}
	box, _ := location.MakeBoxPoint()
	obj["boxPoint"] = &box
	obj["tflight"] = GetLightsFromBD(box)
	obj["flagUnauthorized"] = false
	obj["yaKey"] = license.LicenseFields.YaKey

	//собираю в кучу регионы для отображения
	chosenRegion := make(map[string]string)
	CacheInfo.Mux.Lock()
	for first, second := range CacheInfo.MapRegion {
		chosenRegion[first] = second
	}
	delete(chosenRegion, "*")
	obj["regionInfo"] = chosenRegion

	//собираю в кучу районы для отображения
	chosenArea := make(map[string]map[string]string)
	for first, second := range CacheInfo.MapArea {
		chosenArea[first] = make(map[string]string)
		chosenArea[first] = second
	}
	delete(chosenArea, "Все регионы")
	CacheInfo.Mux.Unlock()
	obj["areaInfo"] = chosenArea
	return
}
