package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apaichon/harvest-monti/internal/infra/qr"
)

func TestQRRedeemHappyPath(t *testing.T) {
	key := []byte("test-key")
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	tok, err := qr.Sign(qr.Claims{
		TenantID: "tA", OutletID: "oA", TableCode: "T07",
		Exp: now.Add(5 * time.Minute).Unix(), Jti: "abc",
	}, key)
	if err != nil {
		t.Fatal(err)
	}
	r := New()
	pub := r.Group("/api/v1/public")
	jti := NewMemoryJtiStore(func() time.Time { return now })
	pub.Post("/qr/redeem", QRRedeemHandler(PublicDeps{
		QRKey: key, JtiBlacklist: jti, Now: func() time.Time { return now },
	}))

	body, _ := json.Marshal(map[string]string{"token": tok})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/qr/redeem", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var resp map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["tenant_id"] != "tA" || resp["table_code"] != "T07" {
		t.Errorf("unexpected redeem body: %v", resp)
	}
	if resp["session_id"] == "" {
		t.Errorf("session_id not returned")
	}

	// Second redeem of the same jti must 409.
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/public/qr/redeem", bytes.NewReader(body))
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("replay must 409, got %d", rec2.Code)
	}
}

func TestQRRedeemRejectsBadSignature(t *testing.T) {
	r := New()
	pub := r.Group("/api/v1/public")
	pub.Post("/qr/redeem", QRRedeemHandler(PublicDeps{
		QRKey: []byte("right"), JtiBlacklist: NewMemoryJtiStore(nil), Now: time.Now,
	}))
	tok, _ := qr.Sign(qr.Claims{Exp: time.Now().Add(time.Minute).Unix()}, []byte("wrong"))
	body, _ := json.Marshal(map[string]string{"token": tok})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/public/qr/redeem", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}
