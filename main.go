package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/JanFant/TLServer/data"
	"github.com/JanFant/TLServer/license"
	"github.com/JanFant/TLServer/logger"
	"github.com/JanFant/TLServer/routes"
	"github.com/JanFant/TLServer/tcpConnect"
)

var err error

func init() {
	//Начало работы, читаем настроечный фаил
	data.GlobalConfig = data.NewConfig()
	if _, err := toml.DecodeFile("config.toml", &data.GlobalConfig); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
	}

}

func main() {
	//Загружаем модуль логирования
	if err = logger.Init(data.GlobalConfig.LoggerPath); err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}

	//Запуск если есть файл с токеном лицензии license.key
	license.LicenseCheck()

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
	tcpConnect.TCPClientStart(data.GlobalConfig.TCPConfig)
	//----------------------------------------------------------------------

	//запуск сервера
	routes.StartServer()

	logger.Info.Println("|Message: Exit working...")
	fmt.Println("Exit working...")
}
