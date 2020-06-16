package apiserver

import (
	"fmt"
	"github.com/JanFant/TLServer/internal/model/data"
	"github.com/JanFant/TLServer/internal/model/license"
	"github.com/JanFant/TLServer/logger"
	"net/http"

	"github.com/JanFant/TLServer/internal/app/handlers"
	"github.com/JanFant/TLServer/internal/app/middleWare"
	"github.com/JanFant/TLServer/internal/model/chat"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//StartServer запуск сервера
func StartServer(conf *ServerConf) {

	go chat.CBroadcast()
	go data.MapBroadcast()
	go data.CrossBroadcast()
	go data.ControlBroadcast()

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
	router.GET("/map", func(c *gin.Context) {
		c.HTML(http.StatusOK, "map.html", gin.H{"yaKey": license.LicenseFields.YaKey})
	})
	router.GET("/mapW", handlers.MapEngine)

	router.GET("/map/screen", func(c *gin.Context) {
		c.HTML(http.StatusOK, "screen.html", nil)
	})

	//cross WebSocket
	router.GET("/crossTest", func(c *gin.Context) {
		c.HTML(http.StatusOK, "crossTest.html", nil)
	})
	router.GET("/crossW", handlers.CrossEngine)

	//------------------------------------------------------------------------------------------------------------------
	//обязательный общий путь
	mainRouter := router.Group("/user")
	mainRouter.Use(middleWare.JwtAuth())
	mainRouter.Use(middleWare.AccessControl())

	mainRouter.GET("/:slug/chat", func(c *gin.Context) { //работа с основной страничкой чата (страница)
		c.HTML(http.StatusOK, "chat.html", nil)
	})
	mainRouter.GET("/:slug/chatW", handlers.ChatEngine) //обработчик веб сокета для чата

	mainRouter.GET("/:slug/techSupp", func(c *gin.Context) { //работа со страничкой тех поддержки
		c.HTML(http.StatusOK, "techSupp.html", nil)
	})
	mainRouter.POST("/:slug/techSupp/send", handlers.TechSupp) //обработчик подключения email тех поддержки

	//---------------------------------------------------------------------------------------------------------------------------------------------------
	mainRouter.GET("/:slug/cross", func(c *gin.Context) { //работа со странички перекрестков (страничка)
		c.HTML(http.StatusOK, "cross.html", nil)
	})
	mainRouter.GET("/:slug/crossW", handlers.CrossEngine)

	mainRouter.GET("/:slug/cross/control", func(c *gin.Context) { //расширеная страничка настройки перекрестка (страничка)
		c.HTML(http.StatusOK, "crossControl.html", nil)
	})
	mainRouter.GET("/:slug/cross/controlW", handlers.CrossControlEngine)

	mainRouter.GET("/:slug/manage", func(c *gin.Context) { //обработка создание и редактирования пользователя (страничка)
		c.HTML(http.StatusOK, "manage.html", nil)
	})
	mainRouter.POST("/:slug/manage", handlers.DisplayAccInfo)          //обработка создание и редактирования пользователя
	mainRouter.POST("/:slug/manage/changepw", handlers.ActChangePw)    //обработчик для изменения пароля
	mainRouter.POST("/:slug/manage/delete", handlers.ActDeleteAccount) //обработчик для удаления аккаунтов
	mainRouter.POST("/:slug/manage/add", handlers.ActAddAccount)       //обработчик для добавления аккаунтов
	mainRouter.POST("/:slug/manage/update", handlers.ActUpdateAccount) //обработчик для редактирования данных аккаунта

	mainRouter.GET("/:slug/manage/crossEditControl", func(c *gin.Context) { //обработчик по управлению занятых перекрестков (страничка)
		c.HTML(http.StatusOK, "crossEditControl.html", nil)
	})
	mainRouter.POST("/:slug/manage/crossEditControl", handlers.CrossEditInfo) //обработчик по управлению занятых перекрестков
	//mainRouter.POST("/:slug/manage/crossEditControl/free", handlers.CrossEditFree) //обработчик по управлению освобождению перекрестка

	mainRouter.GET("/:slug/manage/stateTest", func(c *gin.Context) { //обработчик проверки всего State (страничка)
		c.HTML(http.StatusOK, "stateTest.html", nil)
	})
	mainRouter.POST("/:slug/manage/stateTest", handlers.ControlTestState) //обработчик проверки структуры State

	mainRouter.GET("/:slug/manage/serverLog", func(c *gin.Context) { //обработка лог файлов сервера (страничка)
		c.HTML(http.StatusOK, "serverLog.html", nil)
	})
	mainRouter.POST("/:slug/manage/serverLog", handlers.DisplayServerLogFile)     //обработчик по выгрузке лог файлов сервера
	mainRouter.GET("/:slug/manage/serverLog/info", handlers.DisplayServerLogInfo) //обработчик выбранного лог файла сервера

	mainRouter.GET("/:slug/manage/crossCreator", func(c *gin.Context) { //обработка проверки/создания каталога карты перекрестков (страничка)
		c.HTML(http.StatusOK, "crossCreator.html", nil)
	})

	mainRouter.POST("/:slug/manage/crossCreator", handlers.MainCrossCreator)                    //обработка проверки/создания каталога карты перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/checkAllCross", handlers.CheckAllCross)         //обработка проверки наличия всех каталогов и файлов необходимых для построения перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/checkSelected", handlers.CheckSelectedDirCross) //обработка проверки наличия выбранных каталогов и файлов необходимых для построения перекрестков
	mainRouter.POST("/:slug/manage/crossCreator/makeSelected", handlers.MakeSelectedDirCross)   //обработка создания каталога карты перекрестков

	mainRouter.GET("/:slug/map/deviceLog", func(c *gin.Context) { //обработка лога устройства (страничка)
		c.HTML(http.StatusOK, "deviceLog.html", nil)
	})
	mainRouter.POST("/:slug/map/deviceLog", handlers.DisplayDeviceLogFile) //обработка лога устройства
	mainRouter.POST("/:slug/map/deviceLog/info", handlers.LogDeviceInfo)   //обработка лога устройства по выбранному интеревалу времени

	mainRouter.GET("/:slug/license", func(c *gin.Context) { //обработка работы с лицензиями (страничка)
		c.HTML(http.StatusOK, "license.html", nil)
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
	if err := router.Run(conf.ServerIP); err != nil {
		logger.Error.Println("|Message: Error start server ", err.Error())
		fmt.Println("Error start server ", err.Error())
	}
}
