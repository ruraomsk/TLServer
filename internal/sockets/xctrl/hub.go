package xctrl

type XctrlHub struct {
	clients    map[*CPClient]bool
	broadcast  chan CPMess
	register   chan *CPClient
	unregister chan *CPClient
}

func NewXctrlHub() *XctrlHub {
	return &XctrlHub{
		broadcast:  make(chan CPMess),
		clients:    make(map[*CPClient]bool),
		register:   make(chan *CPClient),
		unregister: make(chan *CPClient),
	}
}

func (h *XctrlHub) Run() {
	for {
		select {
		case client := <-h.register:
			{
				h.clients[client] = true
			}
		case client := <-h.unregister:
			{
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			}
		case mess := <-h.broadcast:
			{
				for client := range h.clients {
					select {
					case client.send <- mess:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
