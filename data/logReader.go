package data

import (
	"time"
)

type Log struct {
	Type     string        `json:"type"`
	Time     time.Duration `json:"time"`
	IP       string        `json:"IP"`
	Login    string        `json:"login"`
	Resource string        `json:"resource"`
	Message  string        `json:"message"`
}
