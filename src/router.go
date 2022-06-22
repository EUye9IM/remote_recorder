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
	Level  string
	Enable bool
}

var (
	engin *gin.Engine
)

func init() {
	if config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	token_maker.Set("0123456789abcdef", 32)
	Users_info = make(map[string]Uinfo)

	engin = gin.New()
	if config.Debug {
		engin.Use(gin.Logger(), gin.Recovery())
	} else {
		engin.Use(gin.Recovery())
	}
	engin.SetTrustedProxies(nil)

	engin.LoadHTMLGlob(config.App.Resource + "/templates/*")
	engin.Static("/static", config.App.Resource+"/static")

	// engin.GET("/", func(c *gin.Context) {
	// 	if _, err := c.Cookie("token"); err == nil {

	// 	} else {
	// 		c.Redirect(http.StatusMovedPermanently, "/login")
	// 	}
	// })
	engin.GET("/", func(ctx *gin.Context) {
		v, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		uinfo, ok := Users_info[v]
		if !ok {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		if !uinfo.Enable {
			ctx.Redirect(http.StatusMovedPermanently, "/passchange")
			return
		}
		ctx.Redirect(http.StatusMovedPermanently, "/index")
	})
	engin.GET("/demo", func(ctx *gin.Context) { handleHtml(ctx, "demo.html") })
	engin.GET("/index", func(ctx *gin.Context) {
		v, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		_, ok := Users_info[v]
		if !ok {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		handleHtml(ctx, "index.html")
	})
	engin.GET("/lobby", func(ctx *gin.Context) { handleHtml(ctx, "lobby.html") })
	engin.GET("/login", func(ctx *gin.Context) {
		v, err := ctx.Cookie("token")
		if err == nil {
			if _, ok := Users_info[v]; ok {
				ctx.Redirect(http.StatusMovedPermanently, "/")
				return
			}
		}
		handleHtml(ctx, "login.html")
	})
	engin.GET("/passchange", func(ctx *gin.Context) {
		v, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		_, ok := Users_info[v]
		if !ok {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		handleHtml(ctx, "passchange.html")
	})
	engin.GET("/remote", func(ctx *gin.Context) {
		v, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		_, ok := Users_info[v]
		if !ok {
			ctx.Redirect(http.StatusMovedPermanently, "/login")
			return
		}
		handleHtml(ctx, "remote.html")
	})

	api := engin.Group("api")
	apiRoute(api)
}

func apiRoute(r *gin.RouterGroup) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.POST("/login", handleLogin)
	r.POST("/uinfo", handleUinfo)
	r.POST("/logout", handleLogout)
	r.POST("/chpw", handleChpw)
	r.POST("/resetpw", handleResetpw)
	r.POST("/getmembers", handleGetmembers)

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
