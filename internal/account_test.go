package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetAccountInfo(t *testing.T) {

	t.Run("valid config with all fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".claude.json")

		config := `{
			"oauthAccount": {
				"emailAddress": "test@example.com",
				"displayName": "Test User",
				"accountUuid": "12345-abcde"
			}
		}`

		if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
			t.Fatal(err)
		}

		t.Setenv("HOME", tmpDir)

		account := GetAccountInfo()
		if account == nil {
			t.Fatal("expected account, got nil")
		}
		if account.EmailAddress != "test@example.com" {
			t.Errorf("EmailAddress = %q, want %q", account.EmailAddress, "test@example.com")
		}
		if account.DisplayName != "Test User" {
			t.Errorf("DisplayName = %q, want %q", account.DisplayName, "Test User")
		}
		if account.AccountUUID != "12345-abcde" {
			t.Errorf("AccountUUID = %q, want %q", account.AccountUUID, "12345-abcde")
		}
	})

	t.Run("file not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		account := GetAccountInfo()
		if account != nil {
			t.Errorf("expected nil for missing file, got %+v", account)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".claude.json")

		if err := os.WriteFile(configPath, []byte("{invalid json}"), 0600); err != nil {
			t.Fatal(err)
		}

		t.Setenv("HOME", tmpDir)

		account := GetAccountInfo()
		if account != nil {
			t.Errorf("expected nil for invalid JSON, got %+v", account)
		}
	})

	t.Run("missing oauthAccount field returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".claude.json")

		config := `{"otherField": "value"}`

		if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
			t.Fatal(err)
		}

		t.Setenv("HOME", tmpDir)

		account := GetAccountInfo()
		if account != nil {
			t.Errorf("expected nil for missing oauthAccount, got %+v", account)
		}
	})

	t.Run("empty emailAddress returns nil", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".claude.json")

		config := `{
			"oauthAccount": {
				"emailAddress": "",
				"displayName": "No Email User"
			}
		}`

		if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
			t.Fatal(err)
		}

		t.Setenv("HOME", tmpDir)

		account := GetAccountInfo()
		if account != nil {
			t.Errorf("expected nil for empty email, got %+v", account)
		}
	})

	t.Run("partial fields are valid", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".claude.json")

		config := `{
			"oauthAccount": {
				"emailAddress": "partial@example.com"
			}
		}`

		if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
			t.Fatal(err)
		}

		t.Setenv("HOME", tmpDir)

		account := GetAccountInfo()
		if account == nil {
			t.Fatal("expected account, got nil")
		}
		if account.EmailAddress != "partial@example.com" {
			t.Errorf("EmailAddress = %q, want %q", account.EmailAddress, "partial@example.com")
		}
		if account.DisplayName != "" {
			t.Errorf("DisplayName = %q, want empty", account.DisplayName)
		}
		if account.AccountUUID != "" {
			t.Errorf("AccountUUID = %q, want empty", account.AccountUUID)
		}
	})
}
