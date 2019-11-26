package main

import (
	"fmt"
	"net/http"
)

func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	// Перенаправляем входящий HTTP запрос. Учтите,
	// что "127.0.0.1:8081" работает только для вашей локальной машина

	http.Redirect(w, r, "https://127.0.0.1:8081"+r.RequestURI,
		http.StatusMovedPermanently)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there!")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi admin!")
}

func main() {
	// Создаем новый ServeMux для HTTP соединений
	httpMux := http.NewServeMux()
	// Создаем новый ServeMux для HTTPS соединений
	httpsMux := http.NewServeMux()
	// Перенаправляем /admin/ на HTTPS
	httpMux.Handle("/admin/", http.HandlerFunc(redirectToHttps))
	// Обрабатываем все остальное
	httpMux.Handle("/", http.HandlerFunc(homeHandler))
	// Так же, обрабатываем все по HTTPS
	httpsMux.Handle("/", http.HandlerFunc(homeHandler))
	httpsMux.Handle("/admin/", http.HandlerFunc(adminHandler))
	// Запуск HTTPS сервера в отдельной go-рутине
	go http.ListenAndServeTLS(":8081", "cert.pem", "key.pem", httpsMux)
	// Запуск HTTPS сервера
	http.ListenAndServe(":8080", httpMux)
}
