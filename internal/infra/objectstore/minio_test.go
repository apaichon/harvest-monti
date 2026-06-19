package objectstore

import (
	"context"
	"reflect"
	"testing"
)

func TestBootstrapEnsuresAllThreeBuckets(t *testing.T) {
	fc := &FakeClient{}
	if err := Bootstrap(context.Background(), fc); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	want := []string{"monti-menu-images", "monti-voice-sessions", "monti-order-receipts"}
	got := fc.Seen()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("bucket bootstrap order mismatch\n got: %v\nwant: %v", got, want)
	}
}

func TestMontiBucketsCarryDesignPolicies(t *testing.T) {
	bs := MontiBuckets()
	if len(bs) != 3 {
		t.Fatalf("want 3 buckets, got %d", len(bs))
	}
	if !bs[0].Public || bs[0].CacheControl == "" {
		t.Errorf("menu-images bucket must be public + cache-controlled, got %+v", bs[0])
	}
	if bs[1].Public || bs[1].LifecycleDays != 30 {
		t.Errorf("voice-sessions bucket must be private with 30-day lifecycle, got %+v", bs[1])
	}
	if bs[2].Public || bs[2].RetentionYears != 7 {
		t.Errorf("order-receipts bucket must be private with 7-year retention, got %+v", bs[2])
	}
}

func TestBootstrapNilClientErrors(t *testing.T) {
	if err := Bootstrap(context.Background(), nil); err == nil {
		t.Fatal("expected error from nil client")
	}
}
