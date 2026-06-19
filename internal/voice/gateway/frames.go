// Package gateway implements the WebSocket endpoint at
// /api/v1/voice/sessions. It mux-es the 8 client→gateway frames and the 10
// gateway→client frames specified in DES-0009 §2 — the names below are
// reproduced verbatim and are the load-bearing public contract every web /
// mobile client subscribes to.
package gateway

// Client → gateway frame types (DES-0009 §2 — 8 frames, last one is keepalive).
const (
	ClientFrameStartAudio     = "start_audio"
	ClientFrameStopAudio      = "stop_audio"
	ClientFrameMute           = "mute"
	ClientFrameUnmute         = "unmute"
	ClientFrameCancelSpeech   = "cancel_speech"
	ClientFrameLanguageSwitch = "language_switch"
	ClientFrameQuickAction    = "quick_action"
	ClientFramePing           = "ping"
)

// ClientFrameNames is the canonical list — tests and admin tooling enumerate
// this. Order matches DES-0009 §2.
var ClientFrameNames = []string{
	ClientFrameStartAudio,
	ClientFrameStopAudio,
	ClientFrameMute,
	ClientFrameUnmute,
	ClientFrameCancelSpeech,
	ClientFrameLanguageSwitch,
	ClientFrameQuickAction,
	ClientFramePing,
}

// Gateway → client frame types (DES-0009 §2 — 10 frames).
const (
	GatewayFrameSessionOpen      = "session_open"
	GatewayFrameSessionClosed    = "session_closed"
	GatewayFrameTranscript       = "transcript"
	GatewayFrameFunctionCall     = "function_call"
	GatewayFrameFunctionResult   = "function_result"
	GatewayFrameAudioMeter       = "audio_meter"
	GatewayFrameProviderSwitched = "provider_switched"
	GatewayFramePong             = "pong"
	GatewayFrameError            = "error"
	GatewayFrameQuotaWarning     = "quota_warning"
)

// GatewayFrameNames is the canonical list. Order matches DES-0009 §2.
var GatewayFrameNames = []string{
	GatewayFrameSessionOpen,
	GatewayFrameSessionClosed,
	GatewayFrameTranscript,
	GatewayFrameFunctionCall,
	GatewayFrameFunctionResult,
	GatewayFrameAudioMeter,
	GatewayFrameProviderSwitched,
	GatewayFramePong,
	GatewayFrameError,
	GatewayFrameQuotaWarning,
}

// ClientFrame is the JSON envelope from a client.
type ClientFrame struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

// GatewayFrame is the JSON envelope to a client.
type GatewayFrame struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload,omitempty"`
}

// QuickActionName enumerates the four quick actions a kiosk button can fire
// (REQ-0011 AC-8, DES-0008 §6).
const (
	QuickActionRecommendations = "recommendations"
	QuickActionPromotions      = "promotions"
	QuickActionAllergy         = "allergy"
	QuickActionHelp            = "help"
)
