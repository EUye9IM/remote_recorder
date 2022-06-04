package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

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

	engin = gin.New()
	if Cfg.Debug {
		engin.Use(gin.Logger(), gin.Recovery())
	} else {
		engin.Use(gin.Recovery())
	}
	engin.SetTrustedProxies(nil)

	engin.Static("/static", Cfg.App.Resource)
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

	initSocketio()
	r.GET("/sio/*any", gin.WrapH(Sio))
	r.POST("/sio/*any", gin.WrapH(Sio))
}

func RunHttp() {
	log.Println("Server start")
	if Cfg.App.Https {
		engin.RunTLS(":"+strconv.Itoa(Cfg.App.Port),
			Cfg.App.Signature.Crt, Cfg.App.Signature.Key)
	} else {
		engin.Run(":" + strconv.Itoa(Cfg.App.Port))
	}
}
func StopHttp() {
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(Cfg.App.Port),
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
