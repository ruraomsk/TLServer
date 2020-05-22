package apiserver

var ServerConfig *ServerConf

type ServerConf struct {
	LoggerPath   string `toml:"logger_path"`   //путь до каталога с логами сервера
	ViewsPath    string `toml:"views_path"`    //путь до каталога views (содержит все ресурсы для отображения перекрестков)
	ResourcePath string `toml:"resource_path"` //путь до каталога frontend
	SSLPath      string `toml:"ssl_path"`      //путь до каталога ssl
	ServerIP     string `toml:"server_ip"`     //ip сервера / порт

}

func NewConfig() *ServerConf {
	return &ServerConf{}
}
