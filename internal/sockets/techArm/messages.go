package techArm

import (
	"github.com/ruraomsk/TLServer/internal/model/accToken"
	"github.com/ruraomsk/ag-server/pudge"
)

var (
	typeError   = "error"
	typeClose   = "close"
	typeArmInfo = "armInfo"
	typeDButton = "dispatch"
	typeGPRS    = "gprs"
	typeCrosses = "crosses"
	typeDevices = "devices"

	errParseType = "Сервер не смог обработать запрос"

	//modeRDK мапа состояний ДК
	modeRDK = map[int]string{
		1: "РУ",
		2: "РУ",
		3: "ЗУ",
		4: "ДУ",
		5: "ДУ",
		6: "ЛУ",
		8: "ЛУ",
		9: "КУ",
	}
	texMode = map[int]string{
		1:  "выбор ПК по времени по суточной карте ВР-СК",
		2:  "выбор ПК по недельной карте ВР-НК",
		3:  "выбор ПК по времени по суточной карте, назначенной оператором ДУ-СК",
		4:  "выбор ПК по недельной карте, назначенной оператором ДУ-НК",
		5:  "план по запросу оператора ДУ-ПК",
		6:  "резервный план (отсутствие точного времени) РП",
		7:  "коррекция привязки с ИП",
		8:  "коррекция привязки с сервера",
		9:  "выбор ПК по годовой карте",
		10: "выбор ПК по ХТ",
		11: "выбор ПК по картограмме",
		12: "противозаторовое управление",
	}
	GPRSInfo = struct {
		IP   string `json:"ip" ,toml:"tcpServerAddress"`
		Port string `json:"port" ,toml:"portGPRS"`
		Send bool   `json:"send" ,toml:"sendGPRS"`
	}{}
)

//armResponse структура для отправки сообщений (map)
type armResponse struct {
	Type string                 `json:"type"` //тип сообщения
	Data map[string]interface{} `json:"data"` //данные
}

//newMapMess создание нового сообщения
func newArmMess(mType string, data map[string]interface{}) armResponse {
	var resp armResponse
	resp.Type = mType
	if data != nil {
		resp.Data = data
	} else {
		resp.Data = make(map[string]interface{})
	}
	return resp
}

//ErrorMessage структура ошибки
type ErrorMessage struct {
	Error string `json:"error"`
}

//ArmInfo информация о запрашиваемом арме
type ArmInfo struct {
	Region  int      `json:"region"` //регион запрошенные пользователем
	Area    []string `json:"area"`   //район запрошенные пользователем
	AccInfo *accToken.Token
}

//CrossInfo информация о перекрестке для техАРМ
type CrossInfo struct {
	Region    int         `json:"region"`    //регион
	Area      int         `json:"area"`      //район
	ID        int         `json:"id"`        //id
	Idevice   int         `json:"idevice"`   //идентификатор устройства
	Subarea   int         `json:"subarea"`   //подрайон
	ArrayType int         `json:"arrayType"` //тип устройства
	Describe  string      `json:"describe"`  //описание
	Phone     string      `json:"phone"`     //телефон
	Model     pudge.Model `json:"Model"`     //модель устройства

}

//DevInfo информация о устройства для техАРМ
type DevInfo struct {
	Region   int              `json:"region"`   //регион
	Area     int              `json:"area"`     //район
	Idevice  int              `json:"idevice"`  //идентификатор устройства
	TechMode string           `json:"techMode"` //тех мод
	ModeRdk  string           `json:"modeRdk"`  //мод РДК
	Device   pudge.Controller `json:"device"`   //контроллер...
}
