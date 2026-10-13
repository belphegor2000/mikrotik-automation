package auth

import (
	"testing"
)

func TestAuthManager(t *testing.T) {
	am, err := LoadOrCreateAuthManager()
	if err != nil {
		t.Fatalf("failed to init auth manager: %v", err)
	}

	// Correct credentials
	if !am.Authenticate("admin", "Loveisdie1@") {
		t.Fatalf("expected admin with Loveisdie1@ to be authenticated")
	}

	// Wrong password
	if am.Authenticate("admin", "wrongpass") {
		t.Fatalf("expected wrong password to fail")
	}

	// Wrong user
	if am.Authenticate("root", "Loveisdie1@") {
		t.Fatalf("expected wrong user to fail")
	}
}

