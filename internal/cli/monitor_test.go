package cli

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/lion-and-bear/freeathome/v2/pkg/freeathome"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestApplyMonitorRawMode(t *testing.T) {
	t.Run("raw enables stdout writer and stderr hints", func(t *testing.T) {
		sysAp := freeathome.MustNewSystemAccessPoint(freeathome.NewConfig("localhost", "u", "p"))
		hintOut := applyMonitorRawMode(sysAp, true)
		assert.Equal(t, os.Stderr, hintOut)
		assert.Equal(t, os.Stdout, sysAp.GetWebSocketRawOutput())
	})
	t.Run("not raw leaves writer nil and stdout hints", func(t *testing.T) {
		sysAp := freeathome.MustNewSystemAccessPoint(freeathome.NewConfig("localhost", "u", "p"))
		hintOut := applyMonitorRawMode(sysAp, false)
		assert.Equal(t, os.Stdout, hintOut)
		assert.Nil(t, sysAp.GetWebSocketRawOutput())
	})
}

func TestMonitorWithMockConnect(t *testing.T) {
	oldConn := monitorConnectWebSocket
	t.Cleanup(func() { monitorConnectWebSocket = oldConn })
	oldSetup := setupFunc
	t.Cleanup(func() { setupFunc = oldSetup })

	baseCfg := func(raw bool) MonitorCommandConfig {
		return MonitorCommandConfig{
			CommandConfig: CommandConfig{
				Viper:         viper.New(),
				TLSEnabled:    false,
				SkipTLSVerify: false,
				LogLevel:      "info",
			},
			Timeout:                 30,
			MaxReconnectionAttempts: 1,
			ExponentialBackoff:      false,
			Raw:                     raw,
		}
	}

	t.Run("raw sets WebSocket output", func(t *testing.T) {
		sysAp := freeathome.MustNewSystemAccessPoint(freeathome.NewConfig("localhost", "u", "p"))
		setupFunc = func(CommandConfig, string) (*freeathome.SystemAccessPoint, error) {
			return sysAp, nil
		}
		monitorConnectWebSocket = func(ctx context.Context, sa *freeathome.SystemAccessPoint, c MonitorCommandConfig) error {
			return nil
		}
		err := Monitor(baseCfg(true))
		require.NoError(t, err)
		assert.Equal(t, os.Stdout, sysAp.GetWebSocketRawOutput())
	})

	t.Run("connect error propagates", func(t *testing.T) {
		sysAp := freeathome.MustNewSystemAccessPoint(freeathome.NewConfig("localhost", "u", "p"))
		setupFunc = func(CommandConfig, string) (*freeathome.SystemAccessPoint, error) {
			return sysAp, nil
		}
		monitorConnectWebSocket = func(ctx context.Context, sa *freeathome.SystemAccessPoint, c MonitorCommandConfig) error {
			return errors.New("dial refused")
		}
		err := Monitor(baseCfg(false))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dial refused")
	})
}
