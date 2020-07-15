package data

import (
	"fmt"

	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/JanFant/TLServer/logger"
	_ "github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
)

var (
	chatTable = `
	CREATE TABLE chat (
		time timestamp with time zone PRIMARY KEY,
		fromU text,
		toU text,
		message text	
	)
	WITH (
		autovacuum_enabled = true		
	);`
	accountsTable = `
	CREATE TABLE accounts (
		id serial PRIMARY KEY,
		description text,
		login text,
		password text,
		work_time bigint,
		token text,
		privilege jsonb
	)
	WITH (
		autovacuum_enabled = true		
	);`
	routesTable = `
	CREATE TABLE routes (
		id serial PRIMARY KEY,
		description text,
		box jsonb,
		listTL jsonb
	)
	WITH (
		autovacuum_enabled = true		
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

	_, err = db.Exec(`SELECT * FROM public.routes;`)
	if err != nil {
		fmt.Println("routes table not found - created")
		logger.Info.Println("|Message: routes table not found - created")
		db.MustExec(routesTable)
	}
	return db, nil
}

//GetDB обращение к БД
func GetDB() *sqlx.DB {
	return db
}
