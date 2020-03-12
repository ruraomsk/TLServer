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
	Name            string `toml:"db_name"`
	Password        string `toml:"db_password"`
	User            string `toml:"db_user"`
	Type            string `toml:"db_type"`
	Host            string `toml:"db_host"`
	Port            string `toml:"db_port"`
	SetMaxOpenConst int    `toml:"db_SetMaxOpenConst"`
	SetMaxIdleConst int    `toml:"db_SetMaxIdleConst"`
	GisTable        string `toml:"gis_table"`
	RegionTable     string `toml:"region_table"`
	AccountTable    string `toml:"account_table"`
	DevicesTable    string `toml:"devices_table"`
}

func (dbConfig *DBConfig) getDBurl() string {
	return fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbConfig.Host, dbConfig.User, dbConfig.Name, dbConfig.Password)
}

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
	return nil
}

//GetDB обращение к БД
func GetDB() *gorm.DB {
	return db
}
