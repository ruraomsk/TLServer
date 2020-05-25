package apiserver

var ServerConfig *ServerConf

type ServerConf struct {
	LoggerPath string `toml:"logger_path"` //путь до каталога с логами сервера
	StaticPath string `toml:"static_path"` //путь до каталога static
	WebPath    string `toml:"web_path"`    //путь до каталога web
	SSLPath    string `toml:"ssl_path"`    //путь до каталога ssl
	ServerIP   string `toml:"server_ip"`   //ip сервера / порт

}

func NewConfig() *ServerConf {
	return &ServerConf{}
}
