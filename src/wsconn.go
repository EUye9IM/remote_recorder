package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// TODO 监控端发送offer
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ConnData struct {
	wsconn        *websocket.Conn
	uinfo         Uinfo
	joined        bool
	stu_stream_id struct {
		screen string
		camera string
	}
	// stu_tracks struct {
	// 	track *webrtc.TrackRemote,
	// }
	close sync.Mutex
}

var conn_set = make(map[*ConnData]bool)

func WebsocketServer(c *gin.Context) {
	userdata := new(ConnData)
	userdata.close.Lock()
	userdata.joined = false
	log.Println("Websocket Connect")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Panicln("Websocket Error: " + err.Error())
	}
	defer ws.Close()
	defer func() {
		if !userdata.joined {
			return
		}
		updata := map[string]interface{}{
			"event": "MemberLeft",
			"no":    userdata.uinfo.No,
			"name":  userdata.uinfo.Name,
			"level": userdata.uinfo.Level,
		}
		boardcastEvent("1", nil, updata)
	}()
	defer log.Println("Websocket Close")
	conn_set[userdata] = true
	defer delete(conn_set, userdata)

	userdata.wsconn = ws

	var peerConnection *webrtc.PeerConnection
	peerConnection = nil
	defer func() {
		if peerConnection == nil {
			return
		}
		userdata.close.Unlock()
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
		if js.Action == "token" {
			var js struct {
				Action string `json:"action"`
				Data   string `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			if err != nil {
				log.Print(err)
				continue
			}
			for i := range conn_set {
				if i.uinfo.No == Users_info[js.Data].No {
					return
				}
			}

			log.Println("Websocket token: ", js.Data)
			userdata.uinfo = Users_info[js.Data]
			userdata.joined = true
			updata := map[string]interface{}{
				"event": "MemberJoined",
				"no":    userdata.uinfo.No,
				"name":  userdata.uinfo.Name,
				"level": userdata.uinfo.Level,
			}
			boardcastEvent("1", ws, updata)
			continue
		}

		if js.Action == "event" {
			if !userdata.joined {
				continue
			}
			var js struct {
				Action string                 `json:"action"`
				Data   map[string]interface{} `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			if err != nil {
				log.Print(err)
				continue
			}
			if js.Data["event"] == "GetMemberStream" {
				if userdata.uinfo.Level != "1" {
					continue
				}
				if peerConnection != nil {
					peerConnection.Close()
					peerConnection = nil
				}

				stu_no := js.Data["no"]
				s_id := ""
				c_id := ""
				for i := range conn_set {
					if i.uinfo.No == stu_no {
						s_id = i.stu_stream_id.screen
						c_id = i.stu_stream_id.camera
					}
				}
				updata := map[string]interface{}{
					"event": "SendStreamId",
					"streamid": map[string]interface{}{
						"screen": s_id,
						"camera": c_id,
					},
				}
				sendEvent(ws, updata)

				// TODO send offer
				peerConnection = newRemoteConnection(ws, userdata)

				offer := webrtc.SessionDescription{}
				//send answer back
				upload := map[string]interface{}{
					"action": "offer",
					"data":   offer,
				}
				ws.WriteJSON(upload)
				log.Println("Websocket write: offer")

				continue
			}
			continue
		}
		if js.Action == "streamid" {
			if !userdata.joined {
				continue
			}
			if userdata.uinfo.Level != "0" {
				continue
			}
			var js struct {
				Action string `json:"action"`
				Data   struct {
					Screen string `json:"screen"`
					Camera string `json:"camera"`
				} `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			if err != nil {
				log.Print(err)
				continue
			}

			userdata.stu_stream_id.screen = js.Data.Screen
			userdata.stu_stream_id.camera = js.Data.Camera

			continue
		}

		if js.Action == "offer" {
			if !userdata.joined {
				continue
			}
			if userdata.uinfo.Level != "0" {
				continue
			}
			if peerConnection != nil {
				peerConnection.Close()
				peerConnection = nil
			}
			var js struct {
				Action string                    `json:"action"`
				Data   webrtc.SessionDescription `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			offer := js.Data
			if err == nil && offer.Type == webrtc.SDPTypeOffer {
				peerConnection = newConnection(ws, userdata)

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
			if !userdata.joined {
				continue
			}
			// if userdata.uinfo.Level != "0" {
			// 	continue
			// }
			if peerConnection == nil {
				continue
			}
			var js struct {
				Action string                  `json:"action"`
				Data   webrtc.ICECandidateInit `json:"data"`
			}
			err = json.Unmarshal(content, &js)
			candidate := js.Data
			if err == nil {
				peerConnection.AddICECandidate(candidate)
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

func boardcastEvent(level string, exc *websocket.Conn, data map[string]interface{}) {
	upload := map[string]interface{}{
		"action": "event",
		"data":   data,
	}
	for k := range conn_set {
		if k.uinfo.Level == level && k.wsconn != exc {
			k.wsconn.WriteJSON(upload)
		}
	}
}
func sendEvent(to *websocket.Conn, data map[string]interface{}) {
	upload := map[string]interface{}{
		"action": "event",
		"data":   data,
	}
	to.WriteJSON(upload)
}
