package logger

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

//LogFile информация о лог файле
type LogFile struct {
	flog  *os.File //дескриптор открытого файла
	mutex sync.Mutex
	path  string //путь до текущего файла лога
	date  string //текущая дата
}

//logfile текущий лог файл
var logfile *LogFile
var (
	//Debug send debug message
	Debug *log.Logger
	//Trace send trace message
	Trace *log.Logger
	//Info Send info message
	Info *log.Logger
	//Warning send warning message
	Warning *log.Logger
	//Error  send error message
	Error *log.Logger
)

//Init запуск системы логирования
func Init(path string) (err error) {
	logfile, err = logOpen(path)
	if err != nil {
		return err
	}
	Debug = log.New(logfile,
		"DEBUG: TIME: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Trace = log.New(logfile,
		"TRACE: TIME: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(logfile,
		"INFO: TIME: ",
		log.Ldate|log.Ltime)

	Warning = log.New(logfile,
		"WARNING: TIME: ",
		log.Ldate|log.Ltime)

	Error = log.New(logfile,
		"ERROR: TIME: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

//LogOpen открытие/создание файла хранящего лог информацию
func logOpen(path string) (log *LogFile, err error) {
	go logClean(path)
	log = new(LogFile)
	log.date = time.Now().Format(time.RFC3339)[0:10]
	log.path = path
	path += "/log" + log.date + ".log"
	log.flog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return
}

//logClean проверка и удаление старых лог файлов
func logClean(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		Error.Println("|Message: Error reading directory with log files (logClean)")
	}
	for {
		for _, file := range files {
			if file.ModTime().Add(time.Hour * 24 * 30).Before(time.Now()) {
				_ = os.Remove(path + "/" + file.Name())
			}
		}
		time.Sleep(time.Hour * 24)
	}
}

//Read чтение лог файла
func (l *LogFile) Read(p []byte) (n int, err error) {
	n, err = l.flog.Read(p)
	return
}

//Write записать лог файла
func (l *LogFile) Write(p []byte) (n int, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	date := time.Now().Format(time.RFC3339)[0:10]
	n = 0
	if strings.Compare(l.date, date) != 0 {
		l.flog.Close()
		l.date = date
		path := l.path + "/log" + l.date + ".log"
		l.flog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return
		}
	}
	n, err = l.flog.Write(p)
	return
}

//Close закрытие лог файла
func (l *LogFile) Close() error {
	err := l.flog.Close()
	return err
}
