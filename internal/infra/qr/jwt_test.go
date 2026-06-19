package qr

import (
	"errors"
	"testing"
	"time"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	key := []byte("super-secret-test-key-do-not-use")
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	claims := Claims{
		TenantID:  "tenant-monti-demo",
		OutletID:  "outlet-a",
		TableCode: "T03",
		Exp:       now.Add(5 * time.Minute).Unix(),
		Jti:       "abc-123",
	}
	tok, err := Sign(claims, key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	out, err := Verify(tok, key, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if out.TenantID != claims.TenantID || out.TableCode != claims.TableCode {
		t.Errorf("claims round-trip mismatch: %+v vs %+v", out, claims)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	key := []byte("k")
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	claims := Claims{Exp: now.Add(-time.Second).Unix()}
	tok, _ := Sign(claims, key)
	_, err := Verify(tok, key, now)
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("want ErrExpired, got %v", err)
	}
}

func TestVerifyWrongSignature(t *testing.T) {
	tok, _ := Sign(Claims{Exp: time.Now().Add(time.Minute).Unix()}, []byte("right"))
	_, err := Verify(tok, []byte("wrong"), time.Now())
	if !errors.Is(err, ErrSignature) {
		t.Fatalf("want ErrSignature, got %v", err)
	}
}

func TestVerifyMalformed(t *testing.T) {
	_, err := Verify("not.a.jwt.too.many.parts", []byte("k"), time.Now())
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("want ErrMalformed, got %v", err)
	}
}
