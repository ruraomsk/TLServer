package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/JanFant/TLServer/data"
	"github.com/JanFant/TLServer/logger"
	"github.com/JanFant/TLServer/whandlers"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var err error

//StartServer запуск сервера
func StartServer() {
	resourcePath := data.GlobalConfig.ResourcePath
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

	//------------------------------------------------------------------------------------------------------------------
	//обязательный общий путь
	subRout := router.PathPrefix("/user").Subrouter()
	subRout.Use(JwtAuth) //добавление к роутеру контроль токена

	subRout.Use(AccessControl) //проверка разрешения на доступ к ресурсу

	subRout.HandleFunc("/{slug}/map", func(w http.ResponseWriter, r *http.Request) { //работа с основной страничкой карты
		http.ServeFile(w, r, resourcePath+"/map.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/map", whandlers.BuildMapPage).Methods("POST")                         //запрос информации для заполнения странички с картой
	subRout.HandleFunc("/{slug}/map/logOut", whandlers.LoginAccOut).Methods("GET")                    //обработчик выхода из системы
	subRout.HandleFunc("/{slug}/map/update", whandlers.UpdateMapPage).Methods("POST")                 //обновление странички с данными которые попали в область пользователя
	subRout.HandleFunc("/{slug}/map/locationButton", whandlers.LocationButtonMapPage).Methods("POST") //обработчик для формирования новых координат отображения карты

	subRout.HandleFunc("/{slug}/cross", func(w http.ResponseWriter, r *http.Request) { //работа со странички перекрестков (страничка)
		http.ServeFile(w, r, resourcePath+"/cross.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/cross", whandlers.BuildCross).Methods("POST")       //информация о состоянии перекрёстка
	subRout.HandleFunc("/{slug}/cross/dev", whandlers.DevCrossInfo).Methods("POST") //информация о состоянии перекрестка (информация о дейвайсе)

	subRout.HandleFunc("/{slug}/cross/DispatchControlButtons", whandlers.DispatchControlButtons).Methods("POST") //обработчик диспетчерского управления (отправка команд управления)

	subRout.HandleFunc("/{slug}/cross/control", func(w http.ResponseWriter, r *http.Request) { //расширеная страничка настройки перекрестка (страничка)
		http.ServeFile(w, r, resourcePath+"/crossControl.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/cross/control", whandlers.ControlCross).Methods("POST")                     //данные по расширенной странички перекрестков
	subRout.HandleFunc("/{slug}/cross/control/close", whandlers.ControlCloseCross).Methods("GET")           //обработчик закрытия перекрестка
	subRout.HandleFunc("/{slug}/cross/control/editable", whandlers.ControlEditableCross).Methods("GET")     //обработчик контроля управления перекрестка
	subRout.HandleFunc("/{slug}/cross/control/sendButton", whandlers.ControlSendButton).Methods("POST")     //обработчик приема данных от пользователя для отправки на устройство
	subRout.HandleFunc("/{slug}/cross/control/checkButton", whandlers.ControlCheckButton).Methods("POST")   //обработчик проверки данных
	subRout.HandleFunc("/{slug}/cross/control/createButton", whandlers.ControlCreateButton).Methods("POST") //обработчик создания перекрестка
	subRout.HandleFunc("/{slug}/cross/control/deleteButton", whandlers.ControlDeleteButton).Methods("POST") //обработчик обработчик удаления перекрсетка

	subRout.HandleFunc("/{slug}/manage", func(w http.ResponseWriter, r *http.Request) { //обработка создание и редактирования пользователя (страничка)
		http.ServeFile(w, r, resourcePath+"/manage.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/manage", whandlers.DisplayAccInfo).Methods("POST")          //обработка создание и редактирования пользователя
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
	subRout.HandleFunc("/{slug}/manage/stateTest", whandlers.ControlTestState).Methods("POST") //обработчик проверки структуры State

	subRout.HandleFunc("/{slug}/manage/serverLog", func(w http.ResponseWriter, r *http.Request) { //обработка лог файлов сервера (страничка)
		http.ServeFile(w, r, resourcePath+"/serverLog.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/manage/serverLog", whandlers.DisplayServerLogFile).Methods("POST")     //обработчик по выгрузке лог файлов сервера
	subRout.HandleFunc("/{slug}/manage/serverLog/info", whandlers.DisplayServerLogInfo).Methods("GET") //обработчик выбранного лог файла сервера

	subRout.HandleFunc("/{slug}/manage/crossCreator", func(w http.ResponseWriter, r *http.Request) { //обработка проверки/создания каталога карты перекрестков (страничка)
		http.ServeFile(w, r, resourcePath+"/crossCreator.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/manage/crossCreator", whandlers.MainCrossCreator).Methods("POST")                    //обработка проверки/создания каталога карты перекрестков
	subRout.HandleFunc("/{slug}/manage/crossCreator/checkAllCross", whandlers.CheckAllCross).Methods("POST")         //обработка проверки наличия всех каталогов и файлов необходимых для построения перекрестков
	subRout.HandleFunc("/{slug}/manage/crossCreator/checkSelected", whandlers.CheckSelectedDirCross).Methods("POST") //обработка проверки наличия выбранных каталогов и файлов необходимых для построения перекрестков
	subRout.HandleFunc("/{slug}/manage/crossCreator/makeSelected", whandlers.MakeSelectedDirCross).Methods("POST")   //обработка создания каталога карты перекрестков

	subRout.HandleFunc("/{slug}/map/deviceLog", func(w http.ResponseWriter, r *http.Request) { //обработка лога устройства (страничка)
		http.ServeFile(w, r, resourcePath+"/deviceLog.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/map/deviceLog", whandlers.DisplayDeviceLogFile).Methods("POST") //обработка лога устройства
	subRout.HandleFunc("/{slug}/map/deviceLog/info", whandlers.LogDeviceInfo).Methods("POST")   //обработка лога устройства по выбранному интеревалу времени

	subRout.HandleFunc("/{slug}/license", func(w http.ResponseWriter, r *http.Request) { //обработка работы с лицензиями (страничка)
		http.ServeFile(w, r, resourcePath+"/license.html")
	}).Methods("GET")
	subRout.HandleFunc("/{slug}/license", whandlers.LicenseInfo).Methods("POST")               //обработчик сбора начальной информации
	subRout.HandleFunc("/{slug}/license/create", whandlers.LicenseCreateToken).Methods("POST") //обработка создания лицензий

	//тест просто тест!
	subRout.HandleFunc("/{slug}/testtoken", whandlers.TestToken).Methods("POST")
	subRout.HandleFunc("/{slug}/test", whandlers.TestHello).Methods("POST")
	//------------------------------------------------------------------------------------------------------------------
	//роутер для фаил сервера, он закрыт токеном, скачивать могут только авторизированные пользователи
	fileRout := router.PathPrefix("/file").Subrouter()
	fileRout.Use(JwtFile)                                                                                                                             //добавление к роутеру контроля токена
	fileRout.PathPrefix("/cross/").Handler(http.Handler(http.StripPrefix("/file/cross/", http.FileServer(http.Dir("./views/cross"))))).Methods("GET") //описание пути и скриптов для получения файлов для перекрестка
	fileRout.PathPrefix("/img/").Handler(http.Handler(http.StripPrefix("/file/img/", http.FileServer(http.Dir("./views/img"))))).Methods("GET")       //описание пути и скриптов для получения файлов для основных картинок
	fileRout.PathPrefix("/icons/").Handler(http.Handler(http.StripPrefix("/file/icons/", http.FileServer(http.Dir("./views/icons"))))).Methods("GET") //описание пути и скриптов для получения файлов для иконок

	//------------------------------------------------------------------------------------------------------------------
	// Запуск HTTP сервера
	if err = http.ListenAndServeTLS(data.GlobalConfig.ServerIP, "domain.crt", "domain.key", handlers.LoggingHandler(os.Stdout, router)); err != nil {
		logger.Error.Println("|Message: Server can't started: ", err.Error())
		fmt.Println("Server can't started ", err.Error())
	}
}
