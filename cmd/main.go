package main

import (
	"flag"
	"fmt"
	"github.com/JanFant/TLServer/internal/sockets/techArm"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/JanFant/TLServer/internal/app/apiserver"
	"github.com/JanFant/TLServer/internal/app/tcpConnect"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/logger"
)

var err error

func init() {
	var configPath string
	//Начало работы, загружаем настроечный файл
	flag.StringVar(&configPath, "config-path", "configs/config.toml", "path to config file")

	//Начало работы, читаем настроечный фаил
	config.GlobalConfig = config.NewConfig()
	apiserver.ServerConfig = apiserver.NewConfig()
	logger.LogGlobalConf = logger.NewConfig()
	if _, err := toml.DecodeFile(configPath, &apiserver.ServerConfig); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
		os.Exit(1)
	}
	if _, err := toml.DecodeFile(configPath, &config.GlobalConfig); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
		os.Exit(1)
	}
	techArm.GPRSInfo.IP = config.GlobalConfig.TCPConfig.ServerAddr
	techArm.GPRSInfo.Port = config.GlobalConfig.TCPConfig.PortGPRS
	if _, err := toml.DecodeFile(configPath, &logger.LogGlobalConf); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
		os.Exit(1)
	}
}

func main() {
	//Загружаем модуль логирования
	if err = logger.Init(logger.LogGlobalConf.LogPath); err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}

	////Запуск если есть файл с токеном лицензии license.key
	license.LicenseCheck()

	////Подключение к базе данных
	dbConn, err := data.ConnectDB()
	if err != nil {
		logger.Error.Println("|Message: Error open DB", err.Error())
		fmt.Println("Error open DB", err.Error())
		return
	}
	defer dbConn.Close() // не забывает закрыть подключение

	logger.Info.Println("|Message: Start work...")
	fmt.Println("Start work...")

	//раз в час обновляем данные регионов, и состояний
	go data.CacheDataUpdate()
	tcpConnect.TCPClientStart(config.GlobalConfig.TCPConfig)
	////----------------------------------------------------------------------
	//
	//запуск сервера
	apiserver.StartServer(apiserver.ServerConfig, dbConn)

	logger.Info.Println("|Message: Exit working...")
	fmt.Println("Exit working...")
}
