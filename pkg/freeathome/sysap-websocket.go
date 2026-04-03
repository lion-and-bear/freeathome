package freeathome

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

type connection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteControl(messageType int, data []byte, deadline time.Time) error
}

// SystemAccessPointWebSocket represents a web socket connection to a system access point.
type SystemAccessPointWebSocket struct {
	// sysAp is the system access point that the web socket connection is connected to.
	sysAp *SystemAccessPoint
	// waitGroup is used to synchronize the web socket connection and message handling.
	waitGroup sync.WaitGroup
	// onMessageHandled is a callback function that is called when a message is handled.
	onMessageHandled func()
	// reconnectionAttempts tracks the number of failed reconnection attempts
	reconnectionAttempts int
	// maxReconnectionAttempts is the maximum number of reconnection attempts before giving up
	maxReconnectionAttempts int
	// exponentialBackoff controls whether exponential backoff is used between reconnection attempts
	exponentialBackoff bool
	// reconnectionMutex protects access to reconnectionAttempts
	reconnectionMutex sync.Mutex
}

// GetWebSocketUrl constructs a WebSocket URL string for the SystemAccessPoint.
func (ws *SystemAccessPointWebSocket) getWebSocketUrl() string {
	var protocol string
	if ws.sysAp.config.TLSEnabled {
		protocol = "wss"
	} else {
		protocol = "ws"
	}
	return fmt.Sprintf("%s://%s/fhapi/v1/api/ws", protocol, ws.sysAp.config.Hostname)
}

// ConnectWebSocket establishes a web socket connection to the system access point.
func (sysAp *SystemAccessPoint) ConnectWebSocket(ctx context.Context, maxReconnectionAttempts int, exponentialBackoff bool, keepaliveInterval time.Duration) error {
	// Create a new web socket connection
	ws := SystemAccessPointWebSocket{
		sysAp:                   sysAp,
		waitGroup:               sync.WaitGroup{},
		maxReconnectionAttempts: maxReconnectionAttempts,
		exponentialBackoff:      exponentialBackoff,
		reconnectionMutex:       sync.Mutex{},
		reconnectionAttempts:    0,
	}

	// Wait for all processes to finish before returning
	defer ws.waitGroup.Wait()

	// Start the connection loop
	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, stop the connection attempts
			ws.sysAp.config.Logger.Log("context cancelled, stopping web socket connection attempts")
			return ctx.Err()
		default:
			// Check if we've exceeded the maximum reconnection attempts
			ws.reconnectionMutex.Lock()
			currentAttempts := ws.reconnectionAttempts
			ws.reconnectionMutex.Unlock()

			if currentAttempts >= ws.maxReconnectionAttempts {
				ws.sysAp.config.Logger.Error("maximum reconnection attempts exceeded", "attempts", currentAttempts, "max", ws.maxReconnectionAttempts)
				return errors.New("maximum reconnection attempts exceeded")
			}

			// Attempt to establish a web socket connection
			ws.webSocketConnectionLoop(ctx, keepaliveInterval)
		}
	}
}

