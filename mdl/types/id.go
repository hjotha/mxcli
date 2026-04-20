// SPDX-License-Identifier: Apache-2.0

package types

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateID generates a new unique UUID v4 for model elements.
func GenerateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3],
		b[4], b[5],
		b[6], b[7],
		b[8], b[9],
		b[10], b[11], b[12], b[13], b[14], b[15])
}

// GenerateDeterministicID generates a stable UUID from a seed string.
// Used for System module entities that aren't in the MPR but need consistent IDs.
func GenerateDeterministicID(seed string) string {
	h := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		h[0:4], h[4:6], h[6:8], h[8:10], h[10:16])
}

// BlobToUUID converts a 16-byte binary ID blob to a UUID string.
func BlobToUUID(data []byte) string {
	if len(data) != 16 {
		return hex.EncodeToString(data)
	}
	return fmt.Sprintf("%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		data[3], data[2], data[1], data[0],
		data[5], data[4],
		data[7], data[6],
		data[8], data[9],
		data[10], data[11], data[12], data[13], data[14], data[15])
}

// UUIDToBlob converts a UUID string to a 16-byte blob in Microsoft GUID format.
func UUIDToBlob(uuid string) []byte {
	if uuid == "" {
		return nil
	}
	var clean strings.Builder
	clean.Grow(32)
	for i := 0; i < len(uuid); i++ {
		if uuid[i] != '-' {
			clean.WriteByte(uuid[i])
		}
	}
	decoded, err := hex.DecodeString(clean.String())
	if err != nil || len(decoded) != 16 {
		return nil
	}
	blob := make([]byte, 16)
	blob[0] = decoded[3]
	blob[1] = decoded[2]
	blob[2] = decoded[1]
	blob[3] = decoded[0]
	blob[4] = decoded[5]
	blob[5] = decoded[4]
	blob[6] = decoded[7]
	blob[7] = decoded[6]
	copy(blob[8:], decoded[8:])
	return blob
}

// ValidateID checks if an ID is a valid UUID format.
func ValidateID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i, c := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

// Hash computes a hash for content (used for content deduplication).
func Hash(content []byte) string {
	var sum uint64
	for i, b := range content {
		sum += uint64(b) * uint64(i+1)
	}
	return fmt.Sprintf("%016x", sum)
}
