package application

import (
	"log"
	"pion-webrtc/internal/dto"
	"pion-webrtc/internal/signalling"
	"pion-webrtc/internal/webrtcpeer"
	"time"
)

type PreviewClient struct {
	signaller  *signalling.Signaller
	webrtcPeer *webrtcpeer.PeerClient
	wsClient   *signalling.WebrtcClient
	videoChan  chan dto.VideoFrame
}

func NewPreviewClient(client *signalling.WebrtcClient, video chan dto.VideoFrame) *PreviewClient {
	return &PreviewClient{
		signaller:  nil,
		webrtcPeer: nil,
		wsClient:   client,
		videoChan:  video,
	}
}

func (pcli *PreviewClient) Init() {
	log.Println("Initializing client...")

	pcli.signaller = signalling.NewSignaller()
	pcli.webrtcPeer = webrtcpeer.NewPeerClient(pcli.signaller)
	pcli.webrtcPeer.ConfigSignaller()
	pcli.wsClient.SetOnSessionDescription(func(message dto.Message) {
		pcli.signaller.OnRemoteOffer(message)
	})
	pcli.wsClient.SetOnRemoteIceCandidate(func(message dto.Message) {
		pcli.signaller.OnRemoteICECandidate(message)
	})
	pcli.signaller.SetOnLocalIceCandidate(func(msg dto.Message) {
		log.Println("sending local ice candidate")
		pcli.wsClient.SendIceCandidate(msg)
	})
	pcli.signaller.SetOnLocalOffer(func(msg dto.Message) {
		log.Println("send local offer")
		pcli.wsClient.SendSessionDescription(msg)
	})
}

func (pcli *PreviewClient) Run() {
	for {
		pcli.webrtcPeer.Tracker()
		time.Sleep(time.Second)
	}
}
