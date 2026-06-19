// Package qr handles the QR-deep-link JWT used by the mobile/kiosk flow.
//
// The JWT is signed by harvest-core auth's signing key (reused per
// ADR-0003 D3 — no new key material). For TASK-0009 we ship a small HMAC
// signer compatible with the standard JWT HS256 algorithm so the API can
// verify tokens without pulling a third-party JWT library. The signing key
// is supplied at boot from the harvest-core JWKS endpoint.
//
// Payload (DES-0007 §7): {tenant_id, outlet_id, table_code, exp, jti}.
package qr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Claims is the QR JWT payload.
type Claims struct {
	TenantID  string `json:"tenant_id"`
	OutletID  string `json:"outlet_id"`
	TableCode string `json:"table_code"`
	Exp       int64  `json:"exp"`
	Jti       string `json:"jti"`
}

// Errors callers may distinguish on.
var (
	ErrMalformed = errors.New("qr/jwt: malformed token")
	ErrSignature = errors.New("qr/jwt: signature mismatch")
	ErrExpired   = errors.New("qr/jwt: token expired")
)

// header is the fixed HS256 JWT header we emit.
var headerJSON = []byte(`{"alg":"HS256","typ":"JWT"}`)

// Sign returns a compact HS256 JWT for the given claims using key.
func Sign(claims Claims, key []byte) (string, error) {
	body, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("qr/jwt: marshal: %w", err)
	}
	h := base64.RawURLEncoding.EncodeToString(headerJSON)
	p := base64.RawURLEncoding.EncodeToString(body)
	signingInput := h + "." + p

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sig, nil
}

// Verify parses a token, checks the signature and expiry, and returns
// the validated claims. The caller is responsible for the Redis-backed
// jti replay check (10-minute blacklist, TASK-0009 edge cases).
func Verify(token string, key []byte, now time.Time) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrMalformed
	}
	signingInput := parts[0] + "." + parts[1]

	gotSig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrMalformed
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signingInput))
	wantSig := mac.Sum(nil)
	if !hmac.Equal(gotSig, wantSig) {
		return nil, ErrSignature
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrMalformed
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrMalformed
	}
	if claims.Exp > 0 && now.Unix() >= claims.Exp {
		return nil, ErrExpired
	}
	return &claims, nil
}
