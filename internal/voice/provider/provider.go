// Package provider defines the VoiceProvider abstraction every voice adapter
// implements. The interface is reproduced verbatim from ADR-0005 §D3; the
// gateway speaks only to this interface (or to a Switcher wrapping it) so the
// concrete adapter (Gemini Live, Grok pipe) is opaque to the rest of the
// system.
//
// No code outside internal/voice/providers/{gemini,grok}/ is permitted to
// import a vendor SDK (ADR-0005 D3).
package provider

import "context"

// VoiceProvider is the single Go interface every voice runtime implements.
type VoiceProvider interface {
	StartSession(ctx context.Context, cfg SessionConfig) (Session, error)
}

// Session is the bidirectional handle returned by StartSession.
type Session interface {
	// SendAudio forwards a PCM 16 kHz mono frame (~20 ms / 640 bytes) to the
	// provider. Implementations MAY buffer briefly; they MUST NOT block the
	// caller for more than one frame interval.
	SendAudio(frame []byte) error

	// ReceiveAudio returns three fan-out channels: model audio frames,
	// transcript turns, and function-call requests. All three channels are
	// closed by the adapter when the session ends.
	ReceiveAudio() (<-chan AudioFrame, <-chan Transcript, <-chan FunctionCall)

	// RegisterFunctions installs the tool declarations the model may invoke.
	// Must be called before the first audio frame.
	RegisterFunctions(decls []FunctionDecl) error

	// SwitchLanguage signals an explicit user request to change reply
	// language. The model is expected to honour it on the next turn.
	SwitchLanguage(bcp47 string) error

	// Close tears down the underlying provider session.
	Close(reason CloseReason) error

	// SubmitFunctionResult feeds a dispatcher result back into the model so
	// it can speak the answer. Not in ADR-0005 D3's literal listing but
	// required by DES-0009 §4 step 5 ("returns the result into the provider
	// session"). Implementations that do not support tool results route
	// this through a synthetic system message.
	SubmitFunctionResult(result FunctionResult) error
}

// SessionConfig is everything the gateway hands the provider at session start.
type SessionConfig struct {
	TenantID     string
	OutletID     string
	TableCode    string
	SessionID    string
	LanguageHint string // BCP-47, e.g. "th-TH"
	SystemPrompt string
}

// AudioFrame is a chunk of synthesized speech from the model. Format is
// implementation-defined; the gateway re-frames to PCM 16k 20 ms before
// shipping to the client.
type AudioFrame struct {
	PCM        []byte // s16le 16 kHz mono, length is a multiple of 640 bytes
	Sequence   uint32
	IsFinal    bool
	Interrupt  bool // model self-cancelled (barge-in)
	ProducedAt int64
}

// Transcript is a streamed transcription turn.
type Transcript struct {
	Role        string // "user" | "assistant"
	Text        string
	Final       bool
	Interrupted bool
	Language    string // BCP-47, optional
	At          int64
}

// FunctionCall is a tool invocation request emitted by the model.
type FunctionCall struct {
	Name      string
	Arguments map[string]any
	CallID    string
}

// FunctionResult is the dispatcher's reply for a previous FunctionCall.
type FunctionResult struct {
	CallID  string
	Success bool
	Payload map[string]any
	Error   string
}

// FunctionDecl is a tool declaration. The Gemini adapter translates this to
// google.ai.generativelanguage's FunctionDeclaration shape; the Grok adapter
// maps it 1:1 to Grok tool-call format.
type FunctionDecl struct {
	Name        string
	Description string
	Parameters  ParameterSchema
}

// ParameterSchema is a deliberately small JSON-schema subset; the six tools
// in ADR-0005 D4 fit inside it.
type ParameterSchema struct {
	Type       string
	Properties map[string]Property
	Required   []string
}

// Property is a single parameter spec.
type Property struct {
	Type        string
	Description string
	Items       *Property // for arrays
	Enum        []string  // optional
}

// CloseReason explains a session shutdown to the adapter.
type CloseReason struct {
	Code    CloseCode
	Detail  string
	Trigger string // optional human-readable cause
}

// CloseCode enumerates the reasons a Session.Close call can fire.
type CloseCode int

const (
	CloseNormal CloseCode = iota
	CloseClientGone
	CloseIdleTimeout
	CloseProviderError
	CloseQuotaExceeded
	CloseFallbackExhausted
	CloseSessionCeiling
)

// String implements fmt.Stringer for clean log lines.
func (c CloseCode) String() string {
	switch c {
	case CloseNormal:
		return "normal"
	case CloseClientGone:
		return "client_gone"
	case CloseIdleTimeout:
		return "idle_timeout"
	case CloseProviderError:
		return "provider_error"
	case CloseQuotaExceeded:
		return "quota_exceeded"
	case CloseFallbackExhausted:
		return "fallback_exhausted"
	case CloseSessionCeiling:
		return "session_ceiling"
	default:
		return "unknown"
	}
}
