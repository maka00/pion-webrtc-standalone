package webrtcpeer

import (
	"encoding/json"
	"log"
	"pion-webrtc/internal/dto"
	"pion-webrtc/internal/signalling"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type PeerServer struct {
	PeerConnection    *webrtc.PeerConnection
	signaller         signalling.SignallingAdapter
	pendingCandidates []*webrtc.ICECandidate
	datachannel       *webrtc.DataChannel
	track             *webrtc.TrackLocalStaticSample
}

func NewPeer(adapter signalling.SignallingAdapter) *PeerServer {
	config := webrtc.Configuration{} //nolint:exhaustruct

	peerCon, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil
	}

	return &PeerServer{
		PeerConnection:    peerCon,
		signaller:         adapter,
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
		datachannel:       nil,
		track:             nil,
	}
}

func (peer *PeerServer) ConfigSignaller() {
	if peer.signaller != nil {
		peer.signaller.SetOnReceiveOffer(peer.OnReceiveOffer)
		peer.signaller.SetOnReceiveIceCandidate(peer.OnReceiveIceCandidate)
		peer.signaller.SetLocalOfferCallaback(peer.getLocalOffer)
		peer.PeerConnection.OnICECandidate(peer.OnLocalICECandidate)
		peer.PeerConnection.OnNegotiationNeeded(peer.OnNegotiationNeeded)

		var err error

		peer.datachannel, err = peer.PeerConnection.CreateDataChannel("data", nil)
		if err != nil {
			log.Println("Error creating data channel")
		}

		peer.datachannel.OnMessage(func(_ webrtc.DataChannelMessage) {
			log.Println("Data channel message received")
		})

		peer.datachannel.OnOpen(func() {
			log.Println("Data channel open")
		})

		peer.track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{ //nolint:exhaustruct
			MimeType: "video/vp8",
		}, "video", "pion")
		if err != nil {
			log.Fatalf("Error creating track: %v", err.Error())
		}

		if _, err := peer.PeerConnection.AddTrack(peer.track); err != nil {
			log.Fatalf("Error adding track: %v", err.Error())
		}
	}
}

func (peer *PeerServer) SendMessage(message string) {
	if peer.datachannel != nil {
		err := peer.datachannel.Send([]byte(message))
		if err != nil {
			log.Println("Error sending message")

			return
		}
	}
}

func (peer *PeerServer) SendFrame(frame dto.VideoFrame) {
	if err := peer.track.WriteSample(media.Sample{ //nolint:exhaustruct
		Data:     frame.Frame,
		Duration: frame.Duration,
	}); err != nil {
		log.Printf("error sending frame: %v", err)
	}
}

func (peer *PeerServer) OnNegotiationNeeded() {
	log.Println("OnNegotiationNeeded")
}

func (peer *PeerServer) OnLocalICECandidate(candidate *webrtc.ICECandidate) {
	log.Println("OnICECandidate")

	if candidate == nil {
		log.Println("OnICECandidate: candidate is nil")

		return
	}

	peer.pendingCandidates = append(peer.pendingCandidates, candidate)
}

func (peer *PeerServer) OnReceiveOffer(description webrtc.SessionDescription) {
	err := peer.PeerConnection.SetRemoteDescription(description)
	if err != nil {
		log.Println("Error setting remote description")

		return
	}

	if peer.PeerConnection.RemoteDescription() != nil {
		for _, candidate := range peer.pendingCandidates {
			payload, err := json.Marshal(candidate.ToJSON())
			if err != nil {
				log.Println("receive offer mashal error")
			}

			peer.signaller.SendIceCandidate(payload)
		}

		peer.pendingCandidates = make([]*webrtc.ICECandidate, 0)
	}
}

func (peer *PeerServer) OnReceiveIceCandidate(candidate webrtc.ICECandidateInit) {
	log.Println("OnReceiveIceCandidate -> AddICECandidate")

	err := peer.PeerConnection.AddICECandidate(candidate)
	if err != nil {
		log.Println("Error adding ICE candidate")
	}
}

func (peer *PeerServer) getLocalOffer() webrtc.SessionDescription {
	offer, err := peer.PeerConnection.CreateOffer(nil)
	if err != nil {
		return webrtc.SessionDescription{} //nolint:exhaustruct
	}

	err = peer.PeerConnection.SetLocalDescription(offer)
	if err != nil {
		log.Println("Error setting local description")

		return webrtc.SessionDescription{} //nolint:exhaustruct
	}

	return offer
}
