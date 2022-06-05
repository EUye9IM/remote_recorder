package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
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

	// 测试事件
	Sio.OnEvent("/", "echo", func(c socketio.Conn, msg string) {
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

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Panicln("NewPeerConnection error: " + err.Error())
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			log.Panicln("cannot close peerConnection: %v\n" + cErr.Error())
		}
	}()

	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		upload := map[string]interface{}{
			"opt":  "candidate",
			"data": i,
		}
		conn.WriteJSON(upload)
		log.Println("Websocket write: candidate")
	})

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			log.Println("Peer Connection has gone to failed exiting")
		}
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			log.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

			// for range time.NewTicker(5 * time.Second).C {
			// 	message := signal.RandSeq(15)
			// 	fmt.Printf("Sending '%s'\n", message)

			// 	// Send the message as text
			// 	sendErr := d.SendText(message)
			// 	if sendErr != nil {
			// 		panic(sendErr)
			// 	}
			// }
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
			d.SendText("Hi, " + d.Label() + ". You just say: '" + string(msg.Data) + "'.")
		})
	})

	conn_set[conn] = true

	for {
		_, content, err := conn.ReadMessage()
		if err != nil {
			log.Println("Websocket Read Error: " + err.Error())
			break
		}
		var js struct {
			Opt  string      `json:"opt"`
			Data interface{} `json:"data"`
		}
		err = json.Unmarshal(content, &js)
		if err != nil {
			log.Println("receive not json: " + string(content))
		}
		if js.Opt == "offer" {
			var js struct {
				Opt  string                    `json:"opt"`
				Data webrtc.SessionDescription `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			offer := js.Data
			if err == nil && offer.Type == webrtc.SDPTypeOffer {
				log.Println("receive offer: " + offer.SDP)

				// Set the remote SessionDescription
				err = peerConnection.SetRemoteDescription(offer)
				if err != nil {
					log.Panicln(err)
				}

				// Create an answer
				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					log.Panicln(err)
				}

				// Sets the LocalDescription, and starts our UDP listeners
				err = peerConnection.SetLocalDescription(answer)
				if err != nil {
					log.Panicln(err)
				}

				//send answer back
				upload := map[string]interface{}{
					"opt":  "answer",
					"data": answer,
				}
				conn.WriteJSON(upload)
				log.Println("Websocket write: answer")
			}
			continue
		}
		if js.Opt == "candidate" {
			var js struct {
				Opt  string                  `json:"opt"`
				Data webrtc.ICECandidateInit `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			candidate := js.Data
			if err == nil {
				peerConnection.AddICECandidate(candidate)
				log.Println("Add ice candidate: " + string(content))
				continue
			}
			continue
		}

		log.Println("Websocket Read unknown: " + string(content))
	}
}
