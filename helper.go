package main

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func calculateWords(messages []map[string]string) int {
	wordsCount := 0
	for _, msg := range messages {
		msgContent := msg["content"]
		wordsCount += len(strings.Fields(msgContent))
	}
	return wordsCount
}

func randomString() (string, error) {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a string
	return base64.URLEncoding.EncodeToString(bytes), nil
}
