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

		for {
			rtpPacket, _, err := track.ReadRTP()
			switch track.Kind() {
			case webrtc.RTPCodecTypeAudio:
				if err != nil {
					if err != io.EOF {
						log.Println(err)
					}
					if saver_screen.audioWriter != nil {
						if err := saver_screen.audioWriter.Close(); err != nil {
							log.Print(err)
						}
					}
					if saver_camera.audioWriter != nil {
						if err := saver_camera.audioWriter.Close(); err != nil {
							log.Print(err)
						}
					}
					break
				}
				saver_screen.PushOpus(rtpPacket)
				saver_camera.PushOpus(rtpPacket)
			case webrtc.RTPCodecTypeVideo:
				if track.StreamID() == conn_data.stream_id.camera {
					if err != nil {
						if err != io.EOF {
							log.Println(err)
						}
						if saver_camera.videoWriter != nil {
							if err := saver_camera.videoWriter.Close(); err != nil {
								log.Print(err)
							}
						}
						break
					}
					saver_camera.PushVP8(rtpPacket)
				}
				if track.StreamID() == conn_data.stream_id.screen {
					if err != nil {
						if err != io.EOF {
							log.Println(err)
						}
						if saver_screen.videoWriter != nil {
							if err := saver_screen.videoWriter.Close(); err != nil {
								log.Print(err)
							}
						}
						break
					}
					saver_screen.PushVP8(rtpPacket)
				}
			}
		}
	})
	return peerConnection
}

// func saveToDisk(saver *webmSaver, conn_data *ConnData, track *webrtc.TrackRemote) {
// 	for {
// 		// 不加这个会panic
// 		if err != nil {
// 			if err != io.EOF {
// 				log.Println(err)
// 				break
// 			}
// 		}
// 		if rtpPacket == nil {
// 			break
// 		}
// 		switch track.Kind() {
// 		case webrtc.RTPCodecTypeAudio:
// 			saver.PushOpus(rtpPacket)
// 		case webrtc.RTPCodecTypeVideo:
// 			saver.PushVP8(rtpPacket)
// 		}
// 	}
// 	switch track.Kind() {
// 	case webrtc.RTPCodecTypeAudio:
// 		if saver.audioWriter != nil {
// 			if err := saver.audioWriter.Close(); err != nil {
// 				log.Print(err)
// 			}
// 		}
// 	case webrtc.RTPCodecTypeVideo:
// 		if saver.videoWriter != nil {
// 			if err := saver.videoWriter.Close(); err != nil {
// 				log.Print(err)
// 			}
// 		}
// 	}

// }
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
