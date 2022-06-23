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
	uuid  string
	close sync.Mutex
}
type ConnDataPair struct {
	s *ConnData
	t *ConnData
}

var conn_set = make(map[*ConnData]bool)
var uuid_map = make(map[string]ConnDataPair)

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
		if userdata.uuid == "" {
			delete(uuid_map, userdata.uuid)
			log.Println("uuid_map delete", userdata.uuid)
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
		var js map[string]interface{}
		err = json.Unmarshal(content, &js)
		if err != nil {
			log.Println("Websocket receive not json: " + string(content))
			continue
		}
		log.Println("Websocket receive: ", string(content))
		if js["action"] == "token" {
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

		if js["action"] == "event" {
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
				var sconn *ConnData
				for i := range conn_set {
					if i.uinfo.No == stu_no {
						s_id = i.stu_stream_id.screen
						c_id = i.stu_stream_id.camera
						// sconn应该在块内
						sconn = i
						break
					}

				}
				uuid := GetTmpID()
				updata := map[string]interface{}{
					"event": "SendStreamId",
					"streamid": map[string]interface{}{
						"screen": s_id,
						"camera": c_id,
					},
				}
				sendEvent(ws, updata)
				uuid_map[uuid] = ConnDataPair{s: sconn, t: userdata}
				log.Println("uuid_map add", uuid)

				// 此处需要发送给 student
				sendUuid(uuid_map[uuid].s.wsconn, uuid)
				// sendUuid(uuid_map[uuid].t.wsconn, uuid)

				continue
			}
			continue
		}
		if js["action"] == "streamid" {
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
				Uuid string `json:"uuid"`
			}
			err = json.Unmarshal(content, &js)
			if err != nil {
				log.Print(err)
				continue
			}

			userdata.stu_stream_id.screen = js.Data.Screen
			userdata.stu_stream_id.camera = js.Data.Camera
			// uuid=serveruuid
			uuid_map[js.Uuid] = ConnDataPair{s: userdata, t: nil}
			userdata.uuid = js.Uuid
			log.Println("uuid_map add", userdata.uuid)

			continue
		}

		if js["action"] == "offer" {
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
				Uuid   string                    `json:"uuid"`
			}
			err = json.Unmarshal(content, &js)
			offer := js.Data
			// check uuid=serveruuid
			if err == nil && offer.Type == webrtc.SDPTypeOffer {
				if uuid_map[js.Uuid].t == nil {
					peerConnection = newConnection(ws, userdata)

					answer := connectionAnswer(peerConnection, offer)
					//send answer back
					upload := map[string]interface{}{
						"action": "answer",
						"data":   answer,
						"uuid":   userdata.uuid,
					}
					wsSend(ws, upload)
					log.Println("Websocket write: answer")
				} else {
					// 此处需要发给对端，也就是 teacher 端
					if uuid_map[js.Uuid].t != nil {
						wsSend(uuid_map[js.Uuid].t.wsconn, js)
					}
				}
				continue
			} else {
				logUnknown(string(content))
			}
		}
		if js["action"] == "candidate" {
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
				Uuid   string                  `json:"uuid"`
			}
			err = json.Unmarshal(content, &js)
			//  check uuid=serveruuid
			candidate := js.Data
			if err == nil {
				if uuid_map[js.Uuid].t == nil {
					peerConnection.AddICECandidate(candidate)
				} else {
					if uuid_map[js.Uuid].s != nil && uuid_map[js.Uuid].t == userdata {
						wsSend(uuid_map[js.Uuid].s.wsconn, js)
						continue
					}
					if uuid_map[js.Uuid].t != nil && uuid_map[js.Uuid].s == userdata {
						wsSend(uuid_map[js.Uuid].t.wsconn, js)
						continue
					}
				}
			} else {
				logUnknown(string(content))
			}
			continue
		}
		if js["action"] == "answer" {
			if !userdata.joined {
				continue
			}
			if userdata.uinfo.Level != "1" {
				continue
			}
			// if peerConnection == nil {
			// 	continue
			// }
			var js struct {
				Action string                  `json:"action"`
				Data   webrtc.ICECandidateInit `json:"data"`
				Uuid   string                  `json:"uuid"`
			}
			err = json.Unmarshal(content, &js)
			// check uuid=serveruuid
			candidate := js.Data
			if err == nil {
				if uuid_map[js.Uuid].t == nil {
					peerConnection.AddICECandidate(candidate)
				} else {
					log.Println("转发answer")
					if uuid_map[js.Uuid].s != nil && uuid_map[js.Uuid].t == userdata {
						log.Println("转发给学生端")
						wsSend(uuid_map[js.Uuid].s.wsconn, string(content))
						continue
					}
					if uuid_map[js.Uuid].t != nil && uuid_map[js.Uuid].s == userdata {
						log.Println("转发给教师端")
						wsSend(uuid_map[js.Uuid].t.wsconn, string(content))
						continue
					}
				}
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

func wsSend(ws *websocket.Conn, data interface{}) {
	log.Println("Websocket send: ", data)
	ws.WriteJSON(data)
}

func boardcastEvent(level string, exc *websocket.Conn, data map[string]interface{}) {
	upload := map[string]interface{}{
		"action": "event",
		"data":   data,
	}
	for k := range conn_set {
		if k.uinfo.Level == level && k.wsconn != exc {
			wsSend(k.wsconn, upload)
		}
	}
}
func sendEvent(to *websocket.Conn, data map[string]interface{}) {
	upload := map[string]interface{}{
		"action": "event",
		"data":   data,
	}
	wsSend(to, upload)
}
func sendUuid(to *websocket.Conn, uuid string) {
	upload := map[string]interface{}{
		"action": "uuid",
		"data":   uuid,
	}
	wsSend(to, upload)
}
