package apiserver

import (
	"fmt"
	"github.com/JanFant/newTLServer/internal/app/handlers"
	"github.com/JanFant/newTLServer/internal/app/middleWare"
	"github.com/JanFant/newTLServer/internal/model/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

var err error

//StartServer запуск сервера
func StartServer(conf *ServerConf) {
	//data.Connections = make(map[*websocket.Conn]string)
	//data.Names.Users = make(map[string]bool)

	// Создаем engine для соединений
	router := gin.Default()
	router.Use(cors.Default())

	router.LoadHTMLGlob(conf.WebPath + "/html/**")

	//скрипт и иконка которые должны быть доступны всем
	router.StaticFile("screen/screen.js", conf.WebPath+"/js/screen.js")
	router.StaticFile("icon/trafficlight.svg", conf.WebPath+"/resources/trafficlight.svg")
	router.StaticFile("/notFound.jpg", conf.WebPath+"/resources/notFound.jpg")

	//заглушка страница 404
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "notFound.html", gin.H{"message": "page not found"})
	})

	//начальная страница
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "screen.html", gin.H{"message": "login page"})
	})
	router.POST("/login", handlers.LoginAcc) //запрос на вход в систему

	//------------------------------------------------------------------------------------------------------------------
	//обязательный общий путь
	mainRouter := router.Group("/user")
	mainRouter.Use(middleWare.JwtAuth())
	//subRout.Use(AccessControl) //проверка разрешения на доступ к ресурсу

	mainRouter.GET("/:slug/map", func(c *gin.Context) { //работа с основной страничкой карты
		c.HTML(http.StatusOK, "map.html", gin.H{"message": "map page"})
	})
	mainRouter.POST("/:slug/map", handlers.BuildMapPage)                         //запрос информации для заполнения странички с картой
	mainRouter.GET("/:slug/map/logOut", handlers.LoginAccOut)                    //обработчик выхода из системы
	mainRouter.POST("/:slug/map/update", handlers.UpdateMapPage)                 //обновление странички с данными которые попали в область пользователя
	mainRouter.POST("/:slug/map/locationButton", handlers.LocationButtonMapPage) //обработчик для формирования новых координат отображения карты

	//subRout.HandleFunc("/{slug}/chat", func(w http.ResponseWriter, r *http.Request) { //начальная страница
	//	http.ServeFile(w, r, resourcePath+"/chat.html")
	//})
	//subRout.HandleFunc("/{slug}/chatTest", func(w http.ResponseWriter, r *http.Request) { //начальная страница
	//	http.ServeFile(w, r, resourcePath+"/chatTest.html")
	//})
	//subRout.HandleFunc("/{slug}/chatW", whandlers.Chat).Methods("GET")
	//

	//subRout.HandleFunc("/{slug}/techSupp", func(w http.ResponseWriter, r *http.Request) { //работа со страничкой тех поддержки
	//	http.ServeFile(w, r, resourcePath+"/techSupp.html")
	//}).Methods("GET")
	//subRout.HandleFunc("/{slug}/techSupp/send", whandlers.TechSupp).Methods("POST") //обработчик подключения email тех поддержки
	//

	//subRout.HandleFunc("/{slug}/cross", func(w http.ResponseWriter, r *http.Request) { //работа со странички перекрестков (страничка)
	//	http.ServeFile(w, r, resourcePath+"/cross.html")
	//}).Methods("GET")
	//subRout.HandleFunc("/{slug}/cross", whandlers.BuildCross).Methods("POST")       //информация о состоянии перекрёстка
	//subRout.HandleFunc("/{slug}/cross/dev", whandlers.DevCrossInfo).Methods("POST") //информация о состоянии перекрестка (информация о дейвайсе)
	//
	//subRout.HandleFunc("/{slug}/cross/DispatchControlButtons", whandlers.DispatchControlButtons).Methods("POST") //обработчик диспетчерского управления (отправка команд управления)
	//
	//subRout.HandleFunc("/{slug}/cross/control", func(w http.ResponseWriter, r *http.Request) { //расширеная страничка настройки перекрестка (страничка)
	//	http.ServeFile(w, r, resourcePath+"/crossControl.html")
	//}).Methods("GET")
	//subRout.HandleFunc("/{slug}/cross/control", whandlers.ControlCross).Methods("POST")                     //данные по расширенной странички перекрестков
	//subRout.HandleFunc("/{slug}/cross/control/close", whandlers.ControlCloseCross).Methods("GET")           //обработчик закрытия перекрестка
	//subRout.HandleFunc("/{slug}/cross/control/editable", whandlers.ControlEditableCross).Methods("GET")     //обработчик контроля управления перекрестка
	//subRout.HandleFunc("/{slug}/cross/control/sendButton", whandlers.ControlSendButton).Methods("POST")     //обработчик приема данных от пользователя для отправки на устройство
	//subRout.HandleFunc("/{slug}/cross/control/checkButton", whandlers.ControlCheckButton).Methods("POST")   //обработчик проверки данных
	//subRout.HandleFunc("/{slug}/cross/control/createButton", whandlers.ControlCreateButton).Methods("POST") //обработчик создания перекрестка
	//subRout.HandleFunc("/{slug}/cross/control/deleteButton", whandlers.ControlDeleteButton).Methods("POST") //обработчик обработчик удаления перекрсетка
	//

	mainRouter.GET("/:slug/manage", func(c *gin.Context) { //обработка создание и редактирования пользователя (страничка)
		c.HTML(http.StatusOK, "manage.html", gin.H{"message": "manage page"})
	})
	mainRouter.POST("/:slug/manage", handlers.DisplayAccInfo)          //обработка создание и редактирования пользователя
	mainRouter.POST("/:slug/manage/changepw", handlers.ActChangePw)    //обработчик для изменения пароля
	mainRouter.POST("/:slug/manage/delete", handlers.ActDeleteAccount) //обработчик для удаления аккаунтов
	mainRouter.POST("/:slug/manage/add", handlers.ActAddAccount)       //обработчик для добавления аккаунтов
	mainRouter.POST("/:slug/manage/update", handlers.ActUpdateAccount) //обработчик для редактирования данных аккаунта

	mainRouter.GET("/:slug/manage/crossEditControl", func(c *gin.Context) { //обработчик по управлению занятых перекрестков (страничка)
		c.HTML(http.StatusOK, "crossEditControl.html", gin.H{"message": "crossEdit page"})
	})
	mainRouter.POST("/:slug/manage/crossEditControl", handlers.CrossEditInfo)      //обработчик по управлению занятых перекрестков
	mainRouter.POST("/:slug/manage/crossEditControl/free", handlers.CrossEditFree) //обработчик по управлению освобождению перекрестка

	//subRout.HandleFunc("/{slug}/manage/stateTest", func(w http.ResponseWriter, r *http.Request) { //обработчик проверки всего State (страничка)
	//	http.ServeFile(w, r, resourcePath+"/stateTest.html")
	//}).Methods("GET")
	//subRout.HandleFunc("/{slug}/manage/stateTest", whandlers.ControlTestState).Methods("POST") //обработчик проверки структуры State
	//

	mainRouter.GET("/:slug/manage/serverLog", func(c *gin.Context) { //обработка лог файлов сервера (страничка)
		c.HTML(http.StatusOK, "serverLog.html", gin.H{"message": "serverLog page"})
	})
	mainRouter.POST("/:slug/manage/serverLog", handlers.DisplayServerLogFile)     //обработчик по выгрузке лог файлов сервера
	mainRouter.GET("/:slug/manage/serverLog/info", handlers.DisplayServerLogInfo) //обработчик выбранного лог файла сервера

	//subRout.HandleFunc("/{slug}/manage/crossCreator", func(w http.ResponseWriter, r *http.Request) { //обработка проверки/создания каталога карты перекрестков (страничка)
	//	http.ServeFile(w, r, resourcePath+"/crossCreator.html")
	//}).Methods("GET")
	//subRout.HandleFunc("/{slug}/manage/crossCreator", whandlers.MainCrossCreator).Methods("POST")                    //обработка проверки/создания каталога карты перекрестков
	//subRout.HandleFunc("/{slug}/manage/crossCreator/checkAllCross", whandlers.CheckAllCross).Methods("POST")         //обработка проверки наличия всех каталогов и файлов необходимых для построения перекрестков
	//subRout.HandleFunc("/{slug}/manage/crossCreator/checkSelected", whandlers.CheckSelectedDirCross).Methods("POST") //обработка проверки наличия выбранных каталогов и файлов необходимых для построения перекрестков
	//subRout.HandleFunc("/{slug}/manage/crossCreator/makeSelected", whandlers.MakeSelectedDirCross).Methods("POST")   //обработка создания каталога карты перекрестков
	//

	mainRouter.GET("/:slug/map/deviceLog", func(c *gin.Context) { //обработка лога устройства (страничка)
		c.HTML(http.StatusOK, "deviceLog.html", gin.H{"message": "crossEdit page"})
	})
	mainRouter.POST("/:slug/map/deviceLog", handlers.DisplayDeviceLogFile) //обработка лога устройства
	mainRouter.POST("/:slug/map/deviceLog/info", handlers.LogDeviceInfo)   //обработка лога устройства по выбранному интеревалу времени

	//subRout.HandleFunc("/{slug}/license", func(w http.ResponseWriter, r *http.Request) { //обработка работы с лицензиями (страничка)
	//	http.ServeFile(w, r, resourcePath+"/license.html")
	//}).Methods("GET")
	//subRout.HandleFunc("/{slug}/license", whandlers.LicenseInfo).Methods("POST")               //обработчик сбора начальной информации
	//subRout.HandleFunc("/{slug}/license/create", whandlers.LicenseCreateToken).Methods("POST") //обработка создания лицензий
	//subRout.HandleFunc("/{slug}/license/newToken", whandlers.LicenseNewKey).Methods("POST")    //обработчик сохранения нового токена
	//

	////тест просто тест!
	//subRout.HandleFunc("/{slug}/testtoken", whandlers.TestToken).Methods("POST")
	//subRout.HandleFunc("/{slug}/test", whandlers.TestHello).Methods("POST")

	//------------------------------------------------------------------------------------------------------------------
	//роутер для фаил сервера, он закрыт токеном, скачивать могут только авторизированные пользователи

	fileServer := router.Group("/file")
	fileServer.Use(middleWare.JwtFile())

	fsStatic := fileServer.Group("/static")
	fsStatic.StaticFS("/cross", http.Dir(conf.StaticPath+"/cross"))
	fsStatic.StaticFS("/icons", http.Dir(conf.StaticPath+"/icons"))
	fsStatic.StaticFS("/img", http.Dir(conf.StaticPath+"/img"))
	fsStatic.StaticFS("/markdown", http.Dir(conf.StaticPath+"/markdown"))

	fsWeb := fileServer.Group("/web")
	fsWeb.StaticFS("/resources", http.Dir(conf.WebPath+"/resources"))
	fsWeb.StaticFS("/js", http.Dir(conf.WebPath+"/js"))
	fsWeb.StaticFS("/css", http.Dir(conf.WebPath+"/css"))

	//------------------------------------------------------------------------------------------------------------------
	// Запуск HTTP сервера
	if err := router.RunTLS(conf.ServerIP, conf.SSLPath+"/domain.crt", conf.SSLPath+"/domain.key"); err != nil {
		logger.Error.Println("|Message: Error start server ", err.Error())
		fmt.Println("Error start server ", err.Error())
	}

}
