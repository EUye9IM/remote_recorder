package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	// set log
	if !Cfg.Debug && Cfg.App.Log != "" {
		logFile, err := os.OpenFile(Cfg.App.Log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic("Cannot open file: " + Cfg.App.Log + ".\n" + err.Error())
		}
		log.SetOutput(logFile)
	}
	log.SetFlags(log.Lmicroseconds | log.Ldate)

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
