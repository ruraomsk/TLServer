package routes

import (
	"../logger"
	"../whandlers"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

var err error

//StartServer запуск сервера
func StartServer() {
	resourcePath := os.Getenv("resourcePath")
	// Создаем новый ServeMux для HTTPS соединений
	router := mux.NewRouter()
	//основной обработчик
	//начальная страница
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, resourcePath+"/screen.html")
	})
	//путь к скриптам они открыты
	router.PathPrefix("/static/").Handler(http.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(resourcePath))))).Methods("GET")

	//запрос на вход в систему
	router.HandleFunc("/login", whandlers.LoginAcc).Methods("POST")
	router.HandleFunc("/test", whandlers.TestHello).Methods("POST")

	//------------------------------------------------------------------------------------------------------------------
	//обязательный общий путь
	subRout := router.PathPrefix("/user").Subrouter()
	//добавление к роутеру контроля токена
	subRout.Use(JwtAuth)

	//работа с основной страничкой карты
	subRout.HandleFunc("/{slug}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, resourcePath+"/workplace.html")
	}).Methods("GET")
	//запрос информации для заполнения странички с картой
	subRout.HandleFunc("/{slug}", whandlers.BuildMapPage).Methods("POST")
	//обработчик выхода из системы
	subRout.HandleFunc("/{slug}/logOut", whandlers.LoginAccOut).Methods("GET")
	//обновление странички с данными которые попали в область пользователя
	subRout.HandleFunc("/{slug}/update", whandlers.UpdateMapPage).Methods("POST")
	//обработчик для формирования новых координат отображения карты
	subRout.HandleFunc("/{slug}/locationButton", whandlers.LocationButtonMapPage).Methods("POST")

	//работа со странички перекрестков (страничка)
	subRout.HandleFunc("/{slug}/cross", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, resourcePath+"/cross.html")
	}).Methods("GET")
	//отправка информации с состояниями перекреста видна всем основная информация
	subRout.HandleFunc("/{slug}/cross", whandlers.BuildCross).Methods("POST")
	//обработчик диспетчерского управления
	subRout.HandleFunc("/{slug}/cross/DispatchControlButtons", whandlers.DispatchControlButtons).Methods("POST")

	//расширеная страничка настройки перекрестка (страничка)
	subRout.HandleFunc("/{slug}/cross/control", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, resourcePath+"/crossControl.html")
	}).Methods("GET")
	//данные по расширенной странички перекрестков
	subRout.HandleFunc("/{slug}/cross/control", whandlers.ControlCross).Methods("POST")
	//обработчик приема данных от пользователя для отправки на устройство
	subRout.HandleFunc("/{slug}/cross/control/sendButton", whandlers.ControlSendButton).Methods("POST")
	//обработчик проверки данных
	subRout.HandleFunc("/{slug}/cross/control/checkButton", whandlers.ControlCheckButton).Methods("POST")
	//обработчик обработчик удаления перекрсетка
	subRout.HandleFunc("/{slug}/cross/control/deleteButton", whandlers.ControlDeleteButton).Methods("POST")

	//обработка создание и редактирования пользователя (страничка)
	subRout.HandleFunc("/{slug}/manage", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, resourcePath+"/manage.html")
	}).Methods("GET")
	//обработчик запроса лог файлов
	subRout.HandleFunc("/{slug}/manage", whandlers.DisplayAccInfo).Methods("POST")
	//обработчик для изменения пароля
	subRout.HandleFunc("/{slug}/manage/changepw", whandlers.ActChangePw).Methods("POST")
	//обработчик для удаления аккаунтов
	subRout.HandleFunc("/{slug}/manage/delete", whandlers.ActDeleteAccount).Methods("POST")
	//обработчик для добавления аккаунтов
	subRout.HandleFunc("/{slug}/manage/add", whandlers.ActAddAccount).Methods("POST")
	//обработчик для редактирования данных аккаунта
	subRout.HandleFunc("/{slug}/manage/update", whandlers.ActUpdateAccount).Methods("POST")

	//обработка лог файлов (страничка)
	subRout.HandleFunc("/{slug}/manage/log", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, resourcePath+"/log.html")
	}).Methods("GET")
	//обработчик по выгрузке лог файлов
	subRout.HandleFunc("/{slug}/manage/log", whandlers.DisplayLogFile).Methods("POST")
	//обработчик выбранного лог файла
	subRout.HandleFunc("/{slug}/manage/log/info", whandlers.DisplayLogInfo).Methods("GET")

	//обработка создания каталога карты перекрестков (страничка)
	subRout.HandleFunc("/{slug}/manage/crossCreator", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, resourcePath+"/crossCreator.html")
	}).Methods("GET")
	//
	subRout.HandleFunc("/{slug}/manage/crossCreator", whandlers.MainCrossCreator).Methods("POST")
	subRout.HandleFunc("/{slug}/manage/crossCreator/checkAllCross", whandlers.CheckAllCross).Methods("POST")
	subRout.HandleFunc("/{slug}/manage/crossCreator/checkSelected", whandlers.CheckSelectedDirCross).Methods("POST")
	subRout.HandleFunc("/{slug}/manage/crossCreator/makeSelected", whandlers.MakeSelectedDirCross).Methods("POST")

	//тест просто тест!
	subRout.HandleFunc("/{slug}/testtoken", whandlers.TestToken).Methods("POST")

	//------------------------------------------------------------------------------------------------------------------
	//роутер для фаил сервера, он закрыт токеном, скачивать могут только авторизированные пользователи
	fileRout := router.PathPrefix("/file").Subrouter()
	//добавление к роутеру контроля токена
	fileRout.Use(JwtFile)
	//описание пути и скриптов для получения файлов для перекрестка
	fileRout.PathPrefix("/cross/").Handler(http.Handler(http.StripPrefix("/file/cross/", http.FileServer(http.Dir("./views/cross"))))).Methods("GET")
	//описание пути и скриптов для получения файлов для основных картинок
	fileRout.PathPrefix("/img/").Handler(http.Handler(http.StripPrefix("/file/img/", http.FileServer(http.Dir("./views/img"))))).Methods("GET")
	//описание пути и скриптов для получения файлов для иконок
	fileRout.PathPrefix("/icons/").Handler(http.Handler(http.StripPrefix("/file/icons/", http.FileServer(http.Dir("./views/icons"))))).Methods("GET")

	//------------------------------------------------------------------------------------------------------------------
	// Запуск HTTP сервера
	// if err = http.ListenAndServe(os.Getenv("server_ip"), handlers.LoggingHandler(os.Stdout, router)); err != nil {
	if err = http.ListenAndServeTLS(os.Getenv("server_ip"), "domain.crt", "domain.key", handlers.LoggingHandler(os.Stdout, router)); err != nil {
		logger.Error.Println("|Message: Server can't started: ", err.Error())
		fmt.Println("Server can't started ", err.Error())
	}
}
