package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type conn_data struct {
	wsconn *websocket.Conn
	uinfo  Uinfo
}

var conn_set = make(map[*conn_data]bool)

func WebsocketServer(c *gin.Context) {
	data := new(conn_data)
	log.Println("Websocket Connect")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Panicln("Websocket Error: " + err.Error())
	}
	defer ws.Close()
	defer log.Println("Websocket Close")
	conn_set[data] = true
	defer delete(conn_set, data)

	data.wsconn = ws

	peerConnection := newConnection(ws)
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			log.Panicln("cannot close peerConnection: %v\n" + cErr.Error())
		}
	}()

	for {
		_, content, err := ws.ReadMessage()
		if err != nil {
			log.Println("Websocket Read Error: " + err.Error())
			break
		}
		var js struct {
			Action string      `json:"action"`
			Data   interface{} `json:"data"`
		}
		err = json.Unmarshal(content, &js)
		if err != nil {
			log.Println("receive not json: " + string(content))
			continue
		}
		// TODO change action name
		if js.Action == "streamid" {
			var js struct {
				Action string `json:"action"`
				Data   struct {
					// Token string `json:"token"`
					Screen string `json:"screen"`
					Camera string `json:"camera"`
				} `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			if err != nil {
				log.Print(err)
				continue
			}
			// data.uinfo = Users_info[js.Data.Token]
			continue
		}
		if js.Action == "offer" {
			var js struct {
				Action string                    `json:"action"`
				Data   webrtc.SessionDescription `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			offer := js.Data
			if err == nil && offer.Type == webrtc.SDPTypeOffer {
				log.Println("receive offer: " + offer.SDP)

				answer := connectionAnswer(peerConnection, offer)
				//send answer back
				upload := map[string]interface{}{
					"action": "answer",
					"data":   answer,
				}
				ws.WriteJSON(upload)
				log.Println("Websocket write: answer")
			} else {
				logUnknown(string(content))
			}
			continue
		}
		if js.Action == "candidate" {
			var js struct {
				Action string                  `json:"action"`
				Data   webrtc.ICECandidateInit `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			candidate := js.Data
			if err == nil {
				peerConnection.AddICECandidate(candidate)
				log.Println("Add ice candidate: " + string(content))
			} else {
				logUnknown(string(content))
			}
			continue
		}
		logUnknown(string(content))
	}
}
func logUnknown(content string) {
	log.Println("Websocket Read unknown: " + string(content))
}

func boardcastEvent(level string, data map[string]interface{}) {
	upload := map[string]interface{}{
		"action": "event",
		"data":   data,
	}
	for k, _ := range conn_set {
		if k.uinfo.Level == level {
			k.wsconn.WriteJSON(upload)
		}
	}
}
