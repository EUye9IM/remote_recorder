package main

import (
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
)

func newConnection(ws *websocket.Conn, conn_data *ConnData) *webrtc.PeerConnection {

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
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		log.Panicln(err)
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i))

	webrtc_config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := api.NewPeerConnection(webrtc_config)
	if err != nil {
		log.Panicln("NewPeerConnection error: " + err.Error())
	}

	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		log.Println("icecandidate转换前：")
		log.Println(i)
		upload := map[string]interface{}{
			"action": "candidate",
			"data":   i.ToJSON(),
			"uuid":   conn_data.uuid,
		}
		wsSend(ws, upload)
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
		var file media.Writer
		class := track.StreamID()
		if track.StreamID() == conn_data.stu_stream_id.camera {
			class = "camera"
		}
		if track.StreamID() == conn_data.stu_stream_id.screen {
			class = "screen"
		}
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			file, err = oggwriter.New(config.Save_path+conn_data.uinfo.No+"_"+conn_data.uinfo.Name+"_"+class+"_"+GetTime()+".ogg", 48000, 2)
			if err != nil {
				log.Panicln(err)
			}
		} else if track.Kind() == webrtc.RTPCodecTypeVideo {
			file, err = ivfwriter.New(config.Save_path + conn_data.uinfo.No + "_" + conn_data.uinfo.Name + "_" + class + "_" + GetTime() + ".ivf")
			if err != nil {
				log.Panicln(err)
			}
		}
		go func() {
			ticker := time.NewTicker(time.Second * 3)
			for range ticker.C {
				errSend := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())}})
				if errSend != nil {
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
			if err != io.EOF {
				log.Println(err)
			}
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
