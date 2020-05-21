package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/JanFant/newTLServer/internal/app/apiserver"
	"github.com/JanFant/newTLServer/internal/app/config"
	"github.com/JanFant/newTLServer/internal/app/db"
	"github.com/JanFant/newTLServer/internal/model/cache"
	"github.com/JanFant/newTLServer/internal/model/logger"
	"os"
)

var err error

func init() {
	var configPath string
	//Начало работы, загружаем настроечный файл
	flag.StringVar(&configPath, "config-path", "configs/config.toml", "path to config file")

	//Начало работы, читаем настроечный фаил
	config.GlobalConfig = config.NewConfig()
	if _, err := toml.DecodeFile(configPath, &config.GlobalConfig); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
		os.Exit(1)
	}

}

func main() {
	//Загружаем модуль логирования
	if err = logger.Init(config.GlobalConfig.LoggerPath); err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}

	////Запуск если есть файл с токеном лицензии license.key
	//license.LicenseCheck()
	//

	////Подключение к базе данных
	dbConn, err := db.ConnectDB()
	if err != nil {
		logger.Error.Println("|Message: Error open DB", err.Error())
		fmt.Println("Error open DB", err.Error())
		return
	}
	defer dbConn.Close() // не забывает закрыть подключение

	logger.Info.Println("|Message: Start work...")
	fmt.Println("Start work...")

	//раз в час обновляем данные регионов, и состояний
	go cache.CacheDataUpdate()
	//tcpConnect.TCPClientStart(data.GlobalConfig.TCPConfig)
	////----------------------------------------------------------------------
	//
	//запуск сервера
	var serverConf = apiserver.NewConfig()
	apiserver.StartServer(serverConf)

	logger.Info.Println("|Message: Exit working...")
	fmt.Println("Exit working...")
}
