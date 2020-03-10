package data

import (
	"fmt"
	"github.com/JanFant/TLServer/logger"
	u "github.com/JanFant/TLServer/utils"
	"io/ioutil"
	"strings"
)

//LogInfo данные для хранения информации лог файлов
type LogInfo struct {
	Type     string `json:"type"`     //тип лог сообщения
	Time     string `json:"time"`     //врямя когда произошло событие
	IP       string `json:"IP"`       //IP с которого делали запрос
	Login    string `json:"login"`    //логин пользователя который делал запрос
	Resource string `json:"resource"` //путь к ресурсу на котором произошло событие
	Message  string `json:"message"`  //расшифровка действия пользователя
}

var logFileSuffix = ".log"

//DisplayLogFiles отображения всех лог файлов в каталоге
func DisplayLogFiles() map[string]interface{} {
	files, err := ioutil.ReadDir(GlobalConfig.LoggerPath)
	if err != nil {
		logger.Error.Println("|Message: Error reading directory with log files")
		resp := u.Message(false, "Log dir can't open")
		return resp
	}
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, strings.TrimSuffix(file.Name(), ".log"))
	}
	resp := u.Message(true, "Display a list of log files")
	resp["fileNames"] = fileNames
	return resp
}

//DisplayFileLog получение данных из заданного файла
func DisplayFileLog(fileName string, mapContex map[string]string) map[string]interface{} {
	path := GlobalConfig.LoggerPath + "//" + fileName + logFileSuffix
	byteFile, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println("|Message: Error reading directory with log files")
		resp := u.Message(false, "Log files not found")
		return resp
	}
	var loginNames []string
	if mapContex["region"] != "*" {
		sqlStr := fmt.Sprintf(`select login from public.accounts where privilege::jsonb @> '{"region":"%s"}'::jsonb`, mapContex["region"])
		rowsTL, _ := GetDB().Raw(sqlStr).Rows()
		for rowsTL.Next() {
			var login string
			err := rowsTL.Scan(&login)
			if err != nil {
				return u.Message(false, "Display info: Bad request")
			}
			loginNames = append(loginNames, login)
		}
	}

	logData := logParser(string(byteFile), loginNames)
	resp := u.Message(true, fmt.Sprintf("Display info from file: %v", fileName))
	resp["logData"] = logData
	return resp
}

//logParser разборщик лог файла
func logParser(rawData string, loginNames []string) (logData []LogInfo) {
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
		if compareLoginNames(loginNames, tempLogData.Login) {
			logData = append(logData, tempLogData)
		}
	}
	return logData
}

//compareLoginNames решение по добавлению записи в ответ
func compareLoginNames(loginNames []string, login string) bool {
	if len(loginNames) == 0 {
		return true
	}
	for _, name := range loginNames {
		if name == login {
			return true
		}
	}
	return false
}

//logStrPars разбор строки с удалением лишних пробелов в начале и конце
func logStrPars(sep, line string) string {
	start := strings.Index(line, sep) + len(sep)
	line = line[start:]
	line = strings.TrimPrefix(line, " ")
	line = strings.TrimSuffix(line, " ")
	return line
}
