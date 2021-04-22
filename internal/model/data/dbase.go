package data

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx"
	"github.com/jmoiron/sqlx"
	"github.com/ruraomsk/TLServer/internal/model/config"
	"github.com/ruraomsk/TLServer/logger"
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
		description text,
		login text PRIMARY KEY,
		password text,
		work_time bigint,
		token text,
		privilege jsonb
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

	//FirstCreate флаг первого создания базы
	FirstCreate bool
)

type usedDb struct {
	db   *sql.DB
	used bool
}
type usedDbx struct {
	db   *sqlx.DB
	used bool
}

var dbxPool []usedDbx
var dbPool []usedDb
var mutex sync.Mutex
var first = true

//ConnectDB подключение к БД
func ConnectDB() error {
	var err error
	if first {
		dbPool = make([]usedDb, 0)
		first = false
		for i := 0; i < config.GlobalConfig.DBConfig.SetMaxOpenConst; i++ {
			d := new(usedDb)
			d.used = false
			d.db, err = sql.Open(config.GlobalConfig.DBConfig.Type, config.GlobalConfig.DBConfig.GetDBurl())
			if err != nil {
				logger.Error.Printf("dbase ConnectDB %s", err.Error())
				return err
			}
			dbPool = append(dbPool, *d)
			dbxPool = append(dbxPool, usedDbx{db: nil, used: false})
		}
	}
	db, id := GetDBX()
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
	FreeDBX(id)
	return nil
}

//GetDB обращение к БД
func GetDB() (db *sql.DB, id int) {
	mutex.Lock()
	defer mutex.Unlock()
	for i, d := range dbPool {
		if !d.used {
			//logger.Info.Printf("Выдали %d", i)
			dbPool[i].db.SetMaxOpenConns(1)
			dbPool[i].used = true
			return dbPool[i].db, i
		}
	}
	logger.Error.Printf("dbase закончился пул соединений")
	return nil, 0
}
func FreeDB(id int) {
	mutex.Lock()
	defer mutex.Unlock()
	if id < 0 || id >= len(dbPool) {
		logger.Error.Printf("dbase freeDb неверный индекс %d", id)
		return
	}
	//logger.Info.Printf("Освободили %d", id)
	//dbPool[id].db.Close()
	dbPool[id].used = false
}

//GetDBX обращение к БД sqlx
func GetDBX() (db *sqlx.DB, id int) {
	mutex.Lock()
	defer mutex.Unlock()
	var err error
	for i, d := range dbxPool {
		if !d.used {
			//logger.Info.Printf("Выдали dbx %d", i)
			dbxPool[i].db, err = sqlx.Open(config.GlobalConfig.DBConfig.Type, config.GlobalConfig.DBConfig.GetDBurl())
			if err != nil {
				logger.Error.Printf("dbase getdb %s", err.Error())
				return nil, 0
			}
			dbxPool[i].db.SetMaxOpenConns(1)
			dbxPool[i].used = true
			return dbxPool[i].db, i
		}
	}
	logger.Error.Printf("dbasex закончился пул соединений")
	return nil, 0
}
func FreeDBX(id int) {
	mutex.Lock()
	defer mutex.Unlock()
	if id < 0 || id >= len(dbxPool) {
		logger.Error.Printf("dbase freeDbx неверный индекс %d", id)
		return
	}
	//logger.Info.Printf("Освободили x %d", id)
	dbxPool[id].db.Close()
	dbxPool[id].used = false
}
