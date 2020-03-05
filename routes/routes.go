package routes

import (
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/JanFant/TLServer/whandlers"
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
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { //начальная страница
		http.ServeFile(w, r, resourcePath+"/screen.html")
	})
	router.PathPrefix("/static/").Handler(http.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(resourcePath))))).Methods("GET") //путь к скриптам они открыты
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //заглушка странички 404
		http.ServeFile(w, r, resourcePath+"/notFound.html")
	})
	router.HandleFunc("/login", whandlers.LoginAcc).Methods("POST") //запрос на вход в систему
	router.HandleFunc("/test", whandlers.TestHello).Methods("POST")
	//------------------------------------------------------------------------------------------------------------------
	//обязательный общий путь
	subRout := router.PathPrefix("/user").Subrouter()
	subRout.Use(JwtAuth) //добавление к роутеру контроль токена
	subRout.HandleFunc("/{slug}", func(w http.ResponseWriter, r *http.Request) { //работа с основной страничкой карты
		http.ServeFile(w, r, resourcePath+"/workplace.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}", whandlers.BuildMapPage).Methods("POST")                         //запрос информации для заполнения странички с картой
	subRout.HandleFunc("/{slug}/logOut", whandlers.LoginAccOut).Methods("GET")                    //обработчик выхода из системы
	subRout.HandleFunc("/{slug}/update", whandlers.UpdateMapPage).Methods("POST")                 //обновление странички с данными которые попали в область пользователя
	subRout.HandleFunc("/{slug}/locationButton", whandlers.LocationButtonMapPage).Methods("POST") //обработчик для формирования новых координат отображения карты
	subRout.HandleFunc("/{slug}/cross", func(w http.ResponseWriter, r *http.Request) { //работа со странички перекрестков (страничка)
		http.ServeFile(w, r, resourcePath+"/cross.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/cross", whandlers.BuildCross).Methods("POST")                                    //отправка информации с состояниями перекреста видна всем основная информация
	subRout.HandleFunc("/{slug}/cross/DispatchControlButtons", whandlers.DispatchControlButtons).Methods("POST") //обработчик диспетчерского управления
	subRout.HandleFunc("/{slug}/cross/control", func(w http.ResponseWriter, r *http.Request) { //расширеная страничка настройки перекрестка (страничка)
		http.ServeFile(w, r, resourcePath+"/crossControl.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/cross/control", whandlers.ControlCross).Methods("POST")                     //данные по расширенной странички перекрестков
	subRout.HandleFunc("/{slug}/cross/control/close", whandlers.ControlCloseCross).Methods("GET")           //обработчик закрытия перекрестка
	subRout.HandleFunc("/{slug}/cross/control/editable", whandlers.ControlEditableCross).Methods("GET")     //обработчик закрытия перекрестка
	subRout.HandleFunc("/{slug}/cross/control/sendButton", whandlers.ControlSendButton).Methods("POST")     //обработчик приема данных от пользователя для отправки на устройство
	subRout.HandleFunc("/{slug}/cross/control/checkButton", whandlers.ControlCheckButton).Methods("POST")   //обработчик проверки данных
	subRout.HandleFunc("/{slug}/cross/control/createButton", whandlers.ControlCreateButton).Methods("POST") //обработчик создания перекрестка
	subRout.HandleFunc("/{slug}/cross/control/deleteButton", whandlers.ControlDeleteButton).Methods("POST") //обработчик обработчик удаления перекрсетка

	subRout.HandleFunc("/{slug}/manage", func(w http.ResponseWriter, r *http.Request) { //обработка создание и редактирования пользователя (страничка)
		http.ServeFile(w, r, resourcePath+"/manage.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/manage", whandlers.DisplayAccInfo).Methods("POST")          //обработчик запроса лог файлов
	subRout.HandleFunc("/{slug}/manage/changepw", whandlers.ActChangePw).Methods("POST")    //обработчик для изменения пароля
	subRout.HandleFunc("/{slug}/manage/delete", whandlers.ActDeleteAccount).Methods("POST") //обработчик для удаления аккаунтов
	subRout.HandleFunc("/{slug}/manage/add", whandlers.ActAddAccount).Methods("POST")       //обработчик для добавления аккаунтов
	subRout.HandleFunc("/{slug}/manage/update", whandlers.ActUpdateAccount).Methods("POST") //обработчик для редактирования данных аккаунта
	subRout.HandleFunc("/{slug}/manage/crossEditControl", func(w http.ResponseWriter, r *http.Request) { //обработчик по управлению занятых перекрестков (страничка)
		http.ServeFile(w, r, resourcePath+"/crossEditControl.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/manage/crossEditControl", whandlers.CrossEditInfo).Methods("POST")      //обработчик по управлению занятых перекрестков
	subRout.HandleFunc("/{slug}/manage/crossEditControl/free", whandlers.CrossEditFree).Methods("POST") //обработчик по управлению освобождению перекрестка

	subRout.HandleFunc("/{slug}/manage/stateTest", func(w http.ResponseWriter, r *http.Request) { //обработчик проверки всего State (страничка)
		http.ServeFile(w, r, resourcePath+"/stateTest.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/manage/stateTest", whandlers.ControlTestState).Methods("POST") //обработчик проверки всего State

	subRout.HandleFunc("/{slug}/manage/log", func(w http.ResponseWriter, r *http.Request) { //обработка лог файлов (страничка)
		http.ServeFile(w, r, resourcePath+"/log.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/manage/log", whandlers.DisplayLogFile).Methods("POST")     //обработчик по выгрузке лог файлов
	subRout.HandleFunc("/{slug}/manage/log/info", whandlers.DisplayLogInfo).Methods("GET") //обработчик выбранного лог файла
	subRout.HandleFunc("/{slug}/manage/crossCreator", func(w http.ResponseWriter, r *http.Request) { //обработка создания каталога карты перекрестков (страничка)
		http.ServeFile(w, r, resourcePath+"/crossCreator.html")
	}).Methods("GET")

	subRout.HandleFunc("/{slug}/manage/crossCreator", whandlers.MainCrossCreator).Methods("POST")
	subRout.HandleFunc("/{slug}/manage/crossCreator/checkAllCross", whandlers.CheckAllCross).Methods("POST")
	subRout.HandleFunc("/{slug}/manage/crossCreator/checkSelected", whandlers.CheckSelectedDirCross).Methods("POST")
	subRout.HandleFunc("/{slug}/manage/crossCreator/makeSelected", whandlers.MakeSelectedDirCross).Methods("POST")

	//тест просто тест!
	subRout.HandleFunc("/{slug}/testtoken", whandlers.TestToken).Methods("POST")
	//------------------------------------------------------------------------------------------------------------------
	//роутер для фаил сервера, он закрыт токеном, скачивать могут только авторизированные пользователи
	fileRout := router.PathPrefix("/file").Subrouter()
	fileRout.Use(JwtFile)                                                                                                                             //добавление к роутеру контроля токена
	fileRout.PathPrefix("/cross/").Handler(http.Handler(http.StripPrefix("/file/cross/", http.FileServer(http.Dir("./views/cross"))))).Methods("GET") //описание пути и скриптов для получения файлов для перекрестка
	fileRout.PathPrefix("/img/").Handler(http.Handler(http.StripPrefix("/file/img/", http.FileServer(http.Dir("./views/img"))))).Methods("GET")       //описание пути и скриптов для получения файлов для основных картинок
	fileRout.PathPrefix("/icons/").Handler(http.Handler(http.StripPrefix("/file/icons/", http.FileServer(http.Dir("./views/icons"))))).Methods("GET") //описание пути и скриптов для получения файлов для иконок

	//------------------------------------------------------------------------------------------------------------------
	// Запуск HTTP сервера
	if err = http.ListenAndServeTLS(os.Getenv("server_ip1"), "domain.crt", "domain.key", handlers.LoggingHandler(os.Stdout, router)); err != nil {
		logger.Error.Println("|Message: Server can't started: ", err.Error())
		fmt.Println("Server can't started ", err.Error())
	}
}
