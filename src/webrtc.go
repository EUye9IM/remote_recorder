package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var conn_set = make(map[*websocket.Conn]bool)
var Sio *socketio.Server

func initSocketio() {
	Sio = socketio.NewServer(nil)

	Sio.OnConnect("/", func(c socketio.Conn) (err error) {
		Sio.BroadcastToNamespace("/", "MemberJoined", c.ID())
		log.Println("societio connected: " + c.ID())
		return
	})
	Sio.OnDisconnect("/", func(c socketio.Conn, reason string) {
		Sio.BroadcastToNamespace("/", "MemberLeft", c.ID())
		log.Println("societio disconnected: " + c.ID() + " " + reason)
	})

	Sio.OnEvent("/", "MessageToPeer", func(c socketio.Conn, msg string) {
		Sio.BroadcastToNamespace("/", "MessageFromPeer", c.ID())
		log.Println("societio rev msg: " + msg + " from: " + c.ID())
	})
}
func RunSocketio() {
	if err := Sio.Serve(); err != nil {
		log.Fatalf("socketio listen error: %s\n", err)
	}
}
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
