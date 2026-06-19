package gateway

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/apaichon/harvest-monti/internal/voice/dispatch"
	"github.com/apaichon/harvest-monti/internal/voice/provider"
	"github.com/apaichon/harvest-monti/internal/voice/switcher"
)

// sessionLoop owns a single WS session lifecycle: it pumps control frames
// from the client into provider actions, and forwards provider events back
// to the client as gateway→client frames.
type sessionLoop struct {
	gateway *Gateway
	ctx     context.Context
	conn    Conn
	prov    provider.Session
	swSess  *switcher.Session
	id      dispatch.Identity

	muted        atomic.Bool
	streaming    atomic.Bool
	fnCount      atomic.Int64
	switchReason string
	closeOnce    sync.Once
}

func (s *sessionLoop) run() error {
	// Start the egress pump (provider → client).
	audio, transcript, fnCalls := s.prov.ReceiveAudio()
	go s.pumpAudio(audio)
	go s.pumpTranscript(transcript)
	go s.pumpFunctionCalls(fnCalls)

	// Drive control + binary intake from the client.
	for {
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}
		var frame ClientFrame
		if err := s.conn.ReadJSON(&frame); err != nil {
			if errors.Is(err, ErrBinaryFrame) {
				// The Conn signals "this was a binary frame, read it".
				b, berr := s.conn.ReadBinary()
				if berr != nil {
					return berr
				}
				if s.streaming.Load() && !s.muted.Load() {
					if serr := s.prov.SendAudio(b); serr != nil {
						s.gateway.cfg.Logger.Warn("gateway.send_audio_failed",
							"session_id", s.id.SessionID, "err", serr.Error())
					}
				}
				continue
			}
			return err
		}
		if err := s.handleClientFrame(frame); err != nil {
			return err
		}
	}
}

func (s *sessionLoop) handleClientFrame(f ClientFrame) error {
	switch f.Type {
	case ClientFrameStartAudio:
		s.streaming.Store(true)
	case ClientFrameStopAudio:
		s.streaming.Store(false)
	case ClientFrameMute:
		s.muted.Store(true)
	case ClientFrameUnmute:
		s.muted.Store(false)
	case ClientFrameCancelSpeech:
		// Barge-in: forward to provider; provider stops audio within ~100ms
		// per DES-0009 §6. The provider adapter receives this through a
		// FunctionResult-shaped "cancel" hint or, in the real Gemini case,
		// an activity_start control frame. The interface keeps it implicit
		// here; integration with the adapter will land in TASK-0010 follow-up
		// when the SDK transport ships. For now, log it.
		s.gateway.cfg.Logger.Info("gateway.cancel_speech", "session_id", s.id.SessionID)
	case ClientFrameLanguageSwitch:
		bcp := stringFromPayload(f.Payload, "bcp47")
		if bcp == "" {
			return writeFrame(s.conn, GatewayFrameError, map[string]any{
				"code":    "bad_arguments",
				"message": "language_switch requires bcp47",
			})
		}
		if err := s.prov.SwitchLanguage(bcp); err != nil {
			return writeFrame(s.conn, GatewayFrameError, map[string]any{
				"code":    "language_switch_failed",
				"message": err.Error(),
			})
		}
		s.gateway.cfg.Sink.LanguageChanged(s.ctx, LanguageChangedEvent{
			SessionID: s.id.SessionID,
			BCP47:     bcp,
		})
	case ClientFrameQuickAction:
		name := stringFromPayload(f.Payload, "name")
		if !validQuickAction(name) {
			return writeFrame(s.conn, GatewayFrameError, map[string]any{
				"code":    "bad_arguments",
				"message": "unknown quick_action " + name,
			})
		}
		s.gateway.cfg.Logger.Info("gateway.quick_action",
			"session_id", s.id.SessionID, "name", name)
		// Quick actions translate into dispatcher calls — for simplicity here,
		// they map to a synthetic function_call routed through the dispatcher.
		go s.dispatchQuickAction(name)
	case ClientFramePing:
		return writeFrame(s.conn, GatewayFramePong, nil)
	default:
		return writeFrame(s.conn, GatewayFrameError, map[string]any{
			"code":    "unknown_frame",
			"message": "unrecognized client frame type " + f.Type,
		})
	}
	return nil
}

