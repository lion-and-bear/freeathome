package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// MonitorCommandConfig is a struct that contains the configuration for the monitor command
type MonitorCommandConfig struct {
	CommandConfig
	Timeout                 int
	MaxReconnectionAttempts int
	ExponentialBackoff      bool
	// Raw, when true, streams WebSocket text frames to stdout; user hints go to stderr.
	Raw bool
}

// Monitor connects to the free@home system access point via WebSocket and monitors real-time events
func Monitor(config MonitorCommandConfig) error {
	// Setup system access point
	sysAp, err := setupFunc(config.CommandConfig, "")
	if err != nil {
		return err
	}

	var hintOut io.Writer = os.Stdout
	if config.Raw {
		sysAp.SetWebSocketRawOutput(os.Stdout)
		hintOut = os.Stderr
	}

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create error channel for the shutdown
	shutdown := make(chan error, 1)

	// Setup keypress handling for graceful shutdown
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				char, _, err := reader.ReadRune()
				if err != nil {
					break
				}
				if char == 'q' || char == 'Q' {
					// Send SIGINT to trigger graceful shutdown
					sigs <- syscall.SIGINT
					return
				}
			}
		}
	}()

	go func() {
		// First signal triggers graceful shutdown
		<-sigs
		_, _ = fmt.Fprintln(hintOut, "Exit signal received, shutting down gracefully...")
		_, _ = fmt.Fprintln(hintOut, "Press Ctrl+C to force exit")
		cancel()

		// Second signal triggers immediate, forced shutdown
		<-sigs
		_, _ = fmt.Fprintln(hintOut, "\nSecond exit signal received, shutting down immediately...")
		shutdown <- fmt.Errorf("forced shutdown requested")
	}()

	_, _ = fmt.Fprintln(hintOut, "Press 'q' or Ctrl+C to exit")

	// Connect to the system access point websocket
	timeout := time.Duration(config.Timeout) * time.Second
	go func() {
		shutdown <- sysAp.ConnectWebSocket(ctx, config.MaxReconnectionAttempts, config.ExponentialBackoff, timeout)
	}()

	// Handle both forced shutdown and WebSocket connection errors
	err = <-shutdown
	if err != nil && err != context.Canceled {
		return err
	}

	return nil
}
