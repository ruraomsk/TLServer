package sockets

//DispatchMessageFromTechArm канал для оправки сообщений с тех арма
var DispatchMessageFromAnotherPlace chan DBMessage

//DBMessage структура для отправки информации о переключение режимов с тех арма
type DBMessage struct {
	Idevice int         //id устройства
	Data    interface{} //информация о выполнении команды
}
