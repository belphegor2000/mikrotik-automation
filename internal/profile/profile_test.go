package profile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mikrotik-automation/internal/modules/vlan"
)

func TestEncryptDecryptString(t *testing.T) {
	orig := "P@ssw0rd123!_Secure#2026"
	encrypted, err := EncryptString(orig)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	if encrypted == orig {
		t.Fatalf("ciphertext should not match plaintext")
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if decrypted != orig {
		t.Fatalf("expected decrypted '%s', got '%s'", orig, decrypted)
	}
}

func TestProfileManagerCRUD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "prof_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "test_profiles.json")
	SetCustomConfigPath(testPath)
	defer SetCustomConfigPath("")

	pm, err := NewProfileManager()
	if err != nil {
		t.Fatalf("NewProfileManager failed: %v", err)
	}

	// 1. Initial list should be empty
	if len(pm.ListProfiles()) != 0 {
		t.Fatalf("expected 0 profiles initially")
	}

	// 2. Add profile
	p1 := DeviceProfile{
		Name:     "Chi nhánh Quận 1",
		Host:     "192.168.1.1",
		Port:     80,
		User:     "admin",
		Password: "MyRouterPassword123",
		Proto:    "REST API",
	}
	if err := pm.SaveProfile(p1); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	// 3. Verify saved and decrypted
	list := pm.ListProfiles()
	if len(list) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(list))
	}
	if list[0].Name != "Chi nhánh Quận 1" || list[0].Password != "MyRouterPassword123" {
		t.Fatalf("unexpected profile data: %+v", list[0])
	}

	// 4. Update profile
	p1Updated := list[0]
	p1Updated.Password = "NewPass456"
	if err := pm.SaveProfile(p1Updated); err != nil {
		t.Fatalf("failed to update profile: %v", err)
	}
	found, ok := pm.FindProfile("Chi nhánh Quận 1")
	if !ok || found.Password != "NewPass456" {
		t.Fatalf("profile was not updated correctly")
	}

	// 5. Add second profile
	p2 := DeviceProfile{
		Name:     "Router Kho",
		Host:     "10.0.0.1",
		Port:     22,
		User:     "admin",
		Password: "KhoPass789",
		Proto:    "SSH",
	}
	_ = pm.SaveProfile(p2)
	if len(pm.ListProfiles()) != 2 {
		t.Fatalf("expected 2 profiles")
	}

	// 6. Delete profile
	if err := pm.DeleteProfile("Router Kho"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if len(pm.ListProfiles()) != 1 {
		t.Fatalf("expected 1 profile after delete")
	}
}

func TestTemplateExportAndParse(t *testing.T) {
	tmpl := NetworkTemplate{
		Version:     "1.0",
		Name:        "van-phong-standard",
		Description: "Mẫu chuẩn mạng văn phòng 2 WAN",
		CreatedAt:   time.Now(),
		PPPoE: PPPoETemplate{
			WANCount:        2,
			User1:           "user_wan1",
			Pass1:           "pass_wan1",
			User2:           "user_wan2",
			Pass2:           "pass_wan2",
			EnableRecursive: true,
		},
		Bridge: BridgeTemplate{
			BridgeIP:   "192.168.88.1/24",
			EnableDHCP: true,
			DNSServers: "8.8.8.8,1.1.1.1",
		},
		VLANs: []vlan.VLANEntry{
			{VLANID: 10, Name: "vlan10_ketoan", CIDR: "192.168.10.0/24"},
			{VLANID: 20, Name: "vlan20_nhanvien", CIDR: "192.168.20.0/24"},
		},
		PBRRules: []PBRRuleEntry{
			{Name: "Ketoan to WAN1", Type: "VLAN_SUBNET", Source: "192.168.10.0/24", TargetWAN: "WAN1"},
			{Name: "Netflix to WAN2", Type: "DOMAIN_IP", Destination: "netflix.com", TargetWAN: "WAN2"},
		},
		CakeQoS: CakeQoSTemplate{
			Enabled:      true,
			DownloadMbps: 150,
			UploadMbps:   150,
		},
	}

	data, err := ExportTemplateToJSON(tmpl)
	if err != nil {
		t.Fatalf("export template error: %v", err)
	}

	parsed, err := ParseTemplateFromJSON(data)
	if err != nil {
		t.Fatalf("parse template error: %v", err)
	}

	if parsed.Name != "van-phong-standard" || parsed.PPPoE.WANCount != 2 || !parsed.PPPoE.EnableRecursive {
		t.Fatalf("parsed template mismatch: %+v", parsed)
	}
	if len(parsed.VLANs) != 2 || len(parsed.PBRRules) != 2 {
		t.Fatalf("parsed template slice lengths incorrect: %+v", parsed)
	}
	if !parsed.CakeQoS.Enabled || parsed.CakeQoS.DownloadMbps != 150 {
		t.Fatalf("parsed CakeQoS mismatch: %+v", parsed.CakeQoS)
	}
}
