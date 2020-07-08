package tcpConnect

//TCPConfig настройки для тсп соединения
type TCPConfig struct {
	ServerAddr  string `toml:"tcpServerAddress"` //адресс сервера
	PortState   string `toml:"portState"`        //порт для обмена Стате
	PortArmComm string `toml:"portArmCommand"`   //порт для обмена арм командами
}

//getStateIP возвращает ip+port для State соединения
func (tcpConfig *TCPConfig) getStateIP() string {
	return tcpConfig.ServerAddr + tcpConfig.PortState
}

//getArmIP возвращает ip+port для ArmCommand соединения
func (tcpConfig *TCPConfig) getArmIP() string {
	return tcpConfig.ServerAddr + tcpConfig.PortArmComm
}
