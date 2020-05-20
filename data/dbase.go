package data

import (
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	db *gorm.DB
	//FirstCreate флаг первого создания базы
	FirstCreate bool
)

//DBConfig настройки для работы с базой данных
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
}

func (dbConfig *DBConfig) getDBurl() string {
	return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbConfig.Host, dbConfig.User, dbConfig.Name, dbConfig.Password)
}

var (
	ChatTable = `
	CREATE TABLE chat (
		time timestamp with time zone PRIMARY KEY,
		fromU text,
		toU text,
		message text	
	);`
)

//ConnectDB подключение к БД
func ConnectDB() error {
	var (
		CreateFuncSQL = `Create or replace function convTo360(x double precision) returns double precision as $$
		begin
		if x < 0 then
		return x + 360;
		else
		return x;
		end if;
		end
		$$ language plpgsql;`
	)

	conn, err := gorm.Open(GlobalConfig.DBConfig.Type, GlobalConfig.DBConfig.getDBurl())
	if err != nil {
		return err
	}

	db = conn
	db.DB().SetMaxOpenConns(GlobalConfig.DBConfig.SetMaxOpenConst)
	db.DB().SetMaxIdleConns(GlobalConfig.DBConfig.SetMaxIdleConst)
	if !db.HasTable(Account{}) {
		FirstCreate = true
		logger.Info.Println("|Message: DBase: Didn't find the Accounts table, created it with SuperUser")
		if err = db.Table("accounts").AutoMigrate(Account{}).Error; err != nil {
			return err
		}
		if err = db.Exec(CreateFuncSQL).Error; err != nil {
			return err
		}
		if err = db.Table("accounts").Exec("alter table accounts add privilege jsonb").Error; err != nil {
			return err
		}

	}

	_, err = db.DB().Query(`SELECT * FROM public.chat;`)
	if err != nil {
		db.DB().Exec(ChatTable)
		fmt.Println("создали таблицу для чата")
	}

	return nil
}

//GetDB обращение к БД
func GetDB() *gorm.DB {
	return db
}
