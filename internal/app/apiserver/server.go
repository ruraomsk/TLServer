package apiserver

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/app/handlers/chatH"
	"github.com/JanFant/TLServer/internal/app/handlers/crossH"
	"github.com/JanFant/TLServer/internal/app/handlers/greenStreet"
	"github.com/JanFant/TLServer/internal/app/handlers/licenseH"
	"github.com/JanFant/TLServer/internal/app/handlers/mapH"
	"github.com/JanFant/TLServer/internal/app/handlers/techArmH"
	"github.com/JanFant/TLServer/internal/sockets/crossSock"
	"github.com/JanFant/TLServer/internal/sockets/mapSock"
	"github.com/JanFant/TLServer/internal/sockets/techArm"
	"net/http"

	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/logger"

	"github.com/JanFant/TLServer/internal/app/handlers"
	"github.com/JanFant/TLServer/internal/app/middleWare"
	"github.com/JanFant/TLServer/internal/sockets/chat"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//StartServer запуск сервера
func StartServer(conf *ServerConf) {

	go chat.CBroadcast()
	go mapSock.MapBroadcast(data.GetDB())
	go crossSock.CrossBroadcast(data.GetDB())
	go crossSock.ControlBroadcast()
	go techArm.ArmTechBroadcast(data.GetDB())
	go mapSock.GSBroadcast(data.GetDB())

	// Создаем engine для соединений
	router := gin.Default()
	router.Use(cors.Default())

	router.LoadHTMLGlob(conf.WebPath + "/html/**")

	//скрипт и иконка которые должны быть доступны всем
	router.StaticFS("/free", http.Dir(conf.FreePath))

	//заглушка страница 404
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "notFound.html", nil)
	})

	//начальная страница
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/map")
	})

	// основная карта
	router.GET("/map", func(c *gin.Context) {
		c.HTML(http.StatusOK, "map.html", gin.H{"yaKey": license.LicenseFields.YaKey})
	})

	//карта сокет
	router.GET("/mapW", mapH.MapEngine)

	//------------------------------------------------------------------------------------------------------------------
	//обязательный общий путь
	mainRouter := router.Group("/user")
	mainRouter.Use(middleWare.JwtAuth())
	mainRouter.Use(middleWare.AccessControl())

	//--------- SocketS--------------
	//чат
	mainRouter.GET("/:slug/chat", func(c *gin.Context) { //работа с основной страничкой чата (страница)
		c.HTML(http.StatusOK, "chat.html", nil)
	})
	mainRouter.GET("/:slug/chatW", chatH.ChatEngine) //обработчик веб сокета для чата

	//перекресток
	mainRouter.GET("/:slug/cross", func(c *gin.Context) { //работа со странички перекрестков (страничка)
		c.HTML(http.StatusOK, "cross.html", nil)
	})
	mainRouter.GET("/:slug/crossW", crossH.CrossEngine)

	//арм перекрестка
	mainRouter.GET("/:slug/cross/control", func(c *gin.Context) { //расширеная страничка настройки перекрестка (страничка)
		c.HTML(http.StatusOK, "crossControl.html", nil)
	})
	mainRouter.GET("/:slug/cross/controlW", crossH.CrossControlEngine)

	//арм технолога
	mainRouter.GET("/:slug/techArm", func(c *gin.Context) {
		c.HTML(http.StatusOK, "techControl.html", nil)
	})
	mainRouter.GET("/:slug/techArmW", techArmH.TechArmEngine)

	//зеленая улица
	mainRouter.GET("/:slug/greenStreet", func(c *gin.Context) {
		c.HTML(http.StatusOK, "greenStreet.html", gin.H{"yaKey": license.LicenseFields.YaKey})
	})
	mainRouter.GET("/:slug/greenStreetW", greenStreet.GSEngine)
	//--------- SocketS--------------

	//тех. поддержка
	mainRouter.GET("/:slug/techSupp", func(c *gin.Context) { //работа со страничкой тех поддержки
		c.HTML(http.StatusOK, "techSupp.html", nil)
	})
	mainRouter.POST("/:slug/techSupp/send", handlers.TechSupp) //обработчик подключения email тех поддержки

	//управленеие (общее)
	mainRouter.GET("/:slug/manage", func(c *gin.Context) { //обработка создание и редактирования пользователя (страничка)
		c.HTML(http.StatusOK, "manage.html", nil)
	})
	mainRouter.POST("/:slug/manage", handlers.DisplayAccInfo)          //обработка создание и редактирования пользователя
	mainRouter.POST("/:slug/manage/changepw", handlers.ActChangePw)    //обработчик для изменения пароля
	mainRouter.POST("/:slug/manage/delete", handlers.ActDeleteAccount) //обработчик для удаления аккаунтов
	mainRouter.POST("/:slug/manage/add", handlers.ActAddAccount)       //обработчик для добавления аккаунтов
	mainRouter.POST("/:slug/manage/update", handlers.ActUpdateAccount) //обработчик для редактирования данных аккаунта

	//управление занятыми перекрестками
	mainRouter.GET("/:slug/manage/crossEditControl", func(c *gin.Context) { //обработчик по управлению занятых перекрестков (страничка)
		c.HTML(http.StatusOK, "crossEditControl.html", nil)
	})
	mainRouter.POST("/:slug/manage/crossEditControl", handlers.CrossEditInfo)      //обработчик по управлению занятых перекрестков
	mainRouter.POST("/:slug/manage/crossEditControl/free", handlers.CrossEditFree) //обработчик по управлению освобождению перекрестка

	//проверка БД на признак не правльно заполененных state
	mainRouter.GET("/:slug/manage/stateTest", func(c *gin.Context) { //обработчик проверки всего State (страничка)
		c.HTML(http.StatusOK, "stateTest.html", nil)
	})
	mainRouter.POST("/:slug/manage/stateTest", crossH.ControlTestState) //обработчик проверки структуры State

	//управление логом сервера
	mainRouter.GET("/:slug/manage/serverLog", func(c *gin.Context) { //обработка лог файлов сервера (страничка)
		c.HTML(http.StatusOK, "serverLog.html", nil)
	})
	mainRouter.POST("/:slug/manage/serverLog", handlers.DisplayServerLogFile)     //обработчик по выгрузке лог файлов сервера
	mainRouter.GET("/:slug/manage/serverLog/info", handlers.DisplayServerLogInfo) //обработчик выбранного лог файла сервера

	//проверка/создание каталога перекрестков (страничка не реализованна)
	mainRouter.GET("/:slug/manage/crossCreator", func(c *gin.Context) { //обработка проверки/создания каталога карты перекрестков (страничка)
		c.HTML(http.StatusOK, "crossCreator.html", nil)
	})
	mainRouter.POST("/:slug/manage/crossCreator", handlers.MainCrossCreator)                    //обработка проверки/создания каталога карты перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/checkAllCross", handlers.CheckAllCross)         //обработка проверки наличия всех каталогов и файлов необходимых для построения перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/checkSelected", handlers.CheckSelectedDirCross) //обработка проверки наличия выбранных каталогов и файлов необходимых для построения перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/makeSelected", handlers.MakeSelectedDirCross)   //обработка создания каталога карты перекрестков

	//лог устройств
	mainRouter.GET("/:slug/map/deviceLog", func(c *gin.Context) { //обработка лога устройства (страничка)
		c.HTML(http.StatusOK, "deviceLog.html", nil)
	})
	mainRouter.POST("/:slug/map/deviceLog", handlers.DisplayDeviceLogFile) //обработка лога устройства
	mainRouter.POST("/:slug/map/deviceLog/info", handlers.LogDeviceInfo)   //обработка лога устройства по выбранному интеревалу времени

	//работа с лицензией
	mainRouter.GET("/:slug/license", func(c *gin.Context) { //обработка работы с лицензиями (страничка)
		c.HTML(http.StatusOK, "license.html", nil)
	})
	mainRouter.POST("/:slug/license", licenseH.LicenseInfo)            //обработчик сбора начальной информаци
	mainRouter.POST("/:slug/license/newToken", licenseH.LicenseNewKey) //обработчик сохранения нового токена

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
	if err := router.Run(conf.ServerIP); err != nil {
		logger.Error.Println("|Message: Error start server ", err.Error())
		fmt.Println("Error start server ", err.Error())
	}
}
