package deduplicator

import (
	"crypto/sha256"
	"fmt"
)

var sentHashes = make(map[string]bool)

type Deduplicator struct {

}

func (d *Deduplicator) ShouldSend(message string) bool {
	messageHash := hashMessage(message)
	if _, exists := sentHashes[messageHash]; exists {
		fmt.Println("Duplicate message detected, not sending.")
		return false
	}

	sentHashes[messageHash] = true

	return true
}

func hashMessage(text string) string {
	hash := sha256.New()
	hash.Write([]byte(text))
	return fmt.Sprintf("%x", hash.Sum(nil))
}