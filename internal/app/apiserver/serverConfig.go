package apiserver

import "github.com/JanFant/newTLServer/internal/app/config"

type ServerConf struct {
	Port    string
	ResPath string
	SSLPath string
}

func NewConfig() *ServerConf {
	return &ServerConf{
		Port:    config.GlobalConfig.ServerIP,
		ResPath: config.GlobalConfig.ResourcePath,
		SSLPath: config.GlobalConfig.SSLPath,
	}
}
