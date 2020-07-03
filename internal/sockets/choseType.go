package sockets

import (
	"encoding/json"
	"fmt"
)

//ChoseTypeMessage определение типа сообщения
func ChoseTypeMessage(raw []byte) (string, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(raw, &temp); err != nil {
		return "", err
	}
	return fmt.Sprint(temp["type"]), nil
}
