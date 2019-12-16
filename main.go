package main

import (
	"fmt"
	"os"

	"github.com/gorilla/handlers"

	"./data"
	"./logger"
	"./routAuth"
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

	//раз в час обновляем данные регионов, и состояний
	go data.CacheDataUpdate()
	//----------------------------------------------------------------------

	// Создаем новый ServeMux для HTTPS соединений
	router := mux.NewRouter()

	//основной обработчик
	//начальная страница
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./views/screen.html")
	})
	//пусть к файлам скриптов и т.д.
	router.PathPrefix("/static/").Handler(http.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("//Fileserver/общая папка/TEMP рабочий/Semyon/lib/js"))))).Methods("GET")
	router.PathPrefix("/img/").Handler(http.Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("./views/img"))))).Methods("GET")
	//запрос на вход в систему
	router.HandleFunc("/login", whandlers.LoginAcc).Methods("POST")
	router.HandleFunc("/test", whandlers.TestHello).Methods("POST")
	router.HandleFunc("/create", whandlers.CreateAcc).Methods("POST")

	subRout := router.PathPrefix("/").Subrouter()
	subRout.Use(routAuth.JwtAuth)
	//запрос на создание пользователя
	subRout.HandleFunc("/{slug}/create", whandlers.CreateAcc).Methods("POST")
	//запрос странички с картой
	subRout.HandleFunc("/{slug}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./views/workplace.html")
	}).Methods("GET")
	//запрос информации для заполнения странички с картой
	subRout.HandleFunc("/{slug}", whandlers.BuildMapPage).Methods("POST")
	subRout.HandleFunc("/{slug}/update", whandlers.UpdateMapPage).Methods("POST")
	//тест
	subRout.HandleFunc("/{slug}/testtoken", whandlers.TestToken).Methods("POST")

	// Запуск HTTP сервера
	// if err = http.ListenAndServe(os.Getenv("server_ip"), handlers.LoggingHandler(os.Stdout, router)); err != nil {
	if err = http.ListenAndServeTLS(os.Getenv("server_ip"), "domain.crt", "domain.key", handlers.LoggingHandler(os.Stdout, router)); err != nil {
		logger.Info.Println("Server can't started ", err.Error())
		fmt.Println("Server can't started ", err.Error())
	}

	logger.Info.Println("Exit working...")
	fmt.Println("Exit working...")
}
