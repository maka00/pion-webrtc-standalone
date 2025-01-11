package dto

const (
	IceCandidateType       = "iceCandidate"
	SessionDescriptionType = "sessionDescription"
)

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}
