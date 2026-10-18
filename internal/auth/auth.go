package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

const (
	AuthConfigFile = "auth.json"
	DefaultUser    = "admin"
	DefaultPass    = "Loveisdie1@"
)

var (
	customConfigPath string
	muPath           sync.RWMutex
)

// SetCustomConfigPath sets a custom path for auth.json (primarily for isolated testing).
func SetCustomConfigPath(path string) {
	muPath.Lock()
	defer muPath.Unlock()
	customConfigPath = path
}

// GetAuthConfigPath determines the optimal, persistent location for auth.json.
// Priority:
// 1. Explicitly set custom path (if any).
// 2. Directory next to the executable (portable mode).
// 3. Fallback to %APPDATA%/MikrotikAutomation/auth.json if executable dir is read-only.
func GetAuthConfigPath() string {
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
		target := filepath.Join(exeDir, AuthConfigFile)

		// Test if writable or file already exists
		if _, err := os.Stat(target); err == nil {
			return target
		}
		testFile := filepath.Join(exeDir, ".write_test_"+fmt.Sprintf("%d", os.Getpid()))
		if err := os.WriteFile(testFile, []byte("ok"), 0600); err == nil {
			_ = os.Remove(testFile)
			return target
		}
	}

	// Fallback to User Config Dir (AppData on Windows)
	if configDir, err := os.UserConfigDir(); err == nil && configDir != "" {
		appDir := filepath.Join(configDir, "MikrotikAutomation")
		_ = os.MkdirAll(appDir, 0755)
		return filepath.Join(appDir, AuthConfigFile)
	}

	// Ultimate fallback to current working directory
	return AuthConfigFile
}

type Account struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type AuthManager struct {
	configPath string
	accounts   []Account
	mu         sync.RWMutex
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// LoadOrCreateAuthManager loads auth config from the resolved path or creates it with default credentials.
func LoadOrCreateAuthManager() (*AuthManager, error) {
	configPath := GetAuthConfigPath()
	am := &AuthManager{
		configPath: configPath,
		accounts:   []Account{},
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Initialize with default admin
		hash, err := HashPassword(DefaultPass)
		if err != nil {
			return nil, err
		}
		am.accounts = append(am.accounts, Account{
			Username:     DefaultUser,
			PasswordHash: hash,
		})
		_ = am.save()
		return am, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var accs []Account
	if err := json.Unmarshal(data, &accs); err != nil || len(accs) == 0 {
		// If corrupted or empty, recreate default
		hash, _ := HashPassword(DefaultPass)
		am.accounts = []Account{
			{
				Username:     DefaultUser,
				PasswordHash: hash,
			},
		}
		_ = am.save()
		return am, nil
	}

	am.accounts = accs
	return am, nil
}

func (am *AuthManager) save() error {
	am.mu.RLock()
	data, err := json.MarshalIndent(am.accounts, "", "  ")
	path := am.configPath
	am.mu.RUnlock()

	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}
	return os.WriteFile(path, data, 0600)
}

func (am *AuthManager) Authenticate(username, password string) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, acc := range am.accounts {
		if acc.Username == username {
			return CheckPasswordHash(password, acc.PasswordHash)
		}
	}
	return false
}

// ChangePassword verifies the old password and updates it with the new password.
func (am *AuthManager) ChangePassword(username, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("mật khẩu mới phải có tối thiểu 6 ký tự")
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	found := false
	for i, acc := range am.accounts {
		if acc.Username == username {
			found = true
			if !CheckPasswordHash(oldPassword, acc.PasswordHash) {
				return errors.New("mật khẩu cũ không chính xác")
			}
			newHash, err := HashPassword(newPassword)
			if err != nil {
				return fmt.Errorf("không thể băm mật khẩu mới: %w", err)
			}
			am.accounts[i].PasswordHash = newHash
			break
		}
	}

	if !found {
		return errors.New("tài khoản không tồn tại")
	}

	// Persist changes
	data, err := json.MarshalIndent(am.accounts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(am.configPath, data, 0600)
}
