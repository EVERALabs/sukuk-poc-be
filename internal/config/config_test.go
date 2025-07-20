package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test with environment variables
	os.Setenv("APP_PORT", "9090")
	os.Setenv("BLOCKCHAIN_CHAIN_ID", "84532")
	os.Setenv("API_API_KEY", "test-key")
	
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("BLOCKCHAIN_CHAIN_ID")
		os.Unsetenv("API_API_KEY")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if config.App.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.App.Port)
	}
	
	if config.Blockchain.ChainID != 84532 {
		t.Errorf("Expected chain ID 84532, got %d", config.Blockchain.ChainID)
	}
	
	if config.API.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", config.API.APIKey)
	}
}

func TestBaseTestnetDefaults(t *testing.T) {
	os.Setenv("API_API_KEY", "test-key")
	defer os.Unsetenv("API_API_KEY")

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify Base Testnet configuration
	if config.Blockchain.ChainID != 84532 {
		t.Errorf("Expected chain ID 84532, got %d", config.Blockchain.ChainID)
	}
	
	if config.Blockchain.RPCEndpoint != "https://sepolia.base.org" {
		t.Errorf("Expected RPC endpoint 'https://sepolia.base.org', got '%s'", config.Blockchain.RPCEndpoint)
	}
}

func TestConfigValidation(t *testing.T) {
	// Test invalid port
	os.Setenv("APP_PORT", "99999")
	os.Setenv("API_API_KEY", "test-key")
	
	defer func() {
		os.Unsetenv("APP_PORT")
		os.Unsetenv("API_API_KEY")
	}()

	_, err := Load()
	if err == nil {
		t.Error("Expected validation error for invalid port")
	}
}