package webrtcpeer

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"pion-webrtc/internal/signalling"

	"github.com/pion/webrtc/v3"
)

type PeerClient struct {
	PeerConnection    *webrtc.PeerConnection
	signaller         signalling.SignallingAdapter
	pendingCandidates []*webrtc.ICECandidate
	candidateMux      sync.Mutex
}

func NewPeerClient(adapter signalling.SignallingAdapter) *PeerClient {
	config := webrtc.Configuration{} //nolint:exhaustruct

	peerCon, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil
	}

	return &PeerClient{
		PeerConnection:    peerCon,
		signaller:         adapter,
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
		candidateMux:      sync.Mutex{},
	}
}

func (peer *PeerClient) ConfigSignaller() {
	if peer.signaller != nil {
		peer.signaller.SetOnReceiveOffer(peer.OnReceiveOffer)
		peer.signaller.SetOnReceiveIceCandidate(peer.OnReceiveIceCandidate)
		peer.signaller.SetLocalOfferCallaback(peer.getAnswerLocalOffer)
		peer.PeerConnection.OnICECandidate(peer.OnLocalICECandidate)
		peer.PeerConnection.OnNegotiationNeeded(peer.OnNegotiationNeeded)
		peer.PeerConnection.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
			log.Printf("got track... %s", track.StreamID())

			for {
				_, _, err := track.ReadRTP()
				if err != nil {
					log.Printf("error reading track packet: %s", err.Error())
				}
			}
		})
		peer.PeerConnection.OnDataChannel(func(ch *webrtc.DataChannel) {
			log.Println("got data channel")
			ch.OnMessage(func(msg webrtc.DataChannelMessage) {
				log.Printf("got message: %s", string(msg.Data))
			})
		})
		peer.PeerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
			fmt.Printf("Peer Connection State has changed: %s\n", state.String())

			if state == webrtc.PeerConnectionStateFailed {
				log.Println("Peer Connection has gone to failed exiting")
				os.Exit(0)
			}

			if state == webrtc.PeerConnectionStateClosed {
				log.Println("Peer Connection has gone to closed exiting")
				os.Exit(0)
			}
		})
	}
}

func (peer *PeerClient) OnNegotiationNeeded() {
	log.Println("OnNegotiationNeeded")
}

func (peer *PeerClient) OnLocalICECandidate(candidate *webrtc.ICECandidate) {
	log.Println("OnICECandidate")

	peer.candidateMux.Lock()
	defer peer.candidateMux.Unlock()

	if candidate == nil {
		log.Println("OnICECandidate: candidate is nil")

		return
	}

	peer.pendingCandidates = append(peer.pendingCandidates, candidate)
}

func (peer *PeerClient) getAnswerLocalOffer() webrtc.SessionDescription {
	log.Println("generating answer...")

	answer, err := peer.PeerConnection.CreateAnswer(nil)
	if err != nil {
		fmt.Printf("Error Answer: %v", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peer.PeerConnection)

	if err = peer.PeerConnection.SetLocalDescription(answer); err != nil {
		fmt.Printf("Error Local Description: %v", err)
	}

	log.Println("sending local offer...")
	peer.signaller.SendLocalOffer(answer)

	<-gatherComplete
	log.Println("sending ice candidates...")
	peer.candidateMux.Lock()
	if peer.PeerConnection.RemoteDescription() != nil {
		for _, candidate := range peer.pendingCandidates {
			payload, err := json.Marshal(candidate.ToJSON())
			if err != nil {
				log.Println("Icecandidate parsing error")
			}

			peer.signaller.SendIceCandidate(payload)
		}

		peer.pendingCandidates = make([]*webrtc.ICECandidate, 0)
	}
	peer.candidateMux.Unlock()

	return answer
}
func (peer *PeerClient) OnReceiveOffer(description webrtc.SessionDescription) {
	log.Println("Receiving offer")

	err := peer.PeerConnection.SetRemoteDescription(description)
	if err != nil {
		log.Printf("Error setting remote description: %v", err)

		return
	}

	peer.getAnswerLocalOffer()
}

func (peer *PeerClient) OnReceiveIceCandidate(candidate webrtc.ICECandidateInit) {
	log.Println("OnReceiveIceCandidate -> AddICECandidate")

	err := peer.PeerConnection.AddICECandidate(candidate)
	if err != nil {
		log.Println("Error adding ICE candidate")
	}
}

func (peer *PeerClient) Tracker() {
	items := peer.PeerConnection.GetTransceivers()
	if len(items) < 1 {
		log.Println("no track")

		return
	}

	log.Printf("Track: %d", len(items))
}
