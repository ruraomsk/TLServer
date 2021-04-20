package serverLog

import (
	"fmt"
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/TLServer/internal/model/data"
	"io/ioutil"
	"net/http"
	"strings"

	u "github.com/ruraomsk/TLServer/internal/utils"
	"github.com/ruraomsk/TLServer/logger"
)

//ServerLogInfo данные для хранения информации лог файлов
type ServerLogInfo struct {
	Type     string `json:"type"`     //тип лог сообщения
	Time     string `json:"time"`     //врямя когда произошло событие
	IP       string `json:"IP"`       //IP с которого делали запрос
	Login    string `json:"login"`    //логин пользователя который делал запрос
	Resource string `json:"resource"` //путь к ресурсу на котором произошло событие
	Message  string `json:"message"`  //расшифровка действия пользователя
}

var logFileSuffix = ".log"

//DisplayServerLogFiles отображения всех лог файлов в каталоге
func DisplayServerLogFiles(logPath string) u.Response {
	files, err := ioutil.ReadDir(logPath)
	if err != nil {
		logger.Error.Println("|Message: Error reading directory with log files")
		resp := u.Message(http.StatusInternalServerError, "Log dir can't open")
		return resp
	}
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, strings.TrimSuffix(file.Name(), ".log"))
	}
	resp := u.Message(http.StatusOK, "display a list of log files")
	resp.Obj["fileNames"] = fileNames
	return resp
}

//DisplayServerFileLog получение данных из заданного файла
func DisplayServerFileLog(fileName, logPath string, accInfo *accToken.Token) u.Response {
	db, id := data.GetDB()
	defer data.FreeDB(id)
	path := logPath + "//" + fileName + logFileSuffix
	byteFile, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Println("|Message: Error reading directory with log files")
		resp := u.Message(http.StatusInternalServerError, "Log files not found")
		return resp
	}
	var loginNames []string
	if accInfo.Region != "*" {
		sqlStr := fmt.Sprintf(`SELECT login FROM public.accounts WHERE privilege::jsonb @> '{"region":"%s"}'::jsonb`, accInfo.Region)
		rowsTL, err := db.Query(sqlStr)
		if err != nil {
			return u.Message(http.StatusBadRequest, "display info: Bad request")
		}
		for rowsTL.Next() {
			var login string
			err := rowsTL.Scan(&login)
			if err != nil {
				return u.Message(http.StatusBadRequest, "display info: Bad request")
			}
			loginNames = append(loginNames, login)
		}
	}

	logData := logParser(string(byteFile), loginNames)
	resp := u.Message(http.StatusOK, fmt.Sprintf("display info from file: %v", fileName))
	resp.Obj["logData"] = logData
	return resp
}

//logParser разборщик лог файла
func logParser(rawData string, loginNames []string) (logData []ServerLogInfo) {
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
		var tempLogData = ServerLogInfo{}
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
