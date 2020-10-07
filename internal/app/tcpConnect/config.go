package tcpConnect

//TCPConfig настройки для тсп соединения
type TCPConfig struct {
	ServerAddr         string `toml:"tcpServerAddress"`   //адресс сервера
	PortState          string `toml:"portState"`          //порт для обмена Стате
	PortArmComm        string `toml:"portArmCommand"`     //порт для обмена арм командами
	PortDkConn         string `toml:"portDkConn"`         //порт для обмена сообщениями DK
	PortChangeProtocol string `toml:"portChangeProtocol"` //порт для обмена сообщениями смены протокола
	PortGPRS           string `toml:"portGPRS"`
	SendGPRS           bool   `toml:"sendGPRS"`
}

//getStateIP возвращает ip+port для State соединения
func (tcpConfig *TCPConfig) getStateIP() string {
	return tcpConfig.ServerAddr + ":" + tcpConfig.PortState
}

//getArmIP возвращает ip+port для ArmCommand соединения
func (tcpConfig *TCPConfig) getArmIP() string {
	return tcpConfig.ServerAddr + ":" + tcpConfig.PortArmComm
}

//getChangeProtocolIP возвращает ip+port для gprs соединения
func (tcpConfig *TCPConfig) getChangeProtocolIP() string {
	return tcpConfig.ServerAddr + ":" + tcpConfig.PortChangeProtocol
}

//getPhaseIP возвращает ip+port для фазы
func (tcpConfig *TCPConfig) getPhaseIP() string {
	return tcpConfig.ServerAddr + ":" + tcpConfig.PortDkConn
}
