package main

import (
	"time"
)

func GetTime() string {
	return time.Now().Local().Format("2006-01-02_15-04-05")
}
