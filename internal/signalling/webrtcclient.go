package signalling

import (
	"pion-webrtc/internal/dto"

	"github.com/gorilla/websocket"
)

type WebrtcClient struct {
	client                     *Client
	onRemoteSessionDescription func(message dto.Message)
	onRemoteIceCadidate        func(message dto.Message)
}

func NewWebrtcClient(conn *websocket.Conn) *WebrtcClient {
	wecl := &WebrtcClient{ //nolint:exhaustruct
		client: NewClient(conn),
	}
	wecl.client.SetOnReceive(dto.SessionDescriptionType, wecl.OnSessionDescription)
	wecl.client.SetOnReceive(dto.IceCandidateType, wecl.OnIceCandidate)
	wecl.client.Start()

	return wecl
}

func (wc *WebrtcClient) ID() string {
	return wc.client.ID()
}

func (wc *WebrtcClient) SetOnSessionDescription(callback func(message dto.Message)) {
	wc.onRemoteSessionDescription = callback
}

func (wc *WebrtcClient) OnSessionDescription(message dto.Message) {
	wc.onRemoteSessionDescription(message)
}

func (wc *WebrtcClient) SetOnRemoteIceCandidate(callback func(message dto.Message)) {
	wc.onRemoteIceCadidate = callback
}

func (wc *WebrtcClient) OnIceCandidate(message dto.Message) {
	wc.onRemoteIceCadidate(message)
}

func (wc *WebrtcClient) SendSessionDescription(message dto.Message) {
	err := wc.client.Send(message)
	if err != nil {
		return
	}
}

func (wc *WebrtcClient) SendIceCandidate(message dto.Message) {
	err := wc.client.Send(message)
	if err != nil {
		return
	}
}

func (wc *WebrtcClient) SetCloseHandler(callback func(clnt *Client)) {
	wc.client.SetCloseHandler(callback)
}
