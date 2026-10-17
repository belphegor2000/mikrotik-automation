package i18n

import (
	"testing"
)

func TestI18NLookupsAndFallback(t *testing.T) {
	SetLang(LangVI)
	if GetLang() != LangVI {
		t.Fatalf("expected LangVI, got %s", GetLang())
	}

	// Test VI lookup
	if T("conn_btn") != "🔌 Kết nối" {
		t.Errorf("expected '🔌 Kết nối', got '%s'", T("conn_btn"))
	}

	// Test EN switch
	SetLang(LangEN)
	if GetLang() != LangEN {
		t.Fatalf("expected LangEN, got %s", GetLang())
	}
	if T("conn_btn") != "🔌 Connect" {
		t.Errorf("expected '🔌 Connect', got '%s'", T("conn_btn"))
	}

	// Test fallback when key missing in EN
	dictENBackup := dictEN["missing_key_test"]
	delete(dictEN, "missing_key_test")
	dictVI["missing_key_test"] = "Tiếng Việt Mặc Định"
	defer func() {
		delete(dictVI, "missing_key_test")
		if dictENBackup != "" {
			dictEN["missing_key_test"] = dictENBackup
		}
	}()

	if T("missing_key_test") != "Tiếng Việt Mặc Định" {
		t.Errorf("expected fallback to VI, got '%s'", T("missing_key_test"))
	}

	// Test non-existent key returns key itself
	if T("completely_nonexistent_key_123") != "completely_nonexistent_key_123" {
		t.Errorf("expected key itself, got '%s'", T("completely_nonexistent_key_123"))
	}

	// Reset to VI
	SetLang(LangVI)
}

func TestToggleAndSubscribe(t *testing.T) {
	SetLang(LangVI)
	calledCount := 0
	Subscribe(func() {
		calledCount++
	})

	newLang := ToggleLang()
	if newLang != LangEN {
		t.Errorf("expected LangEN after toggle, got %s", newLang)
	}
	if calledCount != 1 {
		t.Errorf("expected subscriber to be called 1 time, got %d", calledCount)
	}

	newLang2 := ToggleLang()
	if newLang2 != LangVI {
		t.Errorf("expected LangVI after second toggle, got %s", newLang2)
	}
	if calledCount != 2 {
		t.Errorf("expected subscriber to be called 2 times, got %d", calledCount)
	}
}
