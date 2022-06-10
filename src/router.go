package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Uinfo struct {
	No     string
	Name   string
	Level  byte
	Enable byte
}

var (
	engin *gin.Engine

	Users_info map[string]Uinfo
)

func init() {
	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engin = gin.New()
	if config.Debug {
		engin.Use(gin.Logger(), gin.Recovery())
	} else {
		engin.Use(gin.Recovery())
	}
	engin.SetTrustedProxies(nil)

	engin.Static("/static", config.App.Resource)
	engin.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/static")
	})
	api := engin.Group("api")
	apiRoute(api)
}

func apiRoute(r *gin.RouterGroup) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/ws", WebsocketServer)
}

func RunHttp() {
	log.Println("Server start with config:", config)
	if config.App.Https {
		engin.RunTLS(":"+strconv.Itoa(config.App.Port),
			config.App.Signature.Crt, config.App.Signature.Key)
	} else {
		engin.Run(":" + strconv.Itoa(config.App.Port))
	}
}
func StopHttp() {
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(config.App.Port),
		Handler: engin,
	}

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server forced to shutdown:", err)
	}

	log.Println("Server stop")
}