// webSocketConnectionLoop establishes a web socket connection and starts the message loop.
func (ws *SystemAccessPointWebSocket) webSocketConnectionLoop(ctx context.Context, keepaliveInterval time.Duration) {
	// Add a wait group to ensure all processes are finished before returning
	ws.waitGroup.Add(1)
	defer ws.waitGroup.Done()

	// Create a custom dialer for WebSocket connection
	dialer := websocket.DefaultDialer
	if ws.sysAp.config.TLSEnabled && ws.sysAp.config.SkipTLSVerify {
		dialer = &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
			TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Create a new web socket connection
	basicAuth := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", ws.sysAp.config.Client.UserInfo.Username, ws.sysAp.config.Client.UserInfo.Password))
	conn, _, err := dialer.Dial(ws.getWebSocketUrl(), http.Header{
		"Authorization": []string{fmt.Sprintf("Basic %s", basicAuth)},
	})

	// Check for errors
	if err != nil {
		// Safely increment reconnection attempts
		ws.reconnectionMutex.Lock()
		ws.reconnectionAttempts++
		currentAttempts := ws.reconnectionAttempts
		ws.reconnectionMutex.Unlock()

		// Prepare error message with backoff information
		errorAttrs := []any{"error", err, "attempt", currentAttempts, "max", ws.maxReconnectionAttempts}
		if ws.exponentialBackoff && currentAttempts < ws.maxReconnectionAttempts {
			backoffDuration := calculateBackoffDuration(currentAttempts)
			errorAttrs = append(errorAttrs, "backoff", backoffDuration)
		}

		ws.sysAp.config.Logger.Error("failed to connect to web socket", errorAttrs...)
		ws.sysAp.emitError(err)

		// Apply exponential backoff if enabled and we haven't exceeded max attempts
		if ws.exponentialBackoff && currentAttempts < ws.maxReconnectionAttempts {
			backoffDuration := calculateBackoffDuration(currentAttempts)
			select {
			case <-ctx.Done():
				return
			case <-ws.sysAp.clock.After(backoffDuration):
				// Continue to next attempt
			}
		}

		return
	}

	// Create connection channels
	messageReceivedChannel := make(chan struct{}, 1)
	webSocketMessageChannel := make(chan []byte, 10)
	defer func() {
		close(messageReceivedChannel)
		close(webSocketMessageChannel)
	}()

	// Start keepalive and message handler goroutines
	go ws.webSocketKeepaliveLoop(messageReceivedChannel, conn, keepaliveInterval)
	go ws.webSocketMessageHandler(webSocketMessageChannel)

	// Reset reconnection attempts on successful connection
	ws.reconnectionMutex.Lock()
	ws.reconnectionAttempts = 0
	ws.reconnectionMutex.Unlock()

	// Start the message loop
	ws.sysAp.config.Logger.Log("web socket connected successfully, starting message loop")
	err = ws.webSocketMessageLoop(ctx, messageReceivedChannel, webSocketMessageChannel, conn)

	// Check for errors
	if err != nil {
		ws.sysAp.config.Logger.Error("web socket message loop failed", "error", err)
		ws.sysAp.emitError(err)
	}

	// Close the web socket connection
	err = conn.Close()
	ws.sysAp.config.Logger.Debug("web socket connection closed", "error", err)
}

// webSocketMessageLoop starts a loop to read messages from the web socket connection.
func (ws *SystemAccessPointWebSocket) webSocketMessageLoop(ctx context.Context, messageReceivedChannel chan<- struct{}, webSocketMessageChannel chan<- []byte, conn connection) error {
	// Verify that the connection channels are not nil
	if webSocketMessageChannel == nil || messageReceivedChannel == nil {
		errorMessage := "a connection channel is nil, cannot start message loop"
		ws.sysAp.config.Logger.Error(errorMessage)
		return errors.New(errorMessage)
	}

	// Start a loop to read messages from the web socket
	for {
		select {
		case <-ctx.Done():
			// If the context is cancelled, stop the message loop
			ws.sysAp.config.Logger.Log("context cancelled, stopping message loop")
			return nil
		default:
			// Read messages from the web socket
			messageType, message, err := conn.ReadMessage()

			// Check for errors
			if err != nil {
				ws.sysAp.emitError(err)
				return err
			}

			// Signal that a message has been received
			select {
			case messageReceivedChannel <- struct{}{}:
				// Message sent successfully
			case <-ctx.Done():
				// Context cancelled, exit immediately
				return ctx.Err()
			}

			// Check if the message type is text
			if messageType != websocket.TextMessage {
				ws.sysAp.config.Logger.Warn("received non-text message from web socket", "type", messageType, "message", string(message))
				continue
			}

			// Pipe the message to the message handler
			if ws.sysAp.config.WebSocketRawOutput == nil {
				ws.sysAp.config.Logger.Debug("received text message from web socket")
			}
			select {
			case webSocketMessageChannel <- message:
				// Message sent successfully
			case <-ctx.Done():
				// Context cancelled, exit immediately
				return ctx.Err()
			}
		}
	}
}

// processWebSocketMessage processes a message received from the web socket connection.
func (ws *SystemAccessPointWebSocket) webSocketMessageHandler(webSocketMessageChannel <-chan []byte) {
	// Add a wait group to ensure all processes are finished before returning
	ws.waitGroup.Add(1)
	defer ws.waitGroup.Done()

	// Verify that the webSocketMessageChannel is not nil
	if webSocketMessageChannel == nil {
		ws.sysAp.config.Logger.Error("webSocketMessageChannel is nil, cannot start message handler")
		return
	}

	// Start a loop to handle messages from the web socket
	for message := range webSocketMessageChannel {
		ws.processMessage(message)
	}

	// If the channel is closed, exit the loop
	ws.sysAp.config.Logger.Log("webSocketMessageChannel closed, stopping message handler")
}

func (ws *SystemAccessPointWebSocket) webSocketKeepaliveLoop(messageReceivedChannel <-chan struct{}, conn connection, interval time.Duration) {
	// Add a wait group to ensure all processes are finished before returning
	ws.waitGroup.Add(1)
	defer ws.waitGroup.Done()

	// Verify that the messageReceivedChannel is not nil
	if messageReceivedChannel == nil {
		ws.sysAp.config.Logger.Error("messageReceivedChannel is nil, cannot start keepalive loop")
		return
	}

	// Create a ticker for the keepalive interval
	timer := time.NewTicker(interval)
	defer timer.Stop()

	for {
		select {
		case _, ok := <-messageReceivedChannel:
			if ok {
				// Reset the timer when a message is received
				ws.sysAp.config.Logger.Debug("message received, resetting keepalive timer")
				timer.Reset(interval)
			} else {
				// If the channel is closed, exit the loop
				ws.sysAp.config.Logger.Log("messageReceivedChannel closed, stopping keepalive loop")
				return
			}
		case <-timer.C:
			// Send a ping message to the server
			ws.sysAp.config.Logger.Log("keepalive timer expired, sending ping message...")
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(3*time.Second))
			if err != nil {
				ws.sysAp.config.Logger.Error("failed to send ping message", "error", err)
				ws.sysAp.emitError(err)
				return
			}
		}
	}
}

