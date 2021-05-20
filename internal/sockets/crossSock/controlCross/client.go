package controlCross

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/crossCreator"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/sockets"
	"github.com/ruraomsk/TLServer/internal/sockets/crossSock"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	agspudge "github.com/ruraomsk/ag-server/pudge"
	"reflect"
	"strconv"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod          = (pongWait * 9) / 10
	checkTokensValidity = time.Minute * 1
)

//ClientControlCr информация о подключившемся пользователе
type ClientControlCr struct {
	hub       *HubControlCross
	conn      *websocket.Conn
	send      chan ControlSokResponse
	regStatus chan bool
	crossInfo *crossSock.CrossInfo
}

var UserLogoutCrControl chan string

type History struct {
	Time  time.Time `json:"time"`
	Login string    `json:"login"`
}
type GetHistory struct {
	Type string      `json:"type"`
	Data DataHistory `json:"data""`
}
type DataHistory struct {
	Time string `json:"time"`
}

func (c ClientControlCr) getHistory() []History {
	hs := make([]History, 0)
	var h History
	db, id := data.GetDB()
	defer data.FreeDB(id)
	w := fmt.Sprintf("Select tm,login from public.history where region=%s and area=%s and id=%d;",
		c.crossInfo.Pos.Region, c.crossInfo.Pos.Area, c.crossInfo.Pos.Id)
	rows, err := db.Query(w)
	if err != nil {
		logger.Error.Printf("/History %s", err.Error())
		return hs
	}
	for rows.Next() {
		rows.Scan(&h.Time, &h.Login)
		hs = append(hs, h)
	}
	return hs
}
func (c *ClientControlCr) getdHistoryData(tm string) map[string]interface{} {
	resp := make(map[string]interface{})
	db, id := data.GetDB()
	defer data.FreeDB(id)
	sqlStr := fmt.Sprintf("SELECT state FROM public.history WHERE region=%s and area=%s and id=%d and tm='%s' limit 1;",
		c.crossInfo.Pos.Region, c.crossInfo.Pos.Area, c.crossInfo.Pos.Id, tm)
	rows, err := db.Query(sqlStr)
	if err != nil {
		logger.Error.Println("|Message: control socket (send History), DB not respond : ", err.Error())
		resp["status"] = false
		resp["message"] = "сервер баз данных не отвечает или нет данных"
		return resp
	}
	var strRow []byte
	var oldstate agspudge.Cross
	var nowstate agspudge.Cross

	for rows.Next() {
		_ = rows.Scan(&strRow)
		_ = json.Unmarshal(strRow, &oldstate)
	}
	_ = rows.Close()
	resp[typeSendHistory] = oldstate
	sqlStr = fmt.Sprintf("SELECT state FROM public.\"cross\" WHERE region=%s and area=%s and id=%dlimit 1;",
		c.crossInfo.Pos.Region, c.crossInfo.Pos.Area, c.crossInfo.Pos.Id)
	rows, err = db.Query(sqlStr)
	if err != nil {
		logger.Error.Println("|Message: control socket (send History), DB not respond : ", err.Error())
		resp["status"] = false
		resp["message"] = "сервер баз данных не отвечает или нет данных"
		return resp
	}
	for rows.Next() {
		_ = rows.Scan(&strRow)
		_ = json.Unmarshal(strRow, &nowstate)
	}
	resp[typeDiff] = diff(oldstate, nowstate)

	return resp
}
func diff(ost agspudge.Cross, nst agspudge.Cross) []string {
	res := make([]string, 0)
	if reflect.DeepEqual(&ost.Arrays, &nst.Arrays) {
		res = append(res, "В привязках изменений нет")
		return res
	}
	res = append(res, "В привязках следующие изменения")
	if ost.Arrays.TypeDevice != nst.Arrays.TypeDevice {
		res = append(res, "Изменился тип устройства")
	}
	if !reflect.DeepEqual(&ost.Arrays.SetupDK, &nst.Arrays.SetupDK) {
		res = append(res, "Изменились настройки ДК")
	}
	if !reflect.DeepEqual(&ost.Arrays.SetDK, &nst.Arrays.SetDK) {
		//res = append(res, "Изменились планы координации")
		for i, _ := range ost.Arrays.SetDK.DK {
			opk := ost.Arrays.SetDK.DK[i]
			npk := nst.Arrays.SetDK.DK[i]
			if !reflect.DeepEqual(&opk, &npk) {
				res = append(res, "Изменился ПК"+strconv.Itoa(i+1))
			}
		}
	}
	if !reflect.DeepEqual(&ost.Arrays.MonthSets, &nst.Arrays.MonthSets) {
		for i, _ := range ost.Arrays.MonthSets.MonthSets {
			oms := ost.Arrays.MonthSets.MonthSets[i]
			nms := nst.Arrays.MonthSets.MonthSets[i]
			if !reflect.DeepEqual(&oms, &nms) {
				res = append(res, "Изменился в годовом плане месяц "+strconv.Itoa(i+1))
			}
		}
	}
	if !reflect.DeepEqual(&ost.Arrays.WeekSets, &nst.Arrays.WeekSets) {
		for i, _ := range ost.Arrays.WeekSets.WeekSets {
			oms := ost.Arrays.WeekSets.WeekSets[i]
			nms := nst.Arrays.WeekSets.WeekSets[i]
			if !reflect.DeepEqual(&oms, &nms) {
				res = append(res, "Изменился недельный план номер "+strconv.Itoa(i+1))
			}
		}
	}
	if !reflect.DeepEqual(&ost.Arrays.DaySets, &nst.Arrays.DaySets) {
		for i, _ := range ost.Arrays.DaySets.DaySets {
			oms := ost.Arrays.DaySets.DaySets[i]
			nms := nst.Arrays.DaySets.DaySets[i]
			if !reflect.DeepEqual(&oms, &nms) {
				res = append(res, "Изменился суточный план номер "+strconv.Itoa(i+1))
			}
		}
	}
	if !reflect.DeepEqual(&ost.Arrays.SetCtrl, &nst.Arrays.SetCtrl) {
		for i, _ := range ost.Arrays.SetCtrl.Stage {
			oms := ost.Arrays.SetCtrl.Stage[i]
			nms := nst.Arrays.SetCtrl.Stage[i]
			if !reflect.DeepEqual(&oms, &nms) {
				res = append(res, "Изменился интервал контроля входов номер "+strconv.Itoa(i+1))
			}
		}
	}
	if !reflect.DeepEqual(&ost.Arrays.SetTimeUse, &nst.Arrays.SetTimeUse) {
		res = append(res, "Изменились настройки внешних входов ")
	}
	if !reflect.DeepEqual(&ost.Arrays.TimeDivice, &nst.Arrays.TimeDivice) {
		res = append(res, "Изменились настройки времени на устройстве ")
	}
	if !reflect.DeepEqual(&ost.Arrays.StatDefine, &nst.Arrays.StatDefine) {
		for i, _ := range ost.Arrays.StatDefine.Levels {
			oms := ost.Arrays.StatDefine.Levels[i]
			nms := nst.Arrays.StatDefine.Levels[i]
			if !reflect.DeepEqual(&oms, &nms) {
				res = append(res, "Изменилась настройка канала сбора статистики номер "+strconv.Itoa(i+1))
			}
		}
	}
	if !reflect.DeepEqual(&ost.Arrays.PointSet, &nst.Arrays.PointSet) {
		for i, _ := range ost.Arrays.PointSet.Points {
			oms := ost.Arrays.PointSet.Points[i]
			nms := nst.Arrays.PointSet.Points[i]
			if !reflect.DeepEqual(&oms, &nms) {
				res = append(res, "Изменилась настройка точки сбора статистики номер "+strconv.Itoa(i+1))
			}
		}
	}
	if !reflect.DeepEqual(&ost.Arrays.UseInput, &nst.Arrays.UseInput) {
		for i, _ := range ost.Arrays.UseInput.Used {
			oms := ost.Arrays.UseInput.Used[i]
			nms := nst.Arrays.UseInput.Used[i]
			if !reflect.DeepEqual(&oms, &nms) {
				res = append(res, "Изменилось назначение точки сбора статистики номер "+strconv.Itoa(i+1))
			}
		}
	}
	return res
}

