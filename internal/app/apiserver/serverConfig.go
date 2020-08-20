package apiserver

var ServerConfig *ServerConf

type ServerConf struct {
	LoggerPath string `toml:"logger_path"` //путь до каталога с логами сервера
	StaticPath string `toml:"static_path"` //путь до каталога static
	FreePath   string `toml:"free_path"`   //путь до каталога free
	WebPath    string `toml:"web_path"`    //путь до каталога web
	ServerIP   string `toml:"server_ip"`   //ip сервера / порт
}

func NewConfig() *ServerConf {
	return &ServerConf{}
}
