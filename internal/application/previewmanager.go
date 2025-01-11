package application

import (
	"log"
	"pion-webrtc/internal/dto"
	"pion-webrtc/internal/signalling"
	"pion-webrtc/internal/webrtcpeer"
	"sync"
)

type PreviewManager struct {
	signallerClient *signalling.HTTPSignallerClient
	signaller       map[string]*signalling.Signaller
	webrtcPeer      map[string]*webrtcpeer.PeerServer
	wsClient        map[string]*signalling.WebrtcClient
	clientMtx       sync.Mutex
	dataChan        chan string
	videoChan       chan dto.VideoFrame
}

func NewPreviewManager(client *signalling.HTTPSignallerClient,
	dataChan chan string,
	video chan dto.VideoFrame) *PreviewManager {
	return &PreviewManager{
		signallerClient: client,
		webrtcPeer:      make(map[string]*webrtcpeer.PeerServer),
		wsClient:        make(map[string]*signalling.WebrtcClient),
		signaller:       make(map[string]*signalling.Signaller),
		clientMtx:       sync.Mutex{},
		dataChan:        dataChan,
		videoChan:       video,
	}
}

func (prma *PreviewManager) Init() {
	prma.signallerClient.OnNewSignallerClient(prma.onSignallerClient)
}

func (prma *PreviewManager) onSignallerClient(client *signalling.WebrtcClient) {
	log.Println("new signaller client connected...")
	prma.clientMtx.Lock()
	defer func() {
		prma.clientMtx.Unlock()
	}()

	prma.wsClient[client.ID()] = client
	prma.wsClient[client.ID()].SetCloseHandler(prma.onClose)
	prma.signaller[client.ID()] = signalling.NewSignaller()

	prma.webrtcPeer[client.ID()] = webrtcpeer.NewPeer(prma.signaller[client.ID()])
	prma.webrtcPeer[client.ID()].ConfigSignaller()
	prma.wsClient[client.ID()].SetOnSessionDescription(func(msg dto.Message) {
		prma.signaller[client.ID()].OnRemoteOffer(msg)
	})
	prma.wsClient[client.ID()].SetOnRemoteIceCandidate(func(msg dto.Message) {
		prma.signaller[client.ID()].OnRemoteICECandidate(msg)
	})
	prma.signaller[client.ID()].SetOnLocalIceCandidate(func(msg dto.Message) {
		prma.wsClient[client.ID()].SendIceCandidate(msg)
	})

	prma.wsClient[client.ID()].SendSessionDescription(
		prma.signaller[client.ID()].SendOffer())
}

func (prma *PreviewManager) onClose(client *signalling.Client) {
	prma.clientMtx.Lock()
	defer func() {
		prma.clientMtx.Unlock()
	}()
	delete(prma.wsClient, client.ID())
	delete(prma.signaller, client.ID())
	delete(prma.webrtcPeer, client.ID())
}

func (prma *PreviewManager) Run() {
	go func() {
		for {
			select {
			case a := <-prma.dataChan:
				for _, peer := range prma.webrtcPeer {
					peer.SendMessage(a)
				}
			case frame := <-prma.videoChan:
				for _, peer := range prma.webrtcPeer {
					peer.SendFrame(frame)
				}
			}
		}
	}()
}
