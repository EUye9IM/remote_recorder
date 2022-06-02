package main

import (
	"github.com/gin-gonic/gin"
)

var (
	engin *gin.Engine
)

func init() {
	if Cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engin = gin.Default()
	engin.SetTrustedProxies(nil)

	engin.Static("/", Cfg.App.Resource)
}

func RunHttp() {
	engin.Run()
}
