package dashboard

import (
	"testing"
)

func TestParseBytesValue(t *testing.T) {
	cases := []struct {
		input    interface{}
		expected float64
	}{
		{"1048576", 1.0},
		{"256.0MiB", 256.0},
		{"1.5GiB", 1536.0},
		{"16.0MB", 16.0},
		{"0", 0.0},
		{nil, 0.0},
	}

	for _, tc := range cases {
		got := ParseBytesValue(tc.input)
		if got != tc.expected {
			t.Errorf("ParseBytesValue(%v) = %v; want %v", tc.input, got, tc.expected)
		}
	}
}
