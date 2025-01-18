package webrtcpeer

import (
	"encoding/json"
	"fmt"
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
	track             map[int]*webrtc.TrackLocalStaticSample
	pipelineSources   int
	transmitAudio     bool
}

func NewPeer(adapter signalling.SignallingAdapter, pipelineSources int, transmitAudio bool) *PeerServer {
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
		track:             make(map[int]*webrtc.TrackLocalStaticSample),
		pipelineSources:   pipelineSources,
		transmitAudio:     transmitAudio,
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

		videoPipelines := peer.pipelineSources
		if peer.transmitAudio {
			videoPipelines--
		}

		peer.addVideoTracks(videoPipelines)

		if peer.transmitAudio {
			peer.addAudioTrack(videoPipelines)
		}
	}
}

func (peer *PeerServer) addAudioTrack(audioPipelineID int) {
	streamID := fmt.Sprintf("audio-%d", audioPipelineID)

	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{ //nolint:exhaustruct
		MimeType: "audio/opus",
	}, "audio", streamID)
	if err != nil {
		log.Fatalf("Error creating track: %v", err.Error())
	}

	peer.track[audioPipelineID] = track
	if _, err := peer.PeerConnection.AddTrack(peer.track[audioPipelineID]); err != nil {
		log.Fatalf("Error adding track: %v", err.Error())
	}
}

func (peer *PeerServer) addVideoTracks(videoPipelines int) {
	for pipelineID := range videoPipelines {
		streamID := fmt.Sprintf("pipeline-%d", pipelineID)

		track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{ //nolint:exhaustruct
			MimeType: "video/vp8",
		}, "video", streamID)
		if err != nil {
			log.Fatalf("Error creating track: %v", err.Error())
		}

		peer.track[pipelineID] = track
		if _, err := peer.PeerConnection.AddTrack(peer.track[pipelineID]); err != nil {
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
	track, ok := peer.track[frame.Source]
	if !ok {
		log.Printf("track %d not found", frame.Source)

		return
	}

	if err := track.WriteSample(media.Sample{ //nolint:exhaustruct
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
