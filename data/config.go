package data

import "github.com/JanFant/TLServer/tcpConnect"

//GlobalConfig глобальная переменная для структуры конфиг
var GlobalConfig *Config

//Config структура с обьявлением всех переменных config.toml файла
type Config struct {
	LoggerPath    string               `toml:"logger_path"`
	ViewsPath     string               `toml:"views_path"`
	ResourcePath  string               `toml:"resourcePath"`
	ServerIP      string               `toml:"server_ip"`
	YaKey         string               `toml:"ya_key"`
	TokenPassword string               `toml:"token_password"`
	TCPConfig     tcpConnect.TCPConfig `toml:"tcpServer"`
	DBConfig      DBConfig             `toml:"database"`
	PngSettings   PngSettings          `toml:"picture"`
}

func NewConfig() *Config {
	return &Config{}
}
