package mainCross

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"time"
)

//HubCross структура хаба для cross
type HubCross struct {
	clients    map[*ClientCross]bool
	broadcast  chan crossResponse
	register   chan *ClientCross
	unregister chan *ClientCross
}

//NewCrossHub создание хаба
func NewCrossHub() *HubCross {
	return &HubCross{
		broadcast:  make(chan crossResponse),
		clients:    make(map[*ClientCross]bool),
		register:   make(chan *ClientCross),
		unregister: make(chan *ClientCross),
	}
}

//Run запуск хаба для xctrl
func (h *HubCross) Run(db *sqlx.DB) {

	updateTicker := time.NewTicker(time.Second * 20)
	defer updateTicker.Stop()

	for {
		select {
		case <-updateTicker.C:
			{

			}
		case client := <-h.register:
			{
				//проверка не существование такого перекрестка (сбос если нету)
				_, err := crossSock.GetNewState(client.crossInfo.Pos, db)
				if err != nil {
					close(client.send)
					_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errCrossDoesntExist))
					_ = client.conn.Close()
					client.regStatus <- regStatus{ok: false}
					continue
				}

				//проверка открыт ли у этого пользователя такой перекресток
				for hubClient := range h.clients {
					if client.crossInfo.Pos == hubClient.crossInfo.Pos && client.crossInfo.Login == hubClient.crossInfo.Login {
						close(client.send)
						_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, errDoubleOpeningDevice))
						_ = client.conn.Close()
						client.regStatus <- regStatus{ok: false}
						continue
					}
				}
				regStatus := regStatus{ok: true, edit: false}
				//флаг редактирования перекрестка
				//если роль пришедшего Viewer то влаг ему не ставим
				flagEdit := false
				if client.crossInfo.Role != "Viewer" {
					for hClient := range h.clients {
						if hClient.crossInfo.Pos == client.crossInfo.Pos && hClient.crossInfo.Edit {
							flagEdit = true
							break
						}
					}
					if !flagEdit {
						regStatus.edit = true
					}
				}

				//кромешный пи**** с созданием нормального клиента
				resp, Idevice := takeCrossInfo(client.crossInfo.Pos, db)
				regStatus.idevice = Idevice
				resp.Data["edit"] = regStatus.edit
				resp.Data["controlCrossFlag"] = false
				controlCrossFlag, _ := data.AccessCheck(client.crossInfo.Login, client.crossInfo.Role, 4)
				if (fmt.Sprint(resp.Data["region"]) == client.crossInfo.region) || (client.crossInfo.region == "*") {
					resp.Data["controlCrossFlag"] = controlCrossFlag
				}
				delete(resp.Data, "region")

				client.regStatus <- regStatus
				client.crossInfo.Idevice = Idevice
				client.crossInfo.Edit = flagEdit
				h.clients[client] = true
				client.send <- resp
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
		}
	}
}
