// Package cache holds the cache-tag taxonomy for harvest-monti.
// Every tag MUST start with the "monti:" prefix per DES-0007 §5 and
// ADR-0003 D5 to avoid colliding with the reserved harvest-core
// namespaces (analytics:, duckdb:, clickhouse:, usage:).
//
// Add new helpers here — never inline a tag string in a plugin handler.
package cache

import "fmt"

// Prefix is the immutable namespace prefix for every harvest-monti cache key.
const Prefix = "monti:"

// MenuTag returns the tenant-wide menu cache tag, invalidated when a
// menu is published or its status changes (DES-0007 §5 row 1).
func MenuTag(tenantID string) string {
	return fmt.Sprintf("%smenu:tenant:%s", Prefix, tenantID)
}

// MenuItemTag returns the per-item cache tag (DES-0007 §5 row 2).
func MenuItemTag(tenantID, itemID string) string {
	return fmt.Sprintf("%smenu:%s:item:%s", Prefix, tenantID, itemID)
}

// CartTag returns the cart cache tag bound to a kiosk/mobile session
// (DES-0007 §5 row 3).
func CartTag(tenantID, sessionID string) string {
	return fmt.Sprintf("%scart:%s:%s", Prefix, tenantID, sessionID)
}

// OrderTag returns the order cache tag invalidated on every
// fulfillment_status transition (DES-0007 §5 row 4).
func OrderTag(tenantID, orderID string) string {
	return fmt.Sprintf("%sorder:%s:%s", Prefix, tenantID, orderID)
}

// VoiceTag is reserved for TASK-0010 voice session keys; kept here so
// the prefix-test catches future drift.
func VoiceTag(tenantID, sessionID string) string {
	return fmt.Sprintf("%svoice:%s:%s", Prefix, tenantID, sessionID)
}