func (ws *SystemAccessPointWebSocket) processMessage(message []byte) {
	defer func() {
		// Call the onMessageHandled callback if it is set
		if ws.onMessageHandled != nil {
			ws.onMessageHandled()
		}
	}()

	if w := ws.sysAp.config.WebSocketRawOutput; w != nil {
		out := append(append([]byte(nil), message...), '\n')
		if _, err := w.Write(out); err != nil {
			ws.sysAp.config.Logger.Error("failed to write raw web socket message", "error", err)
			ws.sysAp.emitError(err)
		}
		return
	}

	// Unmarshal the message into a WebSocketMessage struct
	var msg models.WebSocketMessage
	err := json.Unmarshal(message, &msg)

	if err != nil {
		ws.sysAp.config.Logger.Error("failed to unmarshal message", "error", err)
		ws.sysAp.emitError(err)
		return
	}

	// Check if the message is empty
	if len(msg[models.EmptyUUID].Datapoints) == 0 {
		ws.sysAp.config.Logger.Warn("web socket message has no datapoints")
		return
	}

	// Process data point updates
	for key, datapoint := range msg[models.EmptyUUID].Datapoints {
		// Check if the key matches the expected format
		if !ws.sysAp.datapointRegex.MatchString(key) {
			ws.sysAp.config.Logger.Warn(`Ignored datapoint with invalid key format`, "key", key)
			continue
		}

		// Log the datapoint update
		ws.sysAp.config.Logger.Log("data point update",
			"device", ws.sysAp.datapointRegex.FindStringSubmatch(key)[1],
			"channel", ws.sysAp.datapointRegex.FindStringSubmatch(key)[2],
			"datapoint", ws.sysAp.datapointRegex.FindStringSubmatch(key)[3],
			"value", datapoint,
		)
	}
}

// calculateBackoffDuration calculates the exponential backoff duration for a given attempt number.
// The backoff follows the formula: baseDelay * (2^attempt) with a maximum cap.
func calculateBackoffDuration(attempt int) time.Duration {
	baseDelay := time.Second
	maxDelay := 30 * time.Second

	// Calculate exponential backoff: baseDelay * (2^attempt)
	backoffDuration := min(
		baseDelay*time.Duration(1<<attempt),
		maxDelay,
	)

	return backoffDuration
}
