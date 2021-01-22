package apiserver

var ServerConfig *ServerConf

type ServerConf struct {
	LoggerPath     string `toml:"logger_path"`     //путь до каталога с логами сервера
	StaticPath     string `toml:"static_path"`     //путь до каталога static
	FreePath       string `toml:"free_path"`       //путь до каталога free
	WebPath        string `toml:"web_path"`        //путь до каталога web
	PortHTTP       string `toml:"portHTTP"`        // порт http
	PortHTTPS      string `toml:"portHTTPS"`       // порт https
	ServerExchange string `toml:"server_exchange"` //ip  / порт сервера обмена
}

func NewConfig() *ServerConf {
	return &ServerConf{}
}
