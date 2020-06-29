package techArm

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

var connectedUsersTechArm map[*websocket.Conn]ArmInfo
var writeArm chan armResponse

const pingPeriod = time.Second * 30

//ArmTechReader обработчик открытия сокета для тех арм
func ArmTechReader(conn *websocket.Conn, reg int, area []string, db *sqlx.DB) {
	var armInfo = ArmInfo{Region: reg, Area: area}
	fmt.Println(armInfo)
	connectedUsersTechArm[conn] = armInfo

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			//закрытие коннекта
			resp := newArmMess(typeClose, conn, nil)
			resp.send()
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := newArmMess(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case "nothing":
			{

			}
		}
	}
}

//ArmTechBroadcast передатчик для тех арм (techArm)
func ArmTechBroadcast(db *sqlx.DB) {
	connectedUsersTechArm = make(map[*websocket.Conn]ArmInfo)
	writeArm = make(chan armResponse)

	readTick := time.NewTicker(time.Second * 1)
	defer readTick.Stop()
	//var oldCross []crossInfo
	for {
		select {
		case <-readTick.C:
			{
				if len(connectedUsersTechArm) > 0 {
					var newRegions map[int][]crossInfo
					for _, arm := range connectedUsersTechArm {
						newRegions[arm.Region] = nil
					}
					//собрал все кросы по полученным регионам
					for reg := range newRegions {
						newRegions[reg] = getCross(reg, db)
					}

					for range newRegions {

					}

					//
					//rows, _ := db.Query(`SELECT region, area, id, idevice, describ FROM public.cross`)

					//rows, _ := db.Query(`SELECT device FROM public.devices`)
					//for rows.Next() {
					//	var (
					//		temp string
					//		dev  pudge.Controller
					//	)
					//	_ = rows.Scan(&temp)
					//	_ = json.Unmarshal([]byte(temp), &dev)
					//
					//}

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

func getCross(reg int, db *sqlx.DB) []crossInfo {
	var (
		temp    crossInfo
		crosses []crossInfo
	)
	rows, _ := db.Query(`SELECT region, area, id, idevice, describ FROM public.cross WHERE region = $1`, reg)
	for rows.Next() {
		_ = rows.Scan(&temp.region, &temp.area, &temp.id, &temp.idevice, &temp.describ)
		crosses = append(crosses, temp)
	}
	return crosses
}
