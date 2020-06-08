package data

import (
	"github.com/gorilla/websocket"
	"time"
)

var ConnectedCrossUsers map[string][]CrossConn
var WriteCrossMessage chan CrossSokResponse

//delConn удаление подключения из массива подключений пользователя
func delConn(login string, conn *websocket.Conn) {
	for index, userConn := range ConnectedCrossUsers[login] {
		if userConn.Conn == conn {
			ConnectedCrossUsers[login][index] = ConnectedCrossUsers[login][len(ConnectedCrossUsers[login])-1]
			//TODO тут может быть что-то не так
			//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			ConnectedCrossUsers[login][len(ConnectedCrossUsers[login])-1] = CrossConn{} //!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
			ConnectedCrossUsers[login] = ConnectedCrossUsers[login][:len(ConnectedCrossUsers[login])-1]
			break
		}
	}
}

//CrossConn соединение
type CrossConn struct {
	Pos   CrossEditInfo
	Conn  *websocket.Conn //подключение
	First bool
	Login string
}

type CrossEditInfo struct {
	Region string
	Area   string
	Id     int
}

func CrossReader(crossConn CrossConn) {
	//дропаю соединение, если перекресток уже открыт у пользователя
	for _, crConn := range ConnectedCrossUsers[crossConn.Login] {
		if crConn.Pos == crossConn.Pos {
			resp := crossSokMessage(typeError, crossConn, nil)
			resp.Data["message"] = ErrorMessage{Error: errDoubleOpeningDevice}
			_ = crossConn.Conn.WriteJSON(resp)
			_ = crossConn.Conn.Close()
			return
		}
	}

	//это точно не тот перекресток который уже открыт
	//проверка первым ли он занял перекресток
	for user, connInfos := range ConnectedCrossUsers {
		if user == crossConn.Login {
			continue
		}
		for _, crConn := range connInfos {
			if crConn.Pos == crossConn.Pos {
				crossConn.First = false
				break
			}
		}
	}
	crossConn.First = true
	//если все ОК идем дальше
	ConnectedCrossUsers[crossConn.Login] = append(ConnectedCrossUsers[crossConn.Login], crossConn)

	{
		resp := blob(crossConn.Pos)
		resp.crConn.Conn = crossConn.Conn
		resp.Data["first"] = crossConn.First
		resp.send()
	}

	for {
		_, p, err := crossConn.Conn.ReadMessage()
		if err != nil {
			delConn(crossConn.Login, crossConn.Conn)
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := mapSokMessage(typeError, crossConn.Conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		// case
		}
	}
}

type phaseInfo struct {
	idevice int  `json:"-"`
	Fdk     int  `json:"fdk"`
	Tdk     int  `json:"tdk"`
	Pdk     bool `json:"pdk"`
}

func (p *phaseInfo) get() error {
	err := GetDB().QueryRow(`SELECT Fdk, tdk, pdk FROM public.devices WHERE id = $1`, p.idevice).Scan(&p.Fdk, &p.Tdk, &p.Pdk)
	if err != nil {
		return err
	}
	return nil
}

func blob(pos CrossEditInfo) CrossSokResponse {
	var (
		dgis     string
		stateStr string
		phase    phaseInfo
	)
	TLignt := TrafficLights{Area: AreaInfo{Num: pos.Area}, Region: RegionInfo{Num: pos.Region}, ID: pos.Id}
	rowsTL := GetDB().QueryRow(`SELECT area, subarea, idevice, dgis, describ, state FROM public.cross WHERE region = $1 and id = $2 and area = $3`, pos.Region, pos.Id, pos.Area)
	err := rowsTL.Scan(&TLignt.Area.Num, &TLignt.Subarea, &TLignt.Idevice, &dgis, &TLignt.Description, &stateStr)
	if err != nil {
		resp := crossSokMessage(typeError, CrossConn{}, nil)
		resp.Data["message"] = "No result at these points, table cross"
		return resp
	}
	TLignt.Points.StrToFloat(dgis)
	//Состояние светофора!
	rState, err := ConvertStateStrToStruct(stateStr)
	if err != nil {
		resp := crossSokMessage(typeError, CrossConn{}, nil)
		resp.Data["message"] = "failed to parse cross information"
		return resp
	}

	resp := crossSokMessage(typeCrossBuild, CrossConn{}, nil)
	CacheInfo.Mux.Lock()
	TLignt.Region.NameRegion = CacheInfo.MapRegion[TLignt.Region.Num]
	TLignt.Area.NameArea = CacheInfo.MapArea[TLignt.Region.NameRegion][TLignt.Area.Num]
	TLignt.Sost.Num = rState.StatusDevice
	TLignt.Sost.Description = CacheInfo.MapTLSost[TLignt.Sost.Num]
	CacheInfo.Mux.Unlock()
	phase.idevice = TLignt.Idevice
	err = phase.get()
	if err != nil {
		resp.Data["phase"] = phaseInfo{}
	} else {
		resp.Data["phase"] = phase
	}
	resp.Data["cross"] = TLignt
	resp.Data["state"] = rState
	return resp
}

func CrossBroadcast() {
	ConnectedCrossUsers = make(map[string][]CrossConn)
	WriteCrossMessage = make(chan CrossSokResponse)

	readTick := time.Tick(time.Second * 1)
	for {
		select {
		case <-readTick:
			{

			}
		case msg := <-WriteCrossMessage:
			{
				if err := msg.crConn.Conn.WriteJSON(msg); err != nil {
					delConn(msg.crConn.Login, msg.crConn.Conn)
					_ = msg.crConn.Conn.Close()
					return
				}
			}
		}
	}
}
