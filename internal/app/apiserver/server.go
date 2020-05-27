package apiserver

import (
	"fmt"
	"net/http"

	"github.com/JanFant/TLServer/internal/app/handlers"
	"github.com/JanFant/TLServer/internal/app/middleWare"
	"github.com/JanFant/TLServer/internal/model/chat"
	"github.com/JanFant/TLServer/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var err error

//StartServer запуск сервера
func StartServer(conf *ServerConf) {
	chat.Connections = make(map[*websocket.Conn]string)
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
	mainRouter.Use(middleWare.AccessControl())
	//subRout.Use(AccessControl) //проверка разрешения на доступ к ресурсу

	mainRouter.GET("/:slug/map", func(c *gin.Context) { //работа с основной страничкой карты
		c.HTML(http.StatusOK, "map.html", gin.H{"message": "map page"})
	})
	mainRouter.POST("/:slug/map", handlers.BuildMapPage)                         //запрос информации для заполнения странички с картой
	mainRouter.POST("/:slug/map/logOut", handlers.LoginAccOut)                   //обработчик выхода из системы
	mainRouter.POST("/:slug/map/update", handlers.UpdateMapPage)                 //обновление странички с данными которые попали в область пользователя
	mainRouter.POST("/:slug/map/locationButton", handlers.LocationButtonMapPage) //обработчик для формирования новых координат отображения карты

	mainRouter.GET("/:slug/chat", func(c *gin.Context) { //работа с основной страничкой чата (страница)
		c.HTML(http.StatusOK, "chat.html", gin.H{"message": "chat page"})
	})
	mainRouter.GET("/:slug/chatW", handlers.ChatEngine) //обработчик веб сокета для чата

	mainRouter.GET("/:slug/techSupp", func(c *gin.Context) { //работа со страничкой тех поддержки
		c.HTML(http.StatusOK, "techSupp.html", gin.H{"message": "map page"})
	})
	mainRouter.POST("/:slug/techSupp/send", handlers.TechSupp) //обработчик подключения email тех поддержки

	//---------------------------------------------------------------------------------------------------------------------------------------------------
	mainRouter.GET("/:slug/cross", func(c *gin.Context) { //работа со странички перекрестков (страничка)
		c.HTML(http.StatusOK, "cross.html", gin.H{"message": "cross page"})
	})
	mainRouter.POST("/:slug/cross", handlers.BuildCross)                                    //информация о состоянии перекрёстка
	mainRouter.POST("/:slug/cross/dev", handlers.DevCrossInfo)                              //информация о состоянии перекрестка (информация о дейвайсе)
	mainRouter.POST("/:slug/cross/DispatchControlButtons", handlers.DispatchControlButtons) //обработчик диспетчерского управления (отправка команд управления)

	mainRouter.GET("/:slug/cross/control", func(c *gin.Context) { //расширеная страничка настройки перекрестка (страничка)
		c.HTML(http.StatusOK, "crossControl.html", gin.H{"message": "crossControl page"})
	})
	mainRouter.POST("/:slug/cross/control", handlers.ControlCross)                     //данные по расширенной странички перекрестков
	mainRouter.GET("/:slug/cross/control/close", handlers.ControlCloseCross)           //обработчик закрытия перекрестка
	mainRouter.GET("/:slug/cross/control/editable", handlers.ControlEditableCross)     //обработчик контроля управления перекрестка
	mainRouter.POST("/:slug/cross/control/sendButton", handlers.ControlSendButton)     //обработчик приема данных от пользователя для отправки на устройство
	mainRouter.POST("/:slug/cross/control/checkButton", handlers.ControlCheckButton)   //обработчик проверки данных
	mainRouter.POST("/:slug/cross/control/createButton", handlers.ControlCreateButton) //обработчик создания перекрестка
	mainRouter.POST("/:slug/cross/control/deleteButton", handlers.ControlDeleteButton) //обработчик обработчик удаления перекрсетка

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

	mainRouter.GET("/:slug/manage/stateTest", func(c *gin.Context) { //обработчик проверки всего State (страничка)
		c.HTML(http.StatusOK, "stateTest.html", gin.H{"message": "stateTest page"})
	})
	mainRouter.POST("/:slug/manage/stateTest", handlers.ControlTestState) //обработчик проверки структуры State

	mainRouter.GET("/:slug/manage/serverLog", func(c *gin.Context) { //обработка лог файлов сервера (страничка)
		c.HTML(http.StatusOK, "serverLog.html", gin.H{"message": "serverLog page"})
	})
	mainRouter.POST("/:slug/manage/serverLog", handlers.DisplayServerLogFile)     //обработчик по выгрузке лог файлов сервера
	mainRouter.GET("/:slug/manage/serverLog/info", handlers.DisplayServerLogInfo) //обработчик выбранного лог файла сервера

	mainRouter.GET("/:slug/manage/crossCreator", func(c *gin.Context) { //обработка проверки/создания каталога карты перекрестков (страничка)
		c.HTML(http.StatusOK, "crossCreator.html", gin.H{"message": "crossCreator page"})
	})

	mainRouter.POST("/:slug/manage/crossCreator", handlers.MainCrossCreator)                    //обработка проверки/создания каталога карты перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/checkAllCross", handlers.CheckAllCross)         //обработка проверки наличия всех каталогов и файлов необходимых для построения перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/checkSelected", handlers.CheckSelectedDirCross) //обработка проверки наличия выбранных каталогов и файлов необходимых для построения перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/makeSelected", handlers.MakeSelectedDirCross)   //обработка создания каталога карты перекрестков

	mainRouter.GET("/:slug/map/deviceLog", func(c *gin.Context) { //обработка лога устройства (страничка)
		c.HTML(http.StatusOK, "deviceLog.html", gin.H{"message": "crossEdit page"})
	})
	mainRouter.POST("/:slug/map/deviceLog", handlers.DisplayDeviceLogFile) //обработка лога устройства
	mainRouter.POST("/:slug/map/deviceLog/info", handlers.LogDeviceInfo)   //обработка лога устройства по выбранному интеревалу времени

	mainRouter.GET("/:slug/license", func(c *gin.Context) { //обработка работы с лицензиями (страничка)
		c.HTML(http.StatusOK, "license.html", gin.H{"message": "license page"})
	})
	mainRouter.POST("/:slug/license", handlers.LicenseInfo)            //обработчик сбора начальной информаци
	mainRouter.POST("/:slug/license/newToken", handlers.LicenseNewKey) //обработчик сохранения нового токена

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
