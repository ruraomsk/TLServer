package logger

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

//LogFile функция
type LogFile struct {
	flog  *os.File
	mutex sync.Mutex
	path  string
	date  string
}

//var level int
var logfile *LogFile
var (
	//Debug send debug message
	Debug *log.Logger
	//Trace send trace message
	Trace *log.Logger
	//Info Send info message
	Info *log.Logger
	//Warning send warnig message
	Warning *log.Logger
	//Error  send error message
	Error *log.Logger
)

// Init init subsystem
func Init(path string) (err error) {
	logfile, err = logOpen(path)
	if err != nil {
		return err
	}
	Trace = log.New(logfile,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Trace = log.New(logfile,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(logfile,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(logfile,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(logfile,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

//LogOpen функция
func logOpen(path string) (log *LogFile, err error) {
	log = new(LogFile)
	log.date = time.Now().Format(time.RFC3339)[0:10]
	log.path = path
	path += "/log" + log.date + ".log"

	log.flog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	return
}
func (l *LogFile) Read(p []byte) (n int, err error) {
	n, err = l.flog.Read(p)
	return
}

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

//Close закрытие
func (l *LogFile) Close() error {
	err := l.flog.Close()
	return err
}
