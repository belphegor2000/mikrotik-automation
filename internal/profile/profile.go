package profile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"mikrotik-automation/internal/modules/vlan"
)

const (
	ProfilesConfigFile = "profiles.json"
	appSecretSalt      = "MK-AUTO-ROUTER-PROFILE-SEC-v1@2026"
)

var (
	customConfigPath string
	muPath           sync.RWMutex
)

// SetCustomConfigPath overrides the profile configuration file path (useful for testing).
func SetCustomConfigPath(path string) {
	muPath.Lock()
	defer muPath.Unlock()
	customConfigPath = path
}

// GetProfilesConfigPath returns the persistent path for profiles.json.
func GetProfilesConfigPath() string {
	muPath.RLock()
	if customConfigPath != "" {
		p := customConfigPath
		muPath.RUnlock()
		return p
	}
	muPath.RUnlock()

	// Try alongside the executable
	if exePath, err := os.Executable(); err == nil && exePath != "" {
		exeDir := filepath.Dir(exePath)
		target := filepath.Join(exeDir, ProfilesConfigFile)
		if _, err := os.Stat(target); err == nil {
			return target
		}
		testFile := filepath.Join(exeDir, ".write_test_prof_"+fmt.Sprintf("%d", os.Getpid()))
		if err := os.WriteFile(testFile, []byte("ok"), 0600); err == nil {
			_ = os.Remove(testFile)
			return target
		}
	}

	// Fallback to User Config Dir (%APPDATA%/MikrotikAutomation on Windows)
	if configDir, err := os.UserConfigDir(); err == nil && configDir != "" {
		appDir := filepath.Join(configDir, "MikrotikAutomation")
		_ = os.MkdirAll(appDir, 0755)
		return filepath.Join(appDir, ProfilesConfigFile)
	}

	return ProfilesConfigFile
}

// Derive a 32-byte AES key from host identity and app salt
func deriveKey() []byte {
	host, _ := os.Hostname()
	hash := sha256.Sum256([]byte(host + ":" + appSecretSalt))
	return hash[:]
}

// EncryptString encrypts plain text using AES-256-GCM and returns base64 string.
func EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	key := deriveKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts base64 AES-256-GCM ciphertext to plain text.
func DecryptString(cryptoText string) (string, error) {
	if cryptoText == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}
	key := deriveKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("dữ liệu mã hoá không hợp lệ")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// DeviceProfile represents a saved Router in the Address Book.
type DeviceProfile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	User      string    `json:"user"`
	Password  string    `json:"password"` // Encrypted in JSON
	Proto     string    `json:"proto"`    // "REST API" or "SSH"
	CreatedAt time.Time `json:"created_at"`
}

// ProfileManager manages saving, retrieving, and updating router address book profiles.
type ProfileManager struct {
	configPath string
	profiles   []DeviceProfile
	mu         sync.RWMutex
}

func NewProfileManager() (*ProfileManager, error) {
	path := GetProfilesConfigPath()
	pm := &ProfileManager{
		configPath: path,
		profiles:   make([]DeviceProfile, 0),
	}
	_ = pm.load()
	return pm, nil
}

func (pm *ProfileManager) load() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			pm.profiles = []DeviceProfile{}
			return nil
		}
		return err
	}

	var list []DeviceProfile
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	pm.profiles = list
	return nil
}

func (pm *ProfileManager) save() error {
	dir := filepath.Dir(pm.configPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}
	data, err := json.MarshalIndent(pm.profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pm.configPath, data, 0600)
}

// ListProfiles returns all saved profiles with decrypted passwords.
func (pm *ProfileManager) ListProfiles() []DeviceProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	res := make([]DeviceProfile, len(pm.profiles))
	for i, p := range pm.profiles {
		plainPass, err := DecryptString(p.Password)
		if err == nil {
			p.Password = plainPass
		}
		res[i] = p
	}
	return res
}

