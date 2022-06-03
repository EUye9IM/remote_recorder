package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"os"
// 	"sync"
// 	"time"

// 	"github.com/pion/webrtc/v3"
// )

// func RunRTC() { // nolint:gocognit
// 	offerAddr := "localhost:50000"
// 	answerAddr := ":60000"

// 	var candidatesMux sync.Mutex
// 	pendingCandidates := make([]*webrtc.ICECandidate, 0)

// 	// Prepare the configuration
// 	config := webrtc.Configuration{
// 		ICEServers: []webrtc.ICEServer{
// 			{
// 				URLs: Cfg.Webrtc.Ice_list
// 			},
// 		},
// 	}

// 	// Create a new RTCPeerConnection
// 	peerConnection, err := webrtc.NewPeerConnection(config)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer func() {
// 		if err := peerConnection.Close(); err != nil {
// 			fmt.Printf("cannot close peerConnection: %v\n", err)
// 		}
// 	}()

// 	// When an ICE candidate is available send to the other Pion instance
// 	// the other Pion instance will add this candidate by calling AddICECandidate
// 	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
// 		if c == nil {
// 			return
// 		}

// 		candidatesMux.Lock()
// 		defer candidatesMux.Unlock()

// 		desc := peerConnection.RemoteDescription()
// 		if desc == nil {
// 			pendingCandidates = append(pendingCandidates, c)
// 		} else if onICECandidateErr := signalCandidate(*offerAddr, c); onICECandidateErr != nil {
// 			panic(onICECandidateErr)
// 		}
// 	})
// 	// Set the handler for Peer connection state
// 	// This will notify you when the peer has connected/disconnected
// 	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
// 		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

// 		if s == webrtc.PeerConnectionStateFailed {
// 			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
// 			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
// 			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
// 			fmt.Println("Peer Connection has gone to failed exiting")
// 			os.Exit(0)
// 		}
// 	})

// 	// Register data channel creation handling
// 	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
// 		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

// 		// Register channel opening handling
// 		d.OnOpen(func() {
// 			fmt.Printf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

// 			for range time.NewTicker(5 * time.Second).C {
// 				message := signal.RandSeq(15)
// 				fmt.Printf("Sending '%s'\n", message)

// 				// Send the message as text
// 				sendTextErr := d.SendText(message)
// 				if sendTextErr != nil {
// 					panic(sendTextErr)
// 				}
// 			}
// 		})

// 		// Register text message handling
// 		d.OnMessage(func(msg webrtc.DataChannelMessage) {
// 			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
// 		})
// 	})

// 	// Start HTTP server that accepts requests from the offer process to exchange SDP and Candidates
// 	panic(http.ListenAndServe(*answerAddr, nil))
// }
