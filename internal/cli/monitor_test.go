package cli

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMonitorCommandConfig(t *testing.T) {
	// Test MonitorCommandConfig struct
	config := MonitorCommandConfig{
		CommandConfig: CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    true,
			SkipTLSVerify: false,
			LogLevel:      "debug",
		},
		Timeout:                 30,
		MaxReconnectionAttempts: 3,
		ExponentialBackoff:      true,
		Raw:                     false,
	}

	assert.NotNil(t, config.Viper)
	assert.True(t, config.TLSEnabled)
	assert.False(t, config.SkipTLSVerify)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, 30, config.Timeout)
	assert.Equal(t, 3, config.MaxReconnectionAttempts)
	assert.True(t, config.ExponentialBackoff)
	assert.False(t, config.Raw)
}

func TestSetupMonitorWithInvalidConfig(t *testing.T) {
	// Test setupMonitor with invalid configuration
	config := MonitorCommandConfig{
		CommandConfig: CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    true,
			SkipTLSVerify: false,
			LogLevel:      "debug",
		},
	}

	// This should fail because no configuration is loaded
	_, err := setupFunc(config.CommandConfig, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hostname not configured")
}
