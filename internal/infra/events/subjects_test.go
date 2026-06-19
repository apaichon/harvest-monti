package events

import (
	"strings"
	"testing"
)

// All harvest-monti subjects must share the "monti." prefix per
// DES-0007 §6 (the JetStream subject filter is "monti.>" — anything else
// would not bind).
func TestAllSubjectsCarryMontiPrefix(t *testing.T) {
	for _, s := range AllSubjects() {
		if !strings.HasPrefix(s, "monti.") {
			t.Errorf("subject %q missing monti. prefix", s)
		}
	}
}

// The design lists 10 customer subjects (menu x2, cart x4, order x3, qr x1)
// plus 3 voice subjects owned by TASK-0010. AllSubjects must include both
// sets so the contract test in harvest-testing can verify completeness.
func TestSubjectCountMatchesDesign(t *testing.T) {
	got := len(AllSubjects())
	want := 13
	if got != want {
		t.Fatalf("expected %d subjects, got %d", want, got)
	}
}

func TestInMemoryPublisherRecordsCalls(t *testing.T) {
	p := &InMemoryPublisher{}
	_ = p.Publish(nil, SubjectOrderPlaced, Envelope{TenantID: "t1"})
	_ = p.Publish(nil, SubjectOrderStatusChanged, Envelope{TenantID: "t1"})
	subs := p.SeenSubjects()
	if len(subs) != 2 || subs[0] != SubjectOrderPlaced || subs[1] != SubjectOrderStatusChanged {
		t.Fatalf("unexpected publisher recording: %v", subs)
	}
}
