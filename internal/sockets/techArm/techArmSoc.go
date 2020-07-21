package techArm

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/JanFant/TLServer/internal/sockets"
	"github.com/JanFant/TLServer/logger"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/ag-server/comm"
	"reflect"
	"strconv"
	"time"
)

var connectedUsersTechArm map[*websocket.Conn]ArmInfo
var writeArm chan armResponse
var TArmNewCrossData chan bool
var UserLogoutTech chan string

const devUpdate = time.Second * 1

//ArmTechReader обработчик открытия сокета для тех арм
func ArmTechReader(conn *websocket.Conn, reg int, area []string, login string, db *sqlx.DB) {
	var armInfo = ArmInfo{Region: reg, Area: area, Login: login}
	connectedUsersTechArm[conn] = armInfo
	//сформировать список перекрестков которые необходимы пользователю
	{
		var tempCrosses []CrossInfo
		crosses := getCross(armInfo.Region, db)
		for _, cross := range crosses {
			for _, area := range armInfo.Area {
				tArea, _ := strconv.Atoi(area)
				if cross.Area == tArea {
					tempCrosses = append(tempCrosses, cross)
				}
			}
		}
		resp := newArmMess(typeArmInfo, conn, nil)
		resp.Data[typeCrosses] = tempCrosses

		var tempDevises []DevInfo
		devices := getDevice(db)
		for _, dev := range devices {
			for _, area := range armInfo.Area {
				tArea, _ := strconv.Atoi(area)
				if dev.Area == tArea && dev.Region == armInfo.Region {
					tempDevises = append(tempDevises, dev)
				}
			}
		}
		resp.Data[typeDevices] = tempDevises
		resp.Data["gps"] = GPSInfo
		resp.send()
	}

	fmt.Println("tech : ", connectedUsersTechArm)
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//закрытие коннекта
			resp := newArmMess(typeClose, conn, nil)
			resp.send()
			return
		}

		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			resp := newArmMess(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = armInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        armInfo.Login,
					TCPType:     tcpConnect.TypeDispatch,
					Idevice:     arm.ID,
					Data:        arm,
					From:        tcpConnect.TechArmSoc,
					CommandType: typeDButton,
				}
				mess.SendToTCPServer()
			}
		case typeGPS:
			{
				gps := comm.ChangeProtocol{}
				_ = json.Unmarshal(p, &gps)
				gps.User = armInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        armInfo.Login,
					TCPType:     tcpConnect.TypeChangeProtocol,
					Idevice:     gps.ID,
					Data:        gps,
					From:        tcpConnect.TechArmSoc,
					CommandType: typeGPS,
				}
				mess.SendToTCPServer()
			}
		}
	}
}

//ArmTechBroadcast передатчик для тех арм (techArm)
func ArmTechBroadcast(db *sqlx.DB) {
	connectedUsersTechArm = make(map[*websocket.Conn]ArmInfo)
	writeArm = make(chan armResponse, 50)

	TArmNewCrossData = make(chan bool)
	UserLogoutTech = make(chan string)
	sockets.DispatchMessageFromAnotherPlace = make(chan sockets.DBMessage, 5)

	readTick := time.NewTicker(devUpdate)
	defer readTick.Stop()
	var oldDevice = getDevice(db)
	for {
		select {
		case <-readTick.C:
			{
				if len(connectedUsersTechArm) > 0 {
					newDevice := getDevice(db)
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
						for conn, arm := range connectedUsersTechArm {
							var tDev []DevInfo
							for _, area := range arm.Area {
								tArea, _ := strconv.Atoi(area)
								for _, dev := range tempDev {
									if dev.Area == tArea && dev.Region == arm.Region {
										tDev = append(tDev, dev)
									}
								}
							}
							if len(tDev) > 0 {
								resp := newArmMess(typeDevices, conn, nil)
								resp.Data[typeDevices] = tDev
								_ = conn.WriteJSON(resp)
							}
						}
					}
				}
			}
		case <-TArmNewCrossData:
			{
				if len(connectedUsersTechArm) > 0 {
					time.Sleep(time.Second * time.Duration(config.GlobalConfig.DBConfig.DBWait))
					crosses := getCross(-1, db)
					for conn, arm := range connectedUsersTechArm {
						var tempCrosses []CrossInfo
						for _, area := range arm.Area {
							tArea, _ := strconv.Atoi(area)
							for _, cross := range crosses {
								if cross.Region == arm.Region && cross.Area == tArea {
									tempCrosses = append(tempCrosses, cross)
								}
							}
						}
						resp := newArmMess(typeCrosses, conn, nil)
						resp.Data[typeCrosses] = tempCrosses
						_ = resp.conn.WriteJSON(resp)
					}
				}
			}
		case login := <-UserLogoutTech:
			{
				for conn, armInfo := range connectedUsersTechArm {
					if armInfo.Login == login {
						msg := closeMessage{Type: typeClose, Message: "пользователь вышел из системы"}
						_ = conn.WriteJSON(msg)
						//_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "пользователь вышел из системы"))
					}
				}
			}
		case msg := <-tcpConnect.TArmGetTCPResp:
			{
				resp := newArmMess("", nil, nil)
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
				case typeGPS:
					{
						resp.Type = typeGPS
						resp.Data["status"] = msg.Status
					}
				}

				for conn, armInfo := range connectedUsersTechArm {
					if armInfo.Login == msg.User {
						_ = conn.WriteJSON(resp)
					}
				}
			}
		case msg := <-writeArm:
			switch msg.Type {
			case typeClose:
				{
					delete(connectedUsersTechArm, msg.conn)
				}
			default:
				{
					_ = msg.conn.WriteJSON(msg)
				}
			}
		}
	}
}

func getCross(reg int, db *sqlx.DB) []CrossInfo {
	var (
		temp    CrossInfo
		crosses []CrossInfo
		sqlStr  = `SELECT region,
 					area, 
 					id,
  					idevice, 
  					describ, 
  					subarea, 
  					state->'arrays'->'type',
  					state->'phone' 
  					FROM public.cross`
	)
	if reg != -1 {
		sqlStr += fmt.Sprintf(` WHERE region = %v`, reg)
	}
	rows, err := db.Query(sqlStr)
	if err != nil {
		logger.Error.Println("|Message: Error get Cross from BD ", err.Error())
		return make([]CrossInfo, 0)
	}
	for rows.Next() {
		_ = rows.Scan(&temp.Region,
			&temp.Area,
			&temp.ID,
			&temp.Idevice,
			&temp.Describe,
			&temp.Subarea,
			&temp.ArrayType,
			&temp.Phone)
		crosses = append(crosses, temp)
	}
	return crosses
}

func getDevice(db *sqlx.DB) []DevInfo {
	var (
		temp    DevInfo
		devices []DevInfo
		dStr    string
	)
	rows, err := db.Query(`SELECT c.region, 
									c.area, 
									c.idevice, 
									d.device 
									FROM public.cross as c, public.devices as d WHERE c.idevice IN(d.id);`)
	if err != nil {
		logger.Error.Println("|Message: Error get Device from BD ", err.Error())
		return make([]DevInfo, 0)
	}
	for rows.Next() {
		_ = rows.Scan(&temp.Region, &temp.Area, &temp.Idevice, &dStr)
		_ = json.Unmarshal([]byte(dStr), &temp.Device)
		temp.ModeRdk = modeRDK[temp.Device.DK.RDK]
		temp.TechMode = texMode[temp.Device.TechMode]
		devices = append(devices, temp)
	}
	return devices
}
