package main

import (
	"fmt"
	"github.com/JanFant/TLServer/data"
	"github.com/JanFant/TLServer/logger"
	"github.com/JanFant/TLServer/routes"
	"github.com/JanFant/TLServer/tcpConnect"
	"github.com/joho/godotenv"
	"os"
)

var err error

func init() {
	//Начало работы, читаем настроечный фаил
	if err = godotenv.Load(); err != nil {
		fmt.Println("Can't load enc file - ", err.Error())

	}
}

func main() {
	//Загружаем модуль логирования
	if err = logger.Init(os.Getenv("logger_path")); err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}

	//Подключение к базе данных
	if err = data.ConnectDB(); err != nil {
		logger.Error.Println("|Message: Error open DB", err.Error())
		fmt.Println("Error open DB", err.Error())
		return
	}
	defer data.GetDB().Close() // не забывает закрыть подключение

	logger.Info.Println("|Message: Start work...")
	fmt.Println("Start work...")

	//раз в час обновляем данные регионов, и состояний
	go data.CacheDataUpdate()
	tcpConnect.TCPClientStart()
	//----------------------------------------------------------------------

	//запуск сервера
	routes.StartServer()

	logger.Info.Println("|Message: Exit working...")
	fmt.Println("Exit working...")
}
