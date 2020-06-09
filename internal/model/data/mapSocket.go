package data

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/internal/model/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"strings"
	"time"
)

var ConnectedMapUsers map[*websocket.Conn]bool
var WriteMap chan MapSokResponse

func MapReader(conn *websocket.Conn, c *gin.Context) {
	ConnectedMapUsers[conn] = true
	login := ""
	flag, mapContx := checkToken(c)

	{
		resp := mapSokMessage(typeMapInfo, conn, mapOpenInfo())

		if flag {
			login = mapContx["login"]
			resp.Data["manageFlag"], _ = AccessCheck(login, mapContx["role"], 1)
			resp.Data["logDeviceFlag"], _ = AccessCheck(login, mapContx["role"], 11)
			resp.Data["authorizedFlag"] = true
		}
		resp.send()
	}

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			delete(ConnectedMapUsers, conn)
			//закрытие коннекта
			return
		}

		typeSelect, err := setTypeMessage(p)
		if err != nil {
			resp := mapSokMessage(typeError, conn, nil)
			resp.Data["message"] = ErrorMessage{Error: errUnregisteredMessageType}
			resp.send()
		}
		switch typeSelect {
		case typeJump:
			{
				location := &Locations{}
				_ = json.Unmarshal(p, &location)
				box, _ := location.MakeBoxPoint()
				resp := mapSokMessage(typeJump, conn, nil)
				resp.Data["boxPoint"] = box
				resp.send()
			}
		case typeLogin:
			{
				account := &Account{}
				_ = json.Unmarshal(p, &account)
				resp := Login(account.Login, account.Password, conn.RemoteAddr().String())
				if resp.Type == typeLogin {
					login = fmt.Sprint(resp.Data["login"])
				}
				resp.conn = conn
				resp.send()
			}
		case typeLogOut:
			{
				if login != "" {
					resp := LogOut(login)
					resp.conn = conn
					resp.Data["authorizedFlag"] = true
					resp.send()
				}
			}

		}
	}
}

func MapBroadcast() {
	ConnectedMapUsers = make(map[*websocket.Conn]bool)
	WriteMap = make(chan MapSokResponse)
	crossReadTick := time.Tick(time.Second * 5)
	oldTFs := selectTL()
	for {
		select {
		case <-crossReadTick:
			{
				newTFs := selectTL()
				var tempTF []TrafficLights
				for _, nTF := range newTFs {
					for _, oTF := range oldTFs {
						if oTF.Idevice == nTF.Idevice && oTF.Sost.Num != nTF.Sost.Num {
							tempTF = append(tempTF, nTF)
							break
						}
					}
				}
				oldTFs = newTFs
				if len(ConnectedMapUsers) > 0 {
					if len(tempTF) > 0 {
						resp := mapSokMessage(typeTFlight, nil, nil)
						resp.Data["tflight"] = tempTF
						for conn := range ConnectedMapUsers {
							if err := conn.WriteJSON(resp); err != nil {
								_ = conn.Close()
							}
						}
					}
				}
			}
		case msg := <-WriteMap:
			{
				if err := msg.conn.WriteJSON(msg); err != nil {
					_ = msg.conn.Close()
				}
			}
		}
	}
}

//checkToken проверка токена для вебсокета
func checkToken(c *gin.Context) (flag bool, mapCont map[string]string) {
	var tokenString string
	cookie, err := c.Cookie("Authorization")
	//Проверка куков получили ли их вообще
	if err != nil {
		return false, nil
	}
	tokenString = cookie

	ip := strings.Split(c.Request.RemoteAddr, ":")
	//проверка если ли токен, если нету ошибка 403 нужно авторизироваться!
	if tokenString == "" {
		return false, nil
	}
	//токен приходит строкой в формате {слово пробел слово} разделяем строку и забираем нужную нам часть
	splitted := strings.Split(tokenString, " ")
	if len(splitted) != 2 {
		return false, nil
	}

	//берем часть где хранится токен
	tokenSTR := splitted[1]
	tk := &Token{}

	token, err := jwt.ParseWithClaims(tokenSTR, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GlobalConfig.TokenPassword), nil
	})

	//не правильный токен возвращаем ошибку с кодом 403
	if err != nil {
		return false, nil
	}

	//Проверка на уникальность токена
	var (
		userPrivilege  Privilege
		tokenStrFromBd string
	)
	rows, err := GetDB().Query(`SELECT token, privilege FROM public.accounts WHERE login = $1`, tk.Login)
	if err != nil {
		return false, nil
	}
	for rows.Next() {
		_ = rows.Scan(&tokenStrFromBd, &userPrivilege.PrivilegeStr)
	}

	if tokenSTR != tokenStrFromBd || tk.IP != ip[0] || !token.Valid {
		return false, nil
	}

	//проверка токен пришел от правильного URL
	mapCont = make(map[string]string)

	//проверка правильности роли для указанного пользователя
	_ = userPrivilege.ConvertToJson()
	if userPrivilege.Role.Name != tk.Role {
		return false, nil
	}

	mapCont["login"] = tk.Login
	mapCont["role"] = tk.Role

	return true, mapCont
}
