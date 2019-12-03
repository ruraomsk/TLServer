package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"os"

	"net/http"

	"./logger"
	"./whandlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var err error

func main() {
	//Начало работы, читаем настроечный фаил
	if err = godotenv.Load(); err != nil {
		fmt.Println("Can't load enc file - ", err.Error())
		return
	}

	//Загружаем модуль логирования
	if err = logger.Init(os.Getenv("logger_path")); err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
	logger.Info.Println("Start work...")
	fmt.Println("Start work...")

	//----------------------------------------------------------------------

	// Создаем новый ServeMux для HTTPS соединений
	router := mux.NewRouter()

	//основной обработчик
	//корневая страница (загружаемый файл)
	router.Handle("/", http.FileServer(http.Dir("./views/")))
	//страница с ресурсами картинки, подложка и тд...
	router.PathPrefix("/static/").Handler(http.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))))

	//тестовая страница приветствия
	router.HandleFunc("/hello", whandlers.TestHello).Methods("GET")


	// Запуск HTTP сервера
	if err = http.ListenAndServe(os.Getenv("server_ip"), handlers.LoggingHandler(os.Stdout,router)); err != nil {
		logger.Info.Println("Server can't started ", err.Error())
		fmt.Println("Server can't started ", err.Error())
	}

	logger.Info.Println("Exit working...")
	fmt.Println("Exit working...")
}
