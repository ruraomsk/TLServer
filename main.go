package main

import (
	"fmt"
	"os"

	"github.com/gorilla/handlers"

	"./data"
	"./logger"
	"./whandlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"net/http"
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
		logger.Info.Println("Error open DB", err.Error())
		fmt.Println("Error open DB", err.Error())
		return
	}
	defer data.GetDB().Close() // не забывает закрыть подключение

	logger.Info.Println("Start work...")
	fmt.Println("Start work...")
	//----------------------------------------------------------------------

	// Создаем новый ServeMux для HTTPS соединений
	router := mux.NewRouter()

	//основной обработчик
	//корневая страница (загружаемый файл)
	// router.Handle("/", http.FileServer(http.Dir("./views/")))
	//asdasd
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./views/screen.html")
	})
	//страница с ресурсами картинки, подложка и тд...
	router.PathPrefix("/static/").Handler(http.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))))
	//тестовая страница приветствия

	router.HandleFunc("/hello", whandlers.TestHello).Methods("GET")
	router.HandleFunc("/login", whandlers.LoginAcc).Methods("POST")
	router.HandleFunc("/create", whandlers.CreateAcc).Methods("POST")
	router.HandleFunc("/test", whandlers.TestHello).Methods("GET")
	router.HandleFunc("/testtoken", whandlers.TestToken).Methods("GET")
	//router.Use(routAuth.JwtAuth)
	// Запуск HTTP сервера
	if err = http.ListenAndServe(os.Getenv("server_ip"), handlers.LoggingHandler(os.Stdout, router)); err != nil {
		logger.Info.Println("Server can't started ", err.Error())
		fmt.Println("Server can't started ", err.Error())
	}

	logger.Info.Println("Exit working...")
	fmt.Println("Exit working...")
}
