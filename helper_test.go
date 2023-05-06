package main

import (
	"testing"
)

func TestCalculateWords(t *testing.T) {
	cases := []struct {
		messages []map[string]string
		expected int
	}{
		{
			messages: []map[string]string{
				{"role": "user", "content": "Hello, world How are you!"},
				{"role": "bot", "content": "Hi there!"},
			},
			expected: 7,
		},
		{
			messages: []map[string]string{
				{"role": "user", "content": "Hello"},
			},
			expected: 1,
		},
		{
			messages: []map[string]string{
				{"role": "user", "content": ""},
			},
			expected: 0,
		},
	}

	// Run tests
	for _, c := range cases {
		result := calculateWords(c.messages)
		if result != c.expected {
			t.Errorf("TestCalculateWords(%v) == %d, expected %d", c.messages, result, c.expected)
		}
	}
}

func TestRandomString(t *testing.T) {

}
