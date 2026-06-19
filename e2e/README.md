# e2e — TASK-0014 end-to-end harness

Two modes, same fixtures:

## Stack mode (`make e2e`)

Brings up `compose.e2e.yml` (Postgres, Redis, NATS, MinIO, mock-gemini,
monti-gateway, kiosk), runs the fixture suite against the real services over
HTTP/WS, then tears the stack down. Requires Docker daemon.

## Dry-mode (`make e2e-dry`)

When Docker is unavailable, the suite runs against in-memory fakes shipped
by TASK-0009 (cart/menu/order repos), TASK-0010 (Gemini `FakeTransport`,
Grok `FakePipeline`, in-memory events publisher, NoopInvalidator) and
TASK-0013 (`TEST_NO_INFRA=1` seed). The fixtures (`fixtures/voice/*.cassette.json`)
are replayed by the test loop directly into `FakeTransport.PushTranscript /
PushFunctionCall` — the same JSON contract the docker-mode mock-gemini
serves over HTTP.

The two modes share evidence layout: each test case writes its transcripts,
NATS event dump, dispatcher log, and (for stack mode) screenshots into
`docs/harvest-monti/sdlc/05-tests/evidence/TEST-XXXX/case-N/`.

## Files

- `fixtures/voice/*.cassette.json` — recorded Gemini Live frame sequences.
  Single source of truth for both docker mock-gemini (`cmd/mock-gemini`) and
  dry-mode in-process replay.
- `e2e_test.go` — Go-based fixture suite covering golden path, fallback (4
  triggers), multilingual, identity-stripping.
- `live_smoke_test.go` — flagged smoke test gated by `LIVE_VOICE=1` +
  `GEMINI_API_KEY`. Replaces mock-gemini with real upstream.
- `golden/` — checked-in golden images for kiosk + mobile diffs.

## Cassette schema

```json
{
  "name": "happy-path",
  "frames": [
    {"after_ms": 50, "type": "transcript", "role": "user",
     "text": "...", "final": true},
    {"after_ms": 80, "type": "function_call",
     "name": "menu_search", "call_id": "fc-1",
     "arguments": {"query": "truffle", "allergens": ["peanut"]}},
    {"after_ms": 30, "type": "audio", "pcm": "<base64>", "final": true}
  ]
}
```

Frame types match DES-0009 §2 verbatim — `transcript`, `function_call`,
`audio`, `error`. The `after_ms` field is the inter-frame delay used by both
replay paths so latency assertions are deterministic.
