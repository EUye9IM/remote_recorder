package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var conn_set = make(map[*websocket.Conn]bool)

func WebsocketServer(c *gin.Context) {
	log.Println("Websocket Connect")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Panicln("Websocket Error: " + err.Error())
	}
	defer conn.Close()
	defer log.Println("Websocket Close")
	defer delete(conn_set, conn)

	conn_set[conn] = true
	for {
		msg_type, content, err := conn.ReadMessage()
		if err != nil {
			log.Println("Websocket Read Error: " + err.Error())
			break
		}
		log.Println("Websocket Read: " + string(content))
		for connect := range conn_set {
			err = connect.WriteMessage(msg_type, content)
			if err != nil {
				log.Println("Websocket Write Error: " + err.Error())
				break
			}
		}
	}
}