//readPump обработчик чтения сокета
func (c *ClientControlCr) readPump() {
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { _ = c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, p, err := c.conn.ReadMessage()
		if err != nil {
			c.hub.unregister <- c
			break
		}
		//ну отправка и отправка
		typeSelect, err := sockets.ChoseTypeMessage(p)
		if err != nil {
			logger.Error.Printf("|IP: %v |Login: %v |Resource: /cross |Message: %v \n", c.crossInfo.AccInfo.IP, c.crossInfo.AccInfo.Login, err.Error())
			resp := newControlMess(typeError, nil)
			resp.Data["message"] = ErrorMessage{Error: errParseType}
			c.send <- resp
			continue
		}
		switch typeSelect {
		case typeGetHistory: //Отправка state для данной истории
			{
				gh := GetHistory{}
				_ = json.Unmarshal(p, &gh)
				resp := newControlMess(typeSendHistory, nil)
				resp.Data = c.getdHistoryData(gh.Data.Time)
				c.send <- resp
			}
		case typeSendB: //отправка state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := newControlMess(typeSendB, nil)
				resp.Data = sendCrossData(temp.State, c.crossInfo.Idevice, c.crossInfo.Pos, c.crossInfo.AccInfo.Login)
				if len(resp.Data) > 0 {
					c.send <- resp
				} else {
					if temp.RePaint {
						resp := newControlMess(typeRepaintCheck, nil)
						if crossCreator.ShortCreateDirPng(temp.State.Region, temp.State.Area, temp.State.ID, temp.Z, temp.State.Dgis) {
							resp.Data["message"] = "позиция изменена"
							resp.Data["status"] = true
						} else {
							resp.Data["message"] = "при изменении позиции произошла ошибка - свяжитесь с Администраторов"
							resp.Data["status"] = false
						}
						c.send <- resp
					}
				}
			}
		case typeCheckB: //проверка state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := checkCrossData(temp.State)
				c.send <- resp
			}
		case typeCreateB: //создание перекрестка
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				resp := newControlMess(typeCreateB, nil)
				resp.Data = createCrossData(temp.State, c.crossInfo.Pos, c.crossInfo.AccInfo.Login, temp.Z)
				c.send <- resp
			}
		case typeDeleteB: //удаление state
			{
				temp := StateHandler{}
				_ = json.Unmarshal(p, &temp)
				userCross := agspudge.UserCross{User: c.crossInfo.AccInfo.Login, State: temp.State}
				userCross.State.IDevice = -1
				mess := tcpConnect.TCPMessage{
					User:        c.crossInfo.AccInfo.Login,
					TCPType:     tcpConnect.TypeState,
					Idevice:     temp.State.IDevice,
					Data:        userCross,
					From:        tcpConnect.FromCrControlSoc,
					CommandType: typeDeleteB,
					Pos:         c.crossInfo.Pos,
				}
				mess.SendToTCPServer()
			}
		case typeUpdateB: //обновление state
			{
				resp := newControlMess(typeUpdateB, nil)
				resp, _, _ = takeControlInfo(c.crossInfo.Pos)
				resp.Data["edit"] = c.crossInfo.Edit
				c.send <- resp
			}
		case typeEditInfoB: //информация о пользователях что редактируют перекресток
			{
				resp := newControlMess(typeEditInfoB, nil)

				type usersEdit struct {
					User string `json:"user"`
					Edit bool   `json:"edit"`
				}
				var users []usersEdit
				//нужно ха этим посмотреть но проблем не должно быть
				for client := range c.hub.clients {
					if client.crossInfo.Pos == c.crossInfo.Pos {
						temp := usersEdit{User: client.crossInfo.AccInfo.Login, Edit: client.crossInfo.Edit}
						users = append(users, temp)
					}
				}
				resp.Data["users"] = users
				c.send <- resp
			}
		case typeDButton: //отправка сообщения о изменениии режима работы
			{
				arm := comm.CommandARM{}
				_ = json.Unmarshal(p, &arm)
				arm.User = c.crossInfo.AccInfo.Login
				var mess = tcpConnect.TCPMessage{
					User:        c.crossInfo.AccInfo.Login,
					TCPType:     tcpConnect.TypeDispatch,
					Idevice:     arm.ID,
					Data:        arm,
					From:        tcpConnect.FromCrControlSoc,
					CommandType: typeDButton,
					Pos:         c.crossInfo.Pos,
				}
				mess.SendToTCPServer()
			}
		default:
			{
				resp := newControlMess("type", nil)
				resp.Data["type"] = typeSelect
				c.send <- resp
			}
		}
	}
}

//writePump обработчик записи в сокет
func (c *ClientControlCr) writePump() {
	pingTick := time.NewTicker(pingPeriod)
	defer func() {
		pingTick.Stop()
	}()
	for {
		select {
		case mess, ok := <-c.send:
			{
				_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "канал был закрыт"))
					return
				}

				_ = c.conn.WriteJSON(mess)
				// Add queued chat messages to the current websocket message.
				n := len(c.send)
				for i := 0; i < n; i++ {
					_ = c.conn.WriteJSON(<-c.send)
				}
			}
		case <-pingTick.C:
			{
				_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}
}
