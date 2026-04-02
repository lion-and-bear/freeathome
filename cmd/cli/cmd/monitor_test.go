package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonitorCmd(t *testing.T) {
	// Test that monitor command exists
	assert.NotNil(t, monitorCmd)
	assert.Equal(t, "monitor", monitorCmd.Use)
	assert.Equal(t, "Monitor the free@home system access point via WebSocket", monitorCmd.Short)
}

func TestMonitorCmdFlags(t *testing.T) {
	// Test that monitor command has the expected flags
	flags := monitorCmd.Flags()

	// Check timeout flag
	timeoutFlag := flags.Lookup("timeout")
	assert.NotNil(t, timeoutFlag)
	assert.Equal(t, "30", timeoutFlag.DefValue)

	// Check max reconnection attempts flag
	maxReconnectionFlag := flags.Lookup("max-reconnection-attempts")
	assert.NotNil(t, maxReconnectionFlag)
	assert.Equal(t, "3", maxReconnectionFlag.DefValue)

	// Check exponential backoff flag
	exponentialBackoffFlag := flags.Lookup("exponential-backoff")
	assert.NotNil(t, exponentialBackoffFlag)
	assert.Equal(t, "true", exponentialBackoffFlag.DefValue)

	// Check TLS flags
	tlsFlag := flags.Lookup("tls")
	assert.NotNil(t, tlsFlag)
	assert.Equal(t, "true", tlsFlag.DefValue)

	skipTLSFlag := flags.Lookup("skip-tls-verify")
	assert.NotNil(t, skipTLSFlag)
	assert.Equal(t, "false", skipTLSFlag.DefValue)

	// Check log level flag
	logLevelFlag := flags.Lookup("log-level")
	assert.NotNil(t, logLevelFlag)
	assert.Equal(t, "info", logLevelFlag.DefValue)

	// Check raw flag
	rawFlag := flags.Lookup("raw")
	assert.NotNil(t, rawFlag)
	assert.Equal(t, "false", rawFlag.DefValue)
}

func TestMonitorCmdInRoot(t *testing.T) {
	// Test that monitor command is added to root command
	commands := rootCmd.Commands()
	found := false
	for _, cmd := range commands {
		if cmd.Use == "monitor" {
			found = true
			break
		}
	}
	assert.True(t, found, "Monitor command should be added to root command")
}
