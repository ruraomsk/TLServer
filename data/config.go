package data

import "github.com/JanFant/TLServer/tcpConnect"

//GlobalConfig глобальная переменная для структуры конфиг
var GlobalConfig *Config

//Config структура с объявлением всех переменных config.toml файла
type Config struct {
	LoggerPath    string               `toml:"logger_path"`    //путь до каталога с логами сервера
	ViewsPath     string               `toml:"views_path"`     //путь до каталога views (содержит все ресурсы для отображения перекрестков)
	CachePath     string               `toml:"cache_path"`     //путь до каталога cachefile
	ResourcePath  string               `toml:"resourcePath"`   //путь до каталога frontend
	ServerIP      string               `toml:"server_ip"`      //ip сервера / порт
	YaKey         string               `toml:"ya_key"`         //ключ авторизации для яндекса
	TokenPassword string               `toml:"token_password"` //ключ для шифрования токенов доступа
	TCPConfig     tcpConnect.TCPConfig `toml:"tcpServer"`      //информация о tcp соединении с сервером устройств
	DBConfig      DBConfig             `toml:"database"`       //информация о соединении с базой данных
}

//NewConfig создание конфига
func NewConfig() *Config {
	return &Config{}
}
