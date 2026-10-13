package auth

import (
	"encoding/json"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const (
	AuthConfigFile = "auth.json"
	DefaultUser    = "admin"
	DefaultPass    = "Loveisdie1@"
)

type Account struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
}

type AuthManager struct {
	accounts []Account
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// LoadOrCreateAuthManager loads auth.json or creates it with default admin:Loveisdie1@
func LoadOrCreateAuthManager() (*AuthManager, error) {
	am := &AuthManager{accounts: []Account{}}

	if _, err := os.Stat(AuthConfigFile); os.IsNotExist(err) {
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

	data, err := os.ReadFile(AuthConfigFile)
	if err != nil {
		return nil, err
	}

	var accs []Account
	if err := json.Unmarshal(data, &accs); err != nil {
		// If file corrupted, recreate default
		hash, _ := HashPassword(DefaultPass)
		am.accounts = append(am.accounts, Account{
			Username:     DefaultUser,
			PasswordHash: hash,
		})
		_ = am.save()
		return am, nil
	}

	am.accounts = accs
	return am, nil
}

func (am *AuthManager) save() error {
	data, err := json.MarshalIndent(am.accounts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(AuthConfigFile, data, 0600)
}

func (am *AuthManager) Authenticate(username, password string) bool {
	for _, acc := range am.accounts {
		if acc.Username == username {
			return CheckPasswordHash(password, acc.PasswordHash)
		}
	}
	return false
}

