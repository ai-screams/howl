package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AccountInfo represents the logged-in Claude Code account from ~/.claude.json
type AccountInfo struct {
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
	AccountUUID  string `json:"accountUuid"`
}

// GetAccountInfo reads the active account from ~/.claude.json.
// Returns nil on any error or if email is empty (graceful degradation).
// No caching is used because Howl runs as a subprocess each time.
func GetAccountInfo() *AccountInfo {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	configPath := filepath.Join(home, ".claude.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil
	}

	var config struct {
		OAuthAccount AccountInfo `json:"oauthAccount"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}

	// Return nil if emailAddress is empty (no usable account data)
	if config.OAuthAccount.EmailAddress == "" {
		return nil
	}

	return &config.OAuthAccount
}
