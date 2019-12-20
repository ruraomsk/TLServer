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

func StartServer() {
	// Создаем новый ServeMux для HTTPS соединений
	router := mux.NewRouter()
	//основной обработчик
	//начальная страница
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./views/screen.html")
	})
	//путь к скриптам они открыты
	router.PathPrefix("/static/").Handler(http.Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("//Fileserver/общая папка/TEMP рабочий/Semyon/lib/js"))))).Methods("GET")

	//запрос на вход в систему
	router.HandleFunc("/login", whandlers.LoginAcc).Methods("POST")
	router.HandleFunc("/test", whandlers.TestHello).Methods("POST")
	router.HandleFunc("/create", whandlers.CreateAcc).Methods("POST")

	subRout := router.PathPrefix("/user").Subrouter()

	subRout.Use(JwtAuth)
	//запрос на создание пользователя
	subRout.HandleFunc("/{slug}/create", whandlers.CreateAcc).Methods("POST")
	//запрос странички с картой
	subRout.HandleFunc("/{slug}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./views/workplace.html")
	}).Methods("GET")
	//запрос информации для заполнения странички с картой
	subRout.HandleFunc("/{slug}", whandlers.BuildMapPage).Methods("POST")
	//обновление странички с данными которые попали в область пользователя
	subRout.HandleFunc("/{slug}/update", whandlers.UpdateMapPage).Methods("POST")
	//запрос странички с перекрестком
	subRout.HandleFunc("/{slug}/cross", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./views/cross.html")
	}).Methods("GET")
	//отправка информации с состояниями перекреста
	subRout.HandleFunc("/{slug}/cross", whandlers.BuildCross).Methods("POST")

	//тест
	subRout.HandleFunc("/{slug}/testtoken", whandlers.TestToken).Methods("POST")

	//роутер для фаил сервера, он закрыт токеном, скачивать могут только авторизированные пользователи
	fileRout := router.PathPrefix("/file").Subrouter()
	fileRout.Use(JwtFile)
	fileRout.PathPrefix("/cross/").Handler(http.Handler(http.StripPrefix("/file/cross/", http.FileServer(http.Dir("./views/cross"))))).Methods("GET")
	fileRout.PathPrefix("/img/").Handler(http.Handler(http.StripPrefix("/file/img/", http.FileServer(http.Dir("./views/img"))))).Methods("GET")

	// Запуск HTTP сервера
	// if err = http.ListenAndServe(os.Getenv("server_ip"), handlers.LoggingHandler(os.Stdout, router)); err != nil {
	if err = http.ListenAndServeTLS(os.Getenv("server_ip"), "domain.crt", "domain.key", handlers.LoggingHandler(os.Stdout, router)); err != nil {
		logger.Info.Println("Server can't started ", err.Error())
		fmt.Println("Server can't started ", err.Error())
	}
}