// SaveProfile creates or updates a profile with encrypted password.
func (pm *ProfileManager) SaveProfile(prof DeviceProfile) error {
	if prof.Name == "" {
		return errors.New("tên thiết bị / chi nhánh không được để trống")
	}
	if prof.Host == "" {
		return errors.New("địa chỉ IP/Host không được để trống")
	}
	if prof.Port <= 0 {
		if prof.Proto == "SSH" {
			prof.Port = 22
		} else {
			prof.Port = 80
		}
	}
	if prof.User == "" {
		prof.User = "admin"
	}
	if prof.Proto == "" {
		prof.Proto = "REST API"
	}

	encPass, err := EncryptString(prof.Password)
	if err != nil {
		return fmt.Errorf("không thể mã hoá mật khẩu: %w", err)
	}
	prof.Password = encPass

	pm.mu.Lock()
	defer pm.mu.Unlock()

	found := false
	for i, existing := range pm.profiles {
		if existing.Name == prof.Name || (prof.ID != "" && existing.ID == prof.ID) {
			if prof.ID == "" {
				prof.ID = existing.ID
			}
			prof.CreatedAt = existing.CreatedAt
			pm.profiles[i] = prof
			found = true
			break
		}
	}

	if !found {
		if prof.ID == "" {
			prof.ID = fmt.Sprintf("prof-%d", time.Now().UnixNano())
		}
		prof.CreatedAt = time.Now()
		pm.profiles = append(pm.profiles, prof)
	}

	return pm.save()
}

// DeleteProfile removes a profile by name or ID.
func (pm *ProfileManager) DeleteProfile(nameOrID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	newProfiles := make([]DeviceProfile, 0, len(pm.profiles))
	deleted := false
	for _, p := range pm.profiles {
		if p.ID == nameOrID || p.Name == nameOrID {
			deleted = true
			continue
		}
		newProfiles = append(newProfiles, p)
	}
	if !deleted {
		return errors.New("không tìm thấy profile để xoá")
	}
	pm.profiles = newProfiles
	return pm.save()
}

// FindProfile returns a decrypted profile by name.
func (pm *ProfileManager) FindProfile(name string) (*DeviceProfile, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, p := range pm.profiles {
		if p.Name == name {
			plainPass, err := DecryptString(p.Password)
			if err == nil {
				p.Password = plainPass
			}
			return &p, true
		}
	}
	return nil, false
}

// ─────────────────────────────────────────────────────────────────────────────
// Configuration Template Data Structure (JSON Export / Deploy)
// ─────────────────────────────────────────────────────────────────────────────

type NetworkTemplate struct {
	Version     string           `json:"version"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	PPPoE       PPPoETemplate    `json:"pppoe"`
	Bridge      BridgeTemplate   `json:"bridge"`
	VLANs       []vlan.VLANEntry `json:"vlans"`
	PBRRules    []PBRRuleEntry   `json:"pbr_rules,omitempty"`
	CakeQoS     CakeQoSTemplate  `json:"cake_qos,omitempty"`
	SimpleQueues []SimpleQueueEntry `json:"simple_queues,omitempty"`
}

type PPPoETemplate struct {
	WANCount        int    `json:"wan_count"`
	User1           string `json:"user1"`
	Pass1           string `json:"pass1,omitempty"`
	User2           string `json:"user2,omitempty"`
	Pass2           string `json:"pass2,omitempty"`
	EnableRecursive bool   `json:"enable_recursive"`
}

type BridgeTemplate struct {
	BridgeIP   string `json:"bridge_ip"`
	EnableDHCP bool   `json:"enable_dhcp"`
	DNSServers string `json:"dns_servers"`
}

type PBRRuleEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "VLAN_SUBNET" or "DOMAIN_IP"
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	TargetWAN   string `json:"target_wan"` // "WAN1" or "WAN2"
}

type CakeQoSTemplate struct {
	Enabled      bool    `json:"enabled"`
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`
}

type SimpleQueueEntry struct {
	Name          string `json:"name"`
	Target        string `json:"target"`
	UploadLimit   string `json:"upload_limit"`
	DownloadLimit string `json:"download_limit"`
	Comment       string `json:"comment,omitempty"`
}

// ExportTemplateToJSON converts a NetworkTemplate to indented JSON string.
func ExportTemplateToJSON(tmpl NetworkTemplate) ([]byte, error) {
	if tmpl.Version == "" {
		tmpl.Version = "1.0"
	}
	if tmpl.CreatedAt.IsZero() {
		tmpl.CreatedAt = time.Now()
	}
	return json.MarshalIndent(tmpl, "", "  ")
}

// ParseTemplateFromJSON parses and validates a NetworkTemplate from JSON.
func ParseTemplateFromJSON(data []byte) (*NetworkTemplate, error) {
	var tmpl NetworkTemplate
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("file mẫu JSON không hợp lệ: %w", err)
	}
	if tmpl.Bridge.BridgeIP == "" {
		return nil, errors.New("template thiếu thông tin cấu hình Bridge IP")
	}
	return &tmpl, nil
}
