package logger

var LogGlobalConf *LogConfig

type LogConfig struct {
	LogPath    string `toml:"logger_path"`
	GinLogPath string `toml:"ginLog_path"`
}

func NewConfig() *LogConfig {
	return &LogConfig{}
}
