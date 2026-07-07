package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	cases := []struct {
		in   string
		want string
	}{
		{"~", home},
		{"~/projects/app", filepath.Join(home, "projects/app")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~user", "~user"}, // ~ not followed by "/" is left untouched
		{"", ""},
	}

	for _, tc := range cases {
		if got := expandHome(tc.in); got != tc.want {
			t.Errorf("expandHome(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
