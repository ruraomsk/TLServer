package data

import "fmt"

//GlobalConfig глобальная переменная для структуры конфиг
var GlobalConfig *Config

type DBConfig struct {
	Name            string `toml:"db_name"`            //имя БД
	Password        string `toml:"db_password"`        //пароль доступа к БД
	User            string `toml:"db_user"`            //пользователя для обращения к бд
	Type            string `toml:"db_type"`            //тип бд
	Host            string `toml:"db_host"`            //ip сервера бд
	Port            string `toml:"db_port"`            //порт для обращения к бд
	SetMaxOpenConst int    `toml:"db_SetMaxOpenConst"` //максимальное количество пустых соединений с бд
	SetMaxIdleConst int    `toml:"db_SetMaxIdleConst"` //максимальное количество соединенияй с бд
	CrossTable      string `toml:"cross_table"`        //название таблицы cross
	RegionTable     string `toml:"region_table"`       //название таблицы region
	AccountTable    string `toml:"account_table"`      //название таблицы account
	DevicesTable    string `toml:"devices_table"`      //название таблицы devices
	StatusTable     string `toml:"status_table"`       //название таблицы status
	LogDeviceTable  string `toml:"logDevice_table"`    //название таблицы logDevice
	ChatTable       string `toml:"chat_table"`         //название таблицы chat
}

//Config структура с объявлением всех переменных config.toml файла
type Config struct {
	LoggerPath    string `toml:"logger_path"`    //путь до каталога с логами сервера
	ViewsPath     string `toml:"views_path"`     //путь до каталога views (содержит все ресурсы для отображения перекрестков)
	CachePath     string `toml:"cache_path"`     //путь до каталога cachefile
	ResourcePath  string `toml:"resourcePath"`   //путь до каталога frontend
	ServerIP      string `toml:"server_ip"`      //ip сервера / порт
	YaKey         string `toml:"ya_key"`         //ключ авторизации для яндекса
	TokenPassword string `toml:"token_password"` //ключ для шифрования токенов доступа
	//TCPConfig     tcpConnect.TCPConfig `toml:"tcpServer"`      //информация о tcp соединении с сервером устройств
	DBConfig DBConfig `toml:"database"` //информация о соединении с базой данных
}

func (dbConfig *DBConfig) GetDBurl() string {
	return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbConfig.Host, dbConfig.User, dbConfig.Name, dbConfig.Password)
}

//NewConfig создание конфига
func NewConfig() *Config {
	return &Config{}
}
