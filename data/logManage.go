package data

import (
	"../logger"
	u "../utils"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type LogInfo struct {
	Type string `json:"type"`
	//Time     time.Duration `json:"time"`
	Time     string `json:"time"`
	IP       string `json:"IP"`
	Login    string `json:"login"`
	Resource string `json:"resource"`
	Message  string `json:"message"`
}

var logFileSuffix = ".log"

func DisplayLogFiles() map[string]interface{} {
	files, err := ioutil.ReadDir(os.Getenv("logger_path"))
	if err != nil {
		logger.Error.Println("Error reading directory with log files")
		resp := u.Message(false, "Log dir can't open")
		return resp
	}
	var nameFile []string
	for _, file := range files {
		nameFile = append(nameFile, strings.TrimSuffix(file.Name(), ".log"))
	}
	resp := u.Message(true, "Display a list of log files")
	resp["filesName"] = nameFile
	return resp
}

func DisplayFileLog(fileName string) map[string]interface{} {
	path := os.Getenv("logger_path") + "//" + fileName + logFileSuffix
	byteFile, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println("Error reading directory with log files")
		resp := u.Message(false, "Log files not found")
		return resp
	}
	logData := logParser(string(byteFile))
	resp := u.Message(true, fmt.Sprintf("Display info from file: %v", fileName))
	resp["logData"] = logData
	return resp
}

func logParser(rawData string) (logData []LogInfo) {
	var (
		Type     = ":"
		Time     = "TIME:"
		IP       = "IP:"
		Login    = "Login:"
		Resource = "Resource:"
		Message  = "Message:"
	)
	for _, line := range strings.Split(rawData, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var tempLogData = LogInfo{}
		//Type
		start := 0
		finish := strings.Index(line, Type)
		tempLogData.Type = logStrPars(start, finish, line)

		//Time //!!!!!!бля
		start = newStart(start, finish, strings.Index(line, IP), Time)
		finish = strings.Index(line, IP)
		tempLogData.Time = logStrPars(start, finish, line)

		//IP
		start = newStart(start, finish, strings.Index(line, Login), IP)
		finish = strings.Index(line, Login)
		tempLogData.IP = logStrPars(start, finish, line)

		//Login
		start = newStart(start, finish, strings.Index(line, Resource), Login)
		finish = strings.Index(line, Resource)
		tempLogData.Login = logStrPars(start, finish, line)

		//Resource
		start = newStart(start, finish, strings.Index(line, Message), Resource)
		finish = strings.Index(line, Message)
		tempLogData.Resource = logStrPars(start, finish, line)

		//Message
		start = newStart(start, finish, len(line)-1, Message)
		finish = len(line) - 1
		tempLogData.Message = logStrPars(start, finish, line)

		logData = append(logData, tempLogData)
	}
	return logData
}

func newStart(start, oldFin, newFin int, str string) int {
	if newFin > 0 {
		return oldFin + len(str)
	} else {
		return start
	}
}

func logStrPars(start, finish int, str string) string {
	if finish < 0 {
		return "-"
	}
	str = str[start:finish]
	str = strings.TrimPrefix(str, " ")
	str = strings.TrimSuffix(str, " ")
	return str
}
