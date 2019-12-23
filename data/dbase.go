package data

import (
	"../logger"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"os"
)

var db *gorm.DB

//ConnectDB connecting to DB
func ConnectDB() error {
	username := os.Getenv("db_user")
	password := os.Getenv("db_password")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")
	dbType := os.Getenv("db_type")

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)
	conn, err := gorm.Open(dbType, dbURI)
	if err != nil {
		return err
	}

	db = conn
	if !db.HasTable(Account{}) {
		logger.Info.Println("dbase: Didn't find the Accounts table, created it with SuperUser")
		if err = db.Table("accounts").AutoMigrate(Account{}).Error; err != nil {
			return err
		}
		//Добавляю в созданную таблицу 2 колонки с координатами начального поля
		if err = db.Table("accounts").Exec("alter table accounts add points0 point").Error; err != nil {
			return err
		}
		if err = db.Table("accounts").Exec("alter table accounts add points1 point").Error; err != nil {
			return err
		}
		if err = db.Table("accounts").Exec("alter table accounts add privilege jsonb").Error; err != nil {
			return err
		}
		// Супер пользователь
		acc := Account{}

		db.Table("accounts").Create(acc.SuperCreate())
		//Записываю координаты в базу!!!
		db.Exec(acc.BoxPoint.Point0.ToSqlString("accounts", "points0", acc.Login))
		db.Exec(acc.BoxPoint.Point1.ToSqlString("accounts", "points1", acc.Login))

	}

	return nil
}

func GetDB() *gorm.DB {
	return db
}
