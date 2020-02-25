package data

import (
	"fmt"
	"os"
	"strconv"

	"../logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	db          *gorm.DB
	FirstCreate bool
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
		username             = os.Getenv("db_user")
		password             = os.Getenv("db_password")
		dbName               = os.Getenv("db_name")
		dbHost               = os.Getenv("db_host")
		dbType               = os.Getenv("db_type")
		dbURI                = fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)
		dbSetMaxOpenConst, _ = strconv.Atoi(os.Getenv("db_SetMaxOpenConst"))
		dbSetMaxIdleConst, _ = strconv.Atoi(os.Getenv("db_SetMaxIdleConst"))
	)

	conn, err := gorm.Open(dbType, dbURI)
	if err != nil {
		return err
	}

	db = conn
	db.DB().SetMaxOpenConns(dbSetMaxOpenConst)
	db.DB().SetMaxIdleConns(dbSetMaxIdleConst)
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
