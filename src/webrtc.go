package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264writer"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var conn_set = make(map[*websocket.Conn]*websocket.Conn)

func WebsocketServer(c *gin.Context) {
	log.Println("Websocket Connect")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Panicln("Websocket Error: " + err.Error())
	}
	defer ws.Close()
	defer log.Println("Websocket Close")
	defer delete(conn_set, ws)

	peerConnection := newConnection(ws)
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			log.Panicln("cannot close peerConnection: %v\n" + cErr.Error())
		}
	}()

	conn_set[ws] = ws

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

func newConnection(ws *websocket.Conn) *webrtc.PeerConnection {

	// Create a MediaEngine object to configure the supported codec
	m := &webrtc.MediaEngine{}

	// Setup the codecs you want to use.
	// We'll use a H264 and Opus but you can also define your own
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		log.Panicln(err)
	}
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		log.Panicln(err)
	}
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		log.Panicln(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))

	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Panicln("NewPeerConnection error: " + err.Error())
	}

	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		upload := map[string]interface{}{
			"action": "candidate",
			"data":   i,
		}
		ws.WriteJSON(upload)
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

	// Allow us to receive 1 audio track, and 1 video track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		log.Panicln(err)
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		log.Panicln(err)
	} else if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		log.Panicln(err)
	}

	peerConnection.OnTrack(func(
		track *webrtc.TrackRemote,
		receiver *webrtc.RTPReceiver,
	) {
		log.Printf("%v\n", track.StreamID())
		var file media.Writer
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			file, err = oggwriter.New(track.StreamID()+".ogg", 48000, 2)
			if err != nil {
				log.Panicln(err)
			}
		} else if track.Kind() == webrtc.RTPCodecTypeVideo {
			file, err = h264writer.New(track.StreamID() + ".mp4")
			if err != nil {
				log.Panicln(err)
			}
		}
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if errSend != nil {
					log.Println(errSend)
					return
				}
			}
		}()
		saveToDisk(file, track)
	})

	return peerConnection
}
func saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	defer func() {
		if err := i.Close(); err != nil {
			log.Println(err)
		}
	}()

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			log.Println(err)
			break
		}
		if err := i.WriteRTP(rtpPacket); err != nil {
			log.Println(err)
			break
		}
	}
}
func connectionAnswer(
	peerConnection *webrtc.PeerConnection,
	offer webrtc.SessionDescription,
) webrtc.SessionDescription {

	// Set the remote SessionDescription
	err := peerConnection.SetRemoteDescription(offer)
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
	return answer
}
