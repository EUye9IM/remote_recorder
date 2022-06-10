package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// set log
	if !config.Debug && config.App.Log != "" {
		logFile, err := os.OpenFile(config.App.Log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("Cannot open file: " + config.App.Log + ".\n" + err.Error())
		}
		log.SetOutput(logFile)
		log.SetFlags(log.Lmicroseconds | log.Ldate)
	} else {
		log.SetFlags(log.Lmicroseconds | log.Ldate | log.Lshortfile)
	}

	// run server
	go RunHttp()

	// kill signal
	quit := make(chan os.Signal, 10)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	StopHttp()
}
