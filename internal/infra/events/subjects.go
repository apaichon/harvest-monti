// Package events declares the NATS JetStream subjects MONTI publishes on
// the HARVEST_MONTI stream (DES-0007 §6). The list here is the contract —
// plugin services import these constants rather than typing the string,
// so renames flow through the type checker.
package events

// StreamName is the JetStream stream that aggregates every monti.> subject.
// devops-agent creates / confirms this stream prior to TASK-0009 landing
// (DES-0007 §6 first paragraph).
const StreamName = "HARVEST_MONTI"

// SubjectFilter is the subject wildcard the stream binds. Anything matching
// "monti.>" is durable for the 14-day retention window.
const SubjectFilter = "monti.>"

// Menu lifecycle subjects (DES-0007 §6).
const (
	SubjectMenuPublished   = "monti.menu.published"
	SubjectMenuItemUpdated = "monti.menu.item.updated"
)

// Cart lifecycle subjects.
const (
	SubjectCartOpened      = "monti.cart.opened"
	SubjectCartItemAdded   = "monti.cart.item.added"
	SubjectCartItemRemoved = "monti.cart.item.removed"
	SubjectCartSubmitted   = "monti.cart.submitted"
)

// Order lifecycle subjects.
const (
	SubjectOrderPlaced        = "monti.order.placed"
	SubjectOrderStatusChanged = "monti.order.status.changed"
	SubjectOrderCompleted     = "monti.order.completed"
)

// Voice lifecycle subjects — declared here for prefix completeness; the
// voice gateway in TASK-0010 owns the publishers.
const (
	SubjectVoiceSessionStarted    = "monti.voice.session.started"
	SubjectVoiceSessionEnded      = "monti.voice.session.ended"
	SubjectVoiceFallbackTriggered = "monti.voice.fallback.triggered"
)

// QR redeem subject (referenced in TASK-0009 §Data contracts).
const SubjectQRRedeemed = "monti.qr.redeemed"

// AllSubjects returns every harvest-monti NATS subject in a stable order.
// Used by the boot-time stream assertion and by the contract test that
// proves we publish on exactly the documented set.
func AllSubjects() []string {
	return []string{
		SubjectMenuPublished,
		SubjectMenuItemUpdated,
		SubjectCartOpened,
		SubjectCartItemAdded,
		SubjectCartItemRemoved,
		SubjectCartSubmitted,
		SubjectOrderPlaced,
		SubjectOrderStatusChanged,
		SubjectOrderCompleted,
		SubjectVoiceSessionStarted,
		SubjectVoiceSessionEnded,
		SubjectVoiceFallbackTriggered,
		SubjectQRRedeemed,
	}
}