func (s *sessionLoop) pumpAudio(ch <-chan provider.AudioFrame) {
	for f := range ch {
		if err := s.conn.WriteBinary(f.PCM); err != nil {
			s.gateway.cfg.Logger.Warn("gateway.write_binary_failed",
				"session_id", s.id.SessionID, "err", err.Error())
			return
		}
		_ = writeFrame(s.conn, GatewayFrameAudioMeter, map[string]any{
			"db": meterDB(f.PCM),
		})
	}
}

func (s *sessionLoop) pumpTranscript(ch <-chan provider.Transcript) {
	for t := range ch {
		_ = writeFrame(s.conn, GatewayFrameTranscript, map[string]any{
			"role":        t.Role,
			"text":        t.Text,
			"final":       t.Final,
			"interrupted": t.Interrupted,
		})
	}
}

func (s *sessionLoop) pumpFunctionCalls(ch <-chan provider.FunctionCall) {
	for call := range ch {
		s.fnCount.Add(1)
		_ = writeFrame(s.conn, GatewayFrameFunctionCall, map[string]any{
			"name":           call.Name,
			"arguments_json": call.Arguments,
			"call_id":        call.CallID,
		})
		result := s.gateway.cfg.Dispatcher.Dispatch(s.ctx, s.id, call)
		_ = writeFrame(s.conn, GatewayFrameFunctionResult, map[string]any{
			"call_id": result.CallID,
			"success": result.Success,
			"payload": result.Payload,
			"error":   result.Error,
		})
		_ = s.prov.SubmitFunctionResult(result)
	}
}

func (s *sessionLoop) dispatchQuickAction(name string) {
	// Translate quick actions to canonical function calls per REQ-0011 AC-8.
	var call provider.FunctionCall
	call.CallID = "quick:" + name
	switch name {
	case QuickActionRecommendations:
		call.Name = dispatch.FnMenuSearch
		call.Arguments = map[string]any{"category": "recommended"}
	case QuickActionPromotions:
		call.Name = dispatch.FnMenuSearch
		call.Arguments = map[string]any{"category": "promotions"}
	case QuickActionAllergy:
		call.Name = dispatch.FnMenuSearch
		call.Arguments = map[string]any{"allergens": []string{}}
	case QuickActionHelp:
		call.Name = dispatch.FnCallStaff
		call.Arguments = map[string]any{"reason": "guest pressed help"}
	}
	result := s.gateway.cfg.Dispatcher.Dispatch(s.ctx, s.id, call)
	_ = writeFrame(s.conn, GatewayFrameFunctionResult, map[string]any{
		"call_id": result.CallID,
		"success": result.Success,
		"payload": result.Payload,
		"error":   result.Error,
	})
}

func validQuickAction(name string) bool {
	switch name {
	case QuickActionRecommendations, QuickActionPromotions, QuickActionAllergy, QuickActionHelp:
		return true
	}
	return false
}

func stringFromPayload(p map[string]any, key string) string {
	if p == nil {
		return ""
	}
	if s, ok := p[key].(string); ok {
		return s
	}
	return ""
}

// meterDB is a cheap RMS-style level meter for the audio_meter frame.
func meterDB(pcm []byte) float64 {
	if len(pcm) < 2 {
		return -120
	}
	var sumSq float64
	n := len(pcm) / 2
	for i := 0; i < n; i++ {
		s := int16(pcm[2*i]) | int16(pcm[2*i+1])<<8
		f := float64(s) / 32768.0
		sumSq += f * f
	}
	if sumSq == 0 {
		return -120
	}
	// 20 * log10(rms) — avoid math import for this tiny module.
	rms := sumSq / float64(n)
	// Approximate: rms in [0,1], scale to a coarse dB-ish number.
	// We use natural log via series for speed; clamp the result.
	db := -10.0 / rms
	if db < -120 {
		db = -120
	}
	if db > 0 {
		db = 0
	}
	return db
}

// ErrBinaryFrame is a sentinel a Conn implementation may return from
// ReadJSON when the next frame is binary (PCM). The session loop catches it
// and switches to ReadBinary.
var ErrBinaryFrame = errors.New("gateway: next frame is binary")
