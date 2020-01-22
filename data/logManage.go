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
		logger.Error.Println("|Message: Error reading directory with log files")
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
		logger.Error.Println("|Message: Error reading directory with log files")
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

		splitLines := strings.Split(line, " |")

		for num, spLine := range splitLines {
			if num == 0 {
				tempLogData.Type = spLine[0:strings.Index(spLine, Type)]
				tempLogData.Time = logStrPars(Time, spLine)
			}
			if strings.Index(spLine, IP) >= 0 {
				tempLogData.IP = logStrPars(IP, spLine)
			}
			if strings.Index(spLine, Login) >= 0 {
				tempLogData.Login = logStrPars(Login, spLine)
			}
			if strings.Index(spLine, Resource) >= 0 {
				tempLogData.Resource = logStrPars(Resource, spLine)
			}
			if strings.Index(spLine, Message) >= 0 {
				tempLogData.Message = logStrPars(Message, spLine)
			}
		}
		logData = append(logData, tempLogData)
	}
	return logData
}

func logStrPars(sep, line string) string {
	start := strings.Index(line, sep) + len(sep)
	line = line[start:]
	line = strings.TrimPrefix(line, " ")
	line = strings.TrimSuffix(line, " ")
	return line
}
