package data

import (
	"fmt"
	"github.com/JanFant/newTLServer/internal/model/config"
	"github.com/JanFant/newTLServer/internal/model/logger"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/jmoiron/sqlx"
)

var (
	chatTable = `
	CREATE TABLE chat (
		time timestamp with time zone PRIMARY KEY,
		fromU text,
		toU text,
		message text	
	);`
	accountsTable = `
	CREATE TABLE accounts (
		id serial PRIMARY KEY,
		login text,
		password text,
		work_time bigint,
		ya_map_key text,
		token text,
		privilege jsonb
	);`
	createFuncSQL = `Create or replace function convTo360(x double precision) returns double precision as $$
		begin
		if x < 0 then
		return x + 360;
		else
		return x;
		end if;
		end
		$$ language plpgsql;`

	db *sqlx.DB
	//FirstCreate флаг первого создания базы
	FirstCreate bool
)

//ConnectDB подключение к БД
func ConnectDB() (*sqlx.DB, error) {

	conn, err := sqlx.Open(config.GlobalConfig.DBConfig.Type, config.GlobalConfig.DBConfig.GetDBurl())
	if err != nil {
		return nil, err
	}

	db = conn
	db.SetMaxOpenConns(config.GlobalConfig.DBConfig.SetMaxOpenConst)
	db.SetMaxIdleConns(config.GlobalConfig.DBConfig.SetMaxIdleConst)

	_, err = db.Exec(`SELECT * FROM public.accounts;`)
	if err != nil {
		fmt.Println("accounts table not found - created")
		logger.Info.Println("|Message: accounts table not found - created")
		db.MustExec(accountsTable)
		db.MustExec(createFuncSQL)
		FirstCreate = true

	}

	_, err = db.Exec(`SELECT * FROM public.chat;`)
	if err != nil {
		fmt.Println("chat table not found - created")
		logger.Info.Println("|Message: chat table not found - created")
		db.MustExec(chatTable)
	}
	return db, nil
}

//GetDB обращение к БД
func GetDB() *sqlx.DB {
	return db
}
