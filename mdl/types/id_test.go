// SPDX-License-Identifier: Apache-2.0

package types

import (
	"strings"
	"testing"
)

func TestGenerateID_Format(t *testing.T) {
	id := GenerateID()
	if !ValidateID(id) {
		t.Fatalf("GenerateID() returned invalid UUID: %q", id)
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateID()
		if seen[id] {
			t.Fatalf("GenerateID() produced duplicate: %q", id)
		}
		seen[id] = true
	}
}

func TestGenerateID_V4Bits(t *testing.T) {
	id := GenerateID()
	// Version nibble at position 14 (0-indexed in hex chars) should be '4'
	clean := strings.ReplaceAll(id, "-", "")
	if clean[12] != '4' {
		t.Errorf("expected version nibble '4', got %q in %q", string(clean[12]), id)
	}
	// Variant nibble at position 16 should be 8, 9, a, or b
	v := clean[16]
	if v != '8' && v != '9' && v != 'a' && v != 'b' {
		t.Errorf("expected variant nibble in [89ab], got %q in %q", string(v), id)
	}
}

func TestGenerateDeterministicID_V4Bits(t *testing.T) {
	seeds := []string{"test", "hello", "System.User", "System.Session", ""}
	for _, seed := range seeds {
		id := GenerateDeterministicID(seed)
		clean := strings.ReplaceAll(id, "-", "")
		// Version nibble at hex position 12 should be '4'
		if clean[12] != '4' {
			t.Errorf("seed %q: expected version nibble '4', got %q in %q", seed, string(clean[12]), id)
		}
		// Variant nibble at hex position 16 should be 8, 9, a, or b
		v := clean[16]
		if v != '8' && v != '9' && v != 'a' && v != 'b' {
			t.Errorf("seed %q: expected variant nibble in [89ab], got %q in %q", seed, string(v), id)
		}
	}
}

func TestGenerateDeterministicID_Stable(t *testing.T) {
	id1 := GenerateDeterministicID("test-seed")
	id2 := GenerateDeterministicID("test-seed")
	if id1 != id2 {
		t.Fatalf("expected same ID for same seed, got %q and %q", id1, id2)
	}
}

func TestGenerateDeterministicID_DifferentSeeds(t *testing.T) {
	id1 := GenerateDeterministicID("seed-a")
	id2 := GenerateDeterministicID("seed-b")
	if id1 == id2 {
		t.Fatalf("expected different IDs for different seeds")
	}
}

func TestGenerateDeterministicID_Format(t *testing.T) {
	id := GenerateDeterministicID("test")
	// Should be 36 chars: 8-4-4-4-12
	if len(id) != 36 {
		t.Fatalf("expected 36 chars, got %d: %q", len(id), id)
	}
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("expected 5 dash-separated parts, got %d", len(parts))
	}
	expectedLens := []int{8, 4, 4, 4, 12}
	for i, p := range parts {
		if len(p) != expectedLens[i] {
			t.Errorf("part %d: expected %d chars, got %d", i, expectedLens[i], len(p))
		}
	}
}

func TestBlobToUUID_RoundTrip(t *testing.T) {
	uuid := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	blob := UUIDToBlob(uuid)
	if blob == nil {
		t.Fatal("UUIDToBlob returned nil")
	}
	got := BlobToUUID(blob)
	if got != uuid {
		t.Errorf("roundtrip failed: %q -> blob -> %q", uuid, got)
	}
}

func TestBlobToUUID_Non16Bytes(t *testing.T) {
	// Non-16-byte input should return hex-encoded string
	data := []byte{0x01, 0x02, 0x03}
	got := BlobToUUID(data)
	if got != "010203" {
		t.Errorf("expected hex fallback '010203', got %q", got)
	}
}

func TestBlobToUUID_Empty(t *testing.T) {
	got := BlobToUUID(nil)
	if got != "" {
		t.Errorf("expected empty string for nil, got %q", got)
	}
}

func TestUUIDToBlob_Empty(t *testing.T) {
	if got := UUIDToBlob(""); got != nil {
		t.Errorf("expected nil for empty string, got %v", got)
	}
}

func TestUUIDToBlob_Invalid(t *testing.T) {
	if got := UUIDToBlob("not-a-uuid"); got != nil {
		t.Errorf("expected nil for invalid UUID, got %v", got)
	}
}

func TestUUIDToBlob_GUIDByteSwap(t *testing.T) {
	// The first 4 bytes should be reversed, next 2 reversed, next 2 reversed, rest same
	blob := UUIDToBlob("01020304-0506-0708-090a-0b0c0d0e0f10")
	if blob == nil {
		t.Fatal("UUIDToBlob returned nil")
	}
	// First group: 01020304 -> blob[0..3] = 04,03,02,01
	expected := []byte{0x04, 0x03, 0x02, 0x01, 0x06, 0x05, 0x08, 0x07,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	for i, b := range blob {
		if b != expected[i] {
			t.Errorf("byte %d: expected 0x%02x, got 0x%02x", i, expected[i], b)
		}
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"a1b2c3d4-e5f6-7890-abcd-ef1234567890", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"AABBCCDD-EEFF-1122-3344-556677889900", true},
		{"", false},
		{"too-short", false},
		{"a1b2c3d4-e5f6-7890-abcd-ef123456789", false},   // 35 chars
		{"a1b2c3d4-e5f6-7890-abcd-ef12345678901", false}, // 37 chars
		{"a1b2c3d4xe5f6-7890-abcd-ef1234567890", false},  // wrong separator
		{"g1b2c3d4-e5f6-7890-abcd-ef1234567890", false},  // invalid hex
	}
	for _, tt := range tests {
		got := ValidateID(tt.input)
		if got != tt.valid {
			t.Errorf("ValidateID(%q) = %v, want %v", tt.input, got, tt.valid)
		}
	}
}
