package protocol

import "encoding/json"

// Message is the envelope for all WS communication
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// inbound + outbound
type UpdatePayload struct {
	ClientID string          `json:"clientId"`
	Changes  json.RawMessage `json:"changes"`
	Version  int             `json:"version"`
}

// outbound only — sent to a newly joined client
type SyncPayload struct {
	Doc      string   `json:"doc"`
	Version  int      `json:"version"`
	Language string   `json:"language"`
	Users    []string `json:"users"`
}

// outbound only — broadcast when someone joins or leaves
type PresencePayload struct {
	Users []string `json:"users"`
}

// inbound + outbound
type LangChangePayload struct {
	Language string `json:"language"`
}