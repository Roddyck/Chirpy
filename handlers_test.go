package main

import "testing"

func TestReplaceProfanity(t *testing.T) {
	tests := []struct {
		input  string
		expected string
	}{
		{"This is a kerfuffle opinion I need to share with the world", "This is a **** opinion I need to share with the world"},
		{"This is a sharbert opinion I need to share with the world", "This is a **** opinion I need to share with the world"},
		{"This is a fornax opinion I need to share with the world", "This is a **** opinion I need to share with the world"},
		{"This is a bad opinion I need to share with the world", "This is a bad opinion I need to share with the world"},
		{"", ""},
	}

	for _, test := range tests {
		if got := replaceProfanity(test.input); got != test.expected {
			t.Errorf("replaceProfanity(%q) = %q, expected %q", test.input, got, test.expected)
		}
	}
}
