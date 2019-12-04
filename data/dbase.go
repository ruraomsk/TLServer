package data

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
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

	// defer conn.Close()
	
	db = conn
	db.Table("accounts").AutoMigrate(Account{})
	return nil
}

func GetDB() *gorm.DB {
	return db
}
