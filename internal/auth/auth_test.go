package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuthManager(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "test_auth.json")
	SetCustomConfigPath(customPath)
	defer SetCustomConfigPath("")

	am, err := LoadOrCreateAuthManager()
	if err != nil {
		t.Fatalf("failed to init auth manager: %v", err)
	}

	// Correct initial credentials
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

	// Test ChangePassword - validation failure (<6 chars)
	if err := am.ChangePassword("admin", "Loveisdie1@", "123"); err == nil {
		t.Fatalf("expected error for short new password")
	}

	// Test ChangePassword - wrong old password
	if err := am.ChangePassword("admin", "wrongold", "NewValidPass123"); err == nil {
		t.Fatalf("expected error for incorrect old password")
	}

	// Test ChangePassword - success
	if err := am.ChangePassword("admin", "Loveisdie1@", "NewValidPass123"); err != nil {
		t.Fatalf("failed to change password: %v", err)
	}

	// Old password should now fail
	if am.Authenticate("admin", "Loveisdie1@") {
		t.Fatalf("expected old password to fail after change")
	}

	// New password should succeed
	if !am.Authenticate("admin", "NewValidPass123") {
		t.Fatalf("expected new password to succeed")
	}

	// Reload from file to verify persistence
	amReloaded, err := LoadOrCreateAuthManager()
	if err != nil {
		t.Fatalf("failed to reload auth manager: %v", err)
	}
	if !amReloaded.Authenticate("admin", "NewValidPass123") {
		t.Fatalf("expected reloaded auth manager to authenticate new password")
	}
}

func TestGetAuthConfigPath(t *testing.T) {
	p := GetAuthConfigPath()
	if p == "" {
		t.Fatalf("expected non-empty config path")
	}
	_ = os.Getenv("TEMP")
}
