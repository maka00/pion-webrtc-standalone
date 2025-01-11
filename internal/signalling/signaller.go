package signalling

import (
	"encoding/json"
	"log"
	"pion-webrtc/internal/dto"

	"github.com/pion/webrtc/v3"
)

type SignallingAdapter interface { //nolint:revive
	OnRemoteICECandidate(msg dto.Message)
	OnRemoteOffer(msg dto.Message)
	SendIceCandidate(candidate []byte)
	SendOffer() dto.Message
	SetOnLocalIceCandidate(callback func(dto.Message))
	SetOnReceiveIceCandidate(callback func(webrtc.ICECandidateInit))
	SetOnReceiveOffer(callback func(webrtc.SessionDescription))
	SetLocalOfferCallaback(callback func() webrtc.SessionDescription)
	SetOnLocalOffer(callback func(dto.Message))
	SendLocalOffer(description webrtc.SessionDescription)
}

type Signaller struct {
	onReceiveIceCandidate func(webrtc.ICECandidateInit)
	onReceiveOffer        func(webrtc.SessionDescription)
	getLocalOffer         func() webrtc.SessionDescription
	onLocalIceCandidate   func(message dto.Message)
	onLocalOffer          func(message dto.Message)
}

func NewSignaller() *Signaller {
	return &Signaller{ //nolint:exhaustruct
		onReceiveIceCandidate: func(webrtc.ICECandidateInit) {},
		onReceiveOffer:        func(webrtc.SessionDescription) {},
	}
}

func (sig *Signaller) SetOnLocalIceCandidate(callback func(msg dto.Message)) {
	sig.onLocalIceCandidate = callback
}

func (sig *Signaller) SetOnReceiveIceCandidate(callback func(webrtc.ICECandidateInit)) {
	sig.onReceiveIceCandidate = callback
}

func (sig *Signaller) SetOnReceiveOffer(callback func(description webrtc.SessionDescription)) {
	sig.onReceiveOffer = callback
}

func (sig *Signaller) SetLocalOfferCallaback(callback func() webrtc.SessionDescription) {
	sig.getLocalOffer = callback
}

func (sig *Signaller) SetOnLocalOffer(callback func(msg dto.Message)) {
	sig.onLocalOffer = callback
}

func (sig *Signaller) SendLocalOffer(description webrtc.SessionDescription) {
	payload := encodeDescription(description)
	sig.onLocalOffer(payload)
}

func (sig *Signaller) OnRemoteICECandidate(msg dto.Message) {
	sig.onReceiveIceCandidate(decodeIceCandidate(msg))
}

func (sig *Signaller) OnRemoteOffer(msg dto.Message) {
	sig.onReceiveOffer(decodeDescription(msg))
}

func (sig *Signaller) SendIceCandidate(candidate []byte) {
	_, err := json.Marshal(candidate)
	if err != nil {
		panic(err)
	}

	msg := dto.Message{
		Type: dto.IceCandidateType,
		Data: string(candidate),
	}

	sig.onLocalIceCandidate(msg)
}

func (sig *Signaller) SendOffer() dto.Message {
	return encodeDescription(sig.getLocalOffer())
}

/// ---- helper functions ----

func decodeIceCandidate(msg dto.Message) webrtc.ICECandidateInit {
	stream := []byte(msg.Data)

	var msgContainer dto.Message
	if err := json.Unmarshal(stream, &msgContainer); err != nil {
		log.Println("parsing error iceCandidate envelope message")
	}

	iceCand := []byte(msgContainer.Data)

	var candidateObject webrtc.ICECandidateInit
	if err := json.Unmarshal(iceCand, &candidateObject); err != nil {
		log.Println("parsing error iceCandidate message")
	}

	return candidateObject
}

func decodeDescription(msg dto.Message) webrtc.SessionDescription {
	var msgContainer dto.Message
	if err := json.Unmarshal([]byte(msg.Data), &msgContainer); err != nil {
		panic(err)
	}

	var desc webrtc.SessionDescription
	if err := json.Unmarshal([]byte(msgContainer.Data), &desc); err != nil {
		panic(err)
	}

	return desc
}

func encodeDescription(description webrtc.SessionDescription) dto.Message {
	offerMessage, err := json.Marshal(description)
	if err != nil {
		panic(err)
	}

	var message dto.Message
	message.Data = string(offerMessage)
	message.Type = dto.SessionDescriptionType

	return message
}
