// SPDX-License-Identifier: Apache-2.0

package mpr

import (
	"testing"
)

func TestValidateNoPlaceholderIDs_BinaryPattern(t *testing.T) {
	// Simulate BSON contents containing the placeholder binary pattern.
	// The GUID-swapped placeholder prefix is \x00\x00\x00\xaa followed by 9 zero bytes.
	contents := []byte("some bson preamble")
	contents = append(contents, 0x00, 0x00, 0x00, 0xaa)       // GUID-swapped first 4 bytes
	contents = append(contents, 0x00, 0x00, 0x00, 0x00, 0x00) // bytes 4-8
	contents = append(contents, 0x00, 0x00, 0x00, 0x00)       // bytes 9-12
	contents = append(contents, 0x00, 0x00, 0x01)             // counter bytes
	contents = append(contents, []byte("more bson data")...)

	err := validateNoPlaceholderIDs("test-unit-id", contents)
	if err == nil {
		t.Fatal("expected error for placeholder binary pattern, got nil")
	}
	if got := err.Error(); got == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestValidateNoPlaceholderIDs_StringPattern(t *testing.T) {
	// Simulate BSON contents containing a placeholder as an ASCII string
	contents := []byte("some bson preamble aa000000000000000000000000000003 more data")

	err := validateNoPlaceholderIDs("test-unit-id", contents)
	if err == nil {
		t.Fatal("expected error for placeholder string pattern, got nil")
	}
}

func TestValidateNoPlaceholderIDs_Clean(t *testing.T) {
	// Normal BSON-like data with no placeholder patterns
	contents := []byte{
		0x1a, 0x00, 0x00, 0x00, // BSON document length
		0x02,                         // string type
		0x6e, 0x61, 0x6d, 0x65, 0x00, // "name\0"
		0x08, 0x00, 0x00, 0x00, // string length
		0x54, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x00, // "Testing\0"
		0x05,             // binary type
		0x69, 0x64, 0x00, // "id\0"
		0x10, 0x00, 0x00, 0x00, 0x00, // binary length + subtype
		0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0x0a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78, // legitimate UUID
		0x00, // document terminator
	}

	err := validateNoPlaceholderIDs("test-unit-id", contents)
	if err != nil {
		t.Fatalf("expected no error for clean data, got: %v", err)
	}
}
