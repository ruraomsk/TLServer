package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/ruraomsk/TLServer/internal/app/apiserver"
	"github.com/ruraomsk/TLServer/internal/app/tcpConnect"
	"github.com/ruraomsk/TLServer/internal/model/config"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"github.com/ruraomsk/TLServer/internal/model/license"
	"github.com/ruraomsk/TLServer/internal/sockets/techArm"
	"github.com/ruraomsk/TLServer/logger"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	techArm.GPRSInfo.Send = config.GlobalConfig.TCPConfig.SendGPRS
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
	err := data.ConnectDB()
	if err != nil {
		logger.Error.Println("|Message: Error open DB", err.Error())
		fmt.Println("Error open DB", err.Error())
		return
	}

	logger.Info.Println("|Message: Start work...")
	fmt.Println("Start work...")

	//раз в час обновляем данные регионов, и состояний
	go data.CacheDataUpdate()
	tcpConnect.TCPClientStart(config.GlobalConfig.TCPConfig)
	////----------------------------------------------------------------------
	//
	//запуск сервера
	srvHttp, srvHttps := apiserver.MainServer(apiserver.ServerConfig)
	go func() {
		//Сервер на порте 80 - для переадресации
		go func() {
			if err := srvHttp.ListenAndServe(); err != nil {
				logger.Error.Println("|Message: Error start main server ", err.Error())
				fmt.Println("Error start main server ", err.Error())
			}
		}()
		//Запуск сервера на порте 443 - с проверкой ключа в каталоге ssl
		if _, err := ioutil.ReadFile("ssl/cert.crt"); err == nil {
			fmt.Println("Start server with SSL")
			logger.Info.Println("|Message: Start server with SSL")
			if err := srvHttps.ListenAndServeTLS("ssl/cert.crt", "ssl/cert.key"); err != nil && err != http.ErrServerClosed {
				logger.Error.Println("|Message: Error start main server ", err.Error())
				fmt.Println("Error start main server ", err.Error())
			}
		} else {
			fmt.Println("Start server without SSL")
			logger.Info.Println("|Message: Start server without SSL")
			if err := srvHttps.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error.Println("|Message: Error start main server ", err.Error())
				fmt.Println("Error start main server ", err.Error())
			}
		}
	}()
	srv2 := apiserver.ExchangeServer(apiserver.ServerConfig)
	go func() {
		if err := srv2.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error.Println("|Message: Error start exchange server ", err.Error())
			fmt.Println("Error start exchange server ", err.Error())
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := srvHttp.Shutdown(ctx); err != nil {
		logger.Info.Println("|Message: Server (http) forced shutdown...", err)
		fmt.Println("Server forced shutdown...", err)
	}
	if err := srvHttps.Shutdown(ctx); err != nil {
		logger.Info.Println("|Message: Server (https) forced shutdown...", err)
		fmt.Println("Server forced shutdown...", err)
	}

	if err := srv2.Shutdown(ctx); err != nil {
		logger.Info.Println("|Message: Server (exchange) forced shutdown...", err)
		fmt.Println("Server forced shutdown...", err)
	}

	logger.Info.Println("|Message: Shutting down server...")
	fmt.Println("Shutting down server...")
}
