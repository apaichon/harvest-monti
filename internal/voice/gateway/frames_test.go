package gateway_test

import (
	"testing"

	"github.com/apaichon/harvest-monti/internal/voice/gateway"
)

// Lock the frame name lists to DES-0009 §2 verbatim. If anyone renames a
// frame, this test fails and the SDLC contract is violated.
func TestFrameNames_LockedToDES0009(t *testing.T) {
	wantClient := []string{
		"start_audio", "stop_audio", "mute", "unmute",
		"cancel_speech", "language_switch", "quick_action", "ping",
	}
	if len(gateway.ClientFrameNames) != 8 {
		t.Fatalf("expected exactly 8 client frames, got %d", len(gateway.ClientFrameNames))
	}
	for i, n := range wantClient {
		if gateway.ClientFrameNames[i] != n {
			t.Errorf("client frame[%d]: want %q got %q", i, n, gateway.ClientFrameNames[i])
		}
	}

	wantGateway := []string{
		"session_open", "session_closed", "transcript",
		"function_call", "function_result", "audio_meter",
		"provider_switched", "pong", "error", "quota_warning",
	}
	if len(gateway.GatewayFrameNames) != 10 {
		t.Fatalf("expected exactly 10 gateway frames, got %d", len(gateway.GatewayFrameNames))
	}
	for i, n := range wantGateway {
		if gateway.GatewayFrameNames[i] != n {
			t.Errorf("gateway frame[%d]: want %q got %q", i, n, gateway.GatewayFrameNames[i])
		}
	}
}
