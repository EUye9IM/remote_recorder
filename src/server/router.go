package main

import (
	"strconv"

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
	if Cfg.App.Https {
		engin.RunTLS(":"+strconv.Itoa(Cfg.App.Port), Cfg.App.Sign_path.Crt, Cfg.App.Sign_path.Key)
	} else {
		engin.Run(":" + strconv.Itoa(Cfg.App.Port))
	}
}
