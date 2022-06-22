package main

import (
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

func newConnection(ws *websocket.Conn, conn_data *ConnData) *webrtc.PeerConnection {
	saver_screen := newWebmSaver(conn_data.uinfo.No + "_" + conn_data.uinfo.Name + "_screen_" + GetTime() + ".webm")
	saver_camera := newWebmSaver(conn_data.uinfo.No + "_" + conn_data.uinfo.Name + "_camera_" + GetTime() + ".webm")
	// Create a MediaEngine object to configure the supported codec
	m := &webrtc.MediaEngine{}

	// Setup the codecs you want to use.
	// We'll use a H264 and Opus but you can also define your own
	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
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
	intercept := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, intercept); err != nil {
		log.Panicln(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(intercept))

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
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if errSend != nil {
					return
				}
			}
		}()

		switch track.Kind() {
		case webrtc.RTPCodecTypeAudio:
			saveToDisk(saver_camera, conn_data, track)
			saveToDisk(saver_screen, conn_data, track)
		case webrtc.RTPCodecTypeVideo:
			if track.StreamID() == conn_data.stream_id.camera {
				log.Println("ONTRACK-CAMERA")
				saveToDisk(saver_camera, conn_data, track)
			}
			if track.StreamID() == conn_data.stream_id.screen {
				log.Println("ONTRACK-SCREEN")
				saveToDisk(saver_screen, conn_data, track)
			}
		}
	})
	return peerConnection
}
func saveToDisk(saver *webmSaver, conn_data *ConnData, track *webrtc.TrackRemote) {
	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			if err != io.EOF {
				log.Println(err)
				break
			}
		}
		// 不加这个会panic
		if rtpPacket == nil {
			break
		}
		switch track.Kind() {
		case webrtc.RTPCodecTypeAudio:
			saver.PushOpus(rtpPacket)
		case webrtc.RTPCodecTypeVideo:
			saver.PushVP8(rtpPacket)
		}
	}
	conn_data.close.Lock()
	defer conn_data.close.Unlock()
	saver.Close()
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
