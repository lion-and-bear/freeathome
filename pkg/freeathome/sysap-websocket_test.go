package freeathome

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

const testMessageValid = "valid message"

// Duplicated (or production-matched) literals for assertions — keep in sync with sysap-websocket.go log/error text.
const (
	testDatapointKeyValid = "ABB7F595EC47/ch0000/odp0000"

	logMsgDataPointUpdate             = "data point update"
	logMsgFailedUnmarshal             = "failed to unmarshal message"
	logMsgNoDatapoints                = "web socket message has no datapoints"
	logMsgIgnoredInvalidDatapointKey  = "Ignored datapoint with invalid key format"
	logMsgFailedWriteRaw              = "failed to write raw web socket message"
	logMsgReceivedTextMessage         = "received text message from web socket"
	logMsgWebSocketConnected          = "web socket connected successfully"
	logMsgTLSNotRecommended           = "this is not recommended"
	logMsgWebSocketMessageHandlerDone = "webSocketMessageChannel closed, stopping message handler"
	logMsgContextCancelledWSAttempts  = "context cancelled, stopping web socket connection attempts"
	logMsgFailedConnectWebSocket      = "failed to connect to web socket"
	logMsgMaxReconnectionExceeded     = "maximum reconnection attempts exceeded"
	logMsgReceivedNonText             = "received non-text message from web socket"
	logMsgKeepalivePing               = "keepalive timer expired, sending ping"
	logMsgWebSocketMessageChannelNil  = "webSocketMessageChannel is nil"
	logMsgMessageReceivedChannelNil   = "messageReceivedChannel is nil, cannot start keepalive loop"

	errMsgConnectionChannelNil       = "a connection channel is nil, cannot start message loop"
	errMsgWriteFailed                = "write failed"
	errMsgNoMoreMessages             = "no more messages"
	errMsgExpectedLogOutputToContain = "Expected log output to contain %q, got: %s"
	errMsgWebSocketUpgradeFailed     = "Failed to upgrade WebSocket: %v"
	errMsgExpectedNoErrorGot         = "Expected no error, got: %v"
	errMsgUnexpectedErrorGot         = "Unexpected error: %v"

	lookupInvalidHostSubstr = "lookup invalid-host"

	invalidHost        = "invalid-host"
	invalidJSONPayload = "invalid json"
	notJSONPayload     = "not json"
)

const (
	wsURLHTTP  = "ws://localhost/fhapi/v1/api/ws"
	wsURLHTTPS = "wss://localhost/fhapi/v1/api/ws"
)

// TestSystemAccessPointWebSocketMessageHandler tests the webSocketMessageHandler method of SystemAccessPoint.
func TestSystemAccessPointWebSocketMessageHandler(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)
	defer ws.waitGroup.Wait()

	webSocketMessageChannel := make(chan []byte, 10)

	// Mock a valid WebSocketMessage
	validMessage := models.WebSocketMessage{
		models.EmptyUUID: models.Message{
			Datapoints: map[string]string{
				testDatapointKeyValid: "1",
			},
		},
	}
	validMessageBytes, _ := json.Marshal(validMessage)

	// Mock an invalid WebSocketMessage
	invalidMessage := []byte(invalidJSONPayload)

	// Mock a WebSocketMessage with no datapoints
	emptyMessage := models.WebSocketMessage{
		models.EmptyUUID: models.Message{
			Datapoints: map[string]string{},
		},
	}
	emptyMessageBytes, _ := json.Marshal(emptyMessage)

	// Mock a WebSocketMessage with invalid datapoint format
	invalidFormatMessage := models.WebSocketMessage{
		models.EmptyUUID: models.Message{
			Datapoints: map[string]string{
				"Test123": "1",
			},
		},
	}
	invalidFormatMessageBytes, _ := json.Marshal(invalidFormatMessage)

	// Send messages to the WebSocketMessageChannel
	var wg sync.WaitGroup
	wg.Add(4)
	ws.onMessageHandled = wg.Done
	go func() {
		webSocketMessageChannel <- validMessageBytes
		webSocketMessageChannel <- invalidMessage
		webSocketMessageChannel <- emptyMessageBytes
		webSocketMessageChannel <- invalidFormatMessageBytes
		wg.Wait()
		close(webSocketMessageChannel)
		webSocketMessageChannel = nil
	}()

	// Start the handler
	ws.webSocketMessageHandler(webSocketMessageChannel)

	// Check the log output
	logOutput := buf.String()

	// Verify valid message processing
	if !strings.Contains(logOutput, logMsgDataPointUpdate) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgDataPointUpdate, logOutput)
	}

	// Verify invalid message handling
	if !strings.Contains(logOutput, logMsgFailedUnmarshal) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgFailedUnmarshal, logOutput)
	}

	// Verify empty message handling
	if !strings.Contains(logOutput, logMsgNoDatapoints) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgNoDatapoints, logOutput)
	}

	// Verify invalid format message handling
	if !strings.Contains(logOutput, logMsgIgnoredInvalidDatapointKey) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgIgnoredInvalidDatapointKey, logOutput)
	}
}

func TestProcessMessageWebSocketRawOutputWriteError(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)

	errSeen := 0
	ws.sysAp.onError = func(err error) {
		if err != nil && err.Error() == errMsgWriteFailed {
			errSeen++
		}
	}
	ws.sysAp.SetWebSocketRawOutput(failWriter{})
	ws.processMessage([]byte(`{}`))

	if errSeen != 1 {
		t.Errorf("expected onError once for write failure, got %d", errSeen)
	}
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgFailedWriteRaw) {
		t.Errorf("expected log about raw write failure, got: %s", logOutput)
	}
}

type failWriter struct{}

func (failWriter) Write(_ []byte) (n int, err error) {
	return 0, errors.New(errMsgWriteFailed)
}

func TestWebSocketMessageLoopTextMessageRawSkipsDebugLog(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ws, buf, _ := setupSysApWebSocket(t, true, false)
	ws.sysAp.SetWebSocketRawOutput(&bytes.Buffer{})

	webSocketMessageChannel := make(chan []byte, 10)
	messageReceivedChannel := make(chan struct{}, 1)

	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte(testMessageValid),
		err:         nil,
	}

	go func() {
		err := ws.webSocketMessageLoop(ctx, messageReceivedChannel, webSocketMessageChannel, conn)
		if err == nil {
			t.Error(expectedErrorGotNil)
		}
	}()

	message := <-webSocketMessageChannel
	cancel()
	<-ctx.Done()
	close(webSocketMessageChannel)
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

	if string(message) != testMessageValid {
		t.Errorf("Expected message %q, got: %s", testMessageValid, string(message))
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, logMsgReceivedTextMessage) {
		t.Errorf("did not expect debug log when raw output is set, got: %s", logOutput)
	}
}

func TestProcessMessageWebSocketRawOutput(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)

	var rawBuf bytes.Buffer
	ws.sysAp.SetWebSocketRawOutput(&rawBuf)

	validMessage := models.WebSocketMessage{
		models.EmptyUUID: models.Message{
			Datapoints: map[string]string{
				testDatapointKeyValid: "1",
			},
		},
	}
	validMessageBytes, err := json.Marshal(validMessage)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	ws.processMessage(validMessageBytes)

	want := string(validMessageBytes) + "\n"
	if rawBuf.String() != want {
		t.Errorf("raw output: got %q want %q", rawBuf.String(), want)
	}
	logOutput := buf.String()
	if strings.Contains(logOutput, logMsgDataPointUpdate) {
		t.Errorf("expected no datapoint log in raw mode, got: %s", logOutput)
	}

	// Invalid JSON is still written as-is; no unmarshal error log in raw mode
	rawBuf.Reset()
	buf.Reset()
	ws.processMessage([]byte(notJSONPayload))
	if rawBuf.String() != notJSONPayload+"\n" {
		t.Errorf("raw invalid JSON: got %q", rawBuf.String())
	}
	if strings.Contains(buf.String(), logMsgFailedUnmarshal) {
		t.Errorf("expected no unmarshal error log in raw mode, got: %s", buf.String())
	}
}

func TestSystemAccessPointWebSocketMessageHandlerMissingChannel(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)
	defer ws.waitGroup.Wait()

	// Start the handler
	ws.webSocketMessageHandler(nil)

	// Check the log output
	logOutput := buf.String()

	if !strings.Contains(logOutput, logMsgWebSocketMessageChannelNil) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgWebSocketMessageChannelNil, logOutput)
	}
}

// TestSystemAccessPointConnectWebSocketSuccess tests the successful connection of the WebSocket.
func TestSystemAccessPointConnectWebSocketSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sysAp, _, records := setupSysAp(t, false, false)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	var conn *websocket.Conn
	var connMutex sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		connMutex.Lock()
		conn, err = upgrader.Upgrade(w, r, nil)
		connMutex.Unlock()
		if err != nil {
			t.Fatalf(errMsgWebSocketUpgradeFailed, err)
		}
	}))
	defer server.Close()

	sysAp.config.Hostname = strings.TrimPrefix(server.URL, "http://")

	// Wait for the expected record in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case record := <-records:
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, logMsgWebSocketConnected) {
					cancel()
					connMutex.Lock()
					if conn != nil {
						_ = conn.Close()
					}
					connMutex.Unlock()
				}
			}
		}
	}()

	// Run ConnectWebSocket in a separate goroutine
	err := sysAp.ConnectWebSocket(ctx, 1, false, 1*time.Hour)
	if err != nil && err != context.Canceled {
		t.Errorf(errMsgExpectedNoErrorGot, err)
	}
}

// TestSystemAccessPointConnectWebSocketSkipTlsVerify tests the successful connection of the WebSocket with skip TLS verify.
func TestSystemAccessPointConnectWebSocketSkipTlsVerify(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sysAp, buf, records := setupSysAp(t, true, true)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	var conn *websocket.Conn
	var connMutex sync.Mutex
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		connMutex.Lock()
		conn, err = upgrader.Upgrade(w, r, nil)
		connMutex.Unlock()
		if err != nil {
			t.Fatalf(errMsgWebSocketUpgradeFailed, err)
		}
	}))
	defer server.Close()

	sysAp.config.Hostname = strings.TrimPrefix(server.URL, "https://")

	// Wait for the expected record in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Check the log output safely
				logOutput := buf.String()
				if !strings.Contains(logOutput, logMsgTLSNotRecommended) {
					t.Errorf(errMsgExpectedLogOutputToContain, logMsgTLSNotRecommended, logOutput)
				}
				return
			case record := <-records:
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, logMsgWebSocketConnected) {
					cancel()
					connMutex.Lock()
					if conn != nil {
						_ = conn.Close()
					}
					connMutex.Unlock()
				}
			}
		}
	}()

	// Run ConnectWebSocket in a separate goroutine
	err := sysAp.ConnectWebSocket(ctx, 1, false, 1*time.Hour)
	if err != nil && err != context.Canceled {
		t.Errorf(errMsgExpectedNoErrorGot, err)
	}
}

// TestSystemAccessPointConnectWebSocketContextCancelled tests the behavior when the context is cancelled.
func TestSystemAccessPointConnectWebSocketContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	sysAp, _, records := setupSysAp(t, false, false)

	// Mock the WebSocket connection
	dialer := &websocket.Dialer{}
	websocket.DefaultDialer = dialer

	// Mock the WebSocket server
	var conn *websocket.Conn
	var connMutex sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		var err error
		connMutex.Lock()
		conn, err = upgrader.Upgrade(w, r, nil)
		connMutex.Unlock()
		if err != nil {
			t.Fatalf(errMsgWebSocketUpgradeFailed, err)
		}
	}))
	defer server.Close()

	sysAp.config.Hostname = strings.TrimPrefix(server.URL, "http://")

	wg := sync.WaitGroup{}
	wg.Add(2)

	innerCtx, innerCancel := context.WithCancel(context.TODO())
	go func() {
		for {
			select {
			case <-innerCtx.Done():
				return
			case record := <-records:
				// Cancel the context when the web socket is connected successfully
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, logMsgWebSocketConnected) {
					cancel()
					connMutex.Lock()
					if conn != nil {
						_ = conn.WriteMessage(websocket.TextMessage, []byte("test"))
					}
					connMutex.Unlock()
					break
				}
				// Send one done when the message handler is stopped
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, logMsgWebSocketMessageHandlerDone) {
					wg.Done()
					break
				}
				// Send one done when the web socket connection is stopped
				if record.Level == slog.LevelInfo && strings.Contains(record.Message, logMsgContextCancelledWSAttempts) {
					wg.Done()
				}
			}
		}
	}()

	// Run ConnectWebSocket in a separate goroutine
	go func() {
		err := sysAp.ConnectWebSocket(ctx, 1, false, 1*time.Hour)
		if err != nil && err != context.Canceled {
			t.Errorf(errMsgExpectedNoErrorGot, err)
		}
	}()

	// Wait for the expected records to be processed
	wg.Wait()
	// Cancel the inner context to stop the record channel reader
	innerCancel()
}

// TestSystemAccessPointConnectWebSocketFailure tests the behavior when the WebSocket connection fails.
func TestSystemAccessPointConnectWebSocketFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sysAp, buf, _ := setupSysAp(t, false, false)

	// Set an invalid host name to simulate connection failure
	sysAp.config.Hostname = invalidHost

	// set up the error handler
	sysAp.onError = func(err error) {
		if strings.Contains(err.Error(), lookupInvalidHostSubstr) {
			cancel()
		} else {
			t.Errorf(errMsgUnexpectedErrorGot, err)
		}
	}

	// Run ConnectWebSocket in a separate goroutine
	go func() {
		err := sysAp.ConnectWebSocket(ctx, 1, false, 1*time.Hour)
		if err != nil && err != context.Canceled {
			t.Errorf(errMsgExpectedNoErrorGot, err)
		}
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgFailedConnectWebSocket) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgFailedConnectWebSocket, logOutput)
	}
}

// TestSystemAccessPointConnectWebSocketFailureWithBackoff tests the behavior when the WebSocket connection fails.
func TestSystemAccessPointConnectWebSocketFailureWithBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	sysAp, buf, _ := setupSysAp(t, false, false)
	sysAp.clock = &fakeClock{}

	// Set an invalid host name to simulate connection failure
	sysAp.config.Hostname = invalidHost

	// set up the error handler
	sysAp.onError = func(err error) {
		if strings.Contains(err.Error(), lookupInvalidHostSubstr) {
			cancel()
		} else {
			t.Errorf(errMsgUnexpectedErrorGot, err)
		}
	}

	// Run ConnectWebSocket in a separate goroutine
	go func() {
		err := sysAp.ConnectWebSocket(ctx, 3, true, 1*time.Hour)
		if err != nil && err != context.Canceled {
			t.Errorf(errMsgExpectedNoErrorGot, err)
		}
	}()

	// Wait for the context to be cancelled
	<-ctx.Done()

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgFailedConnectWebSocket) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgFailedConnectWebSocket, logOutput)
	}
}

func TestSystemAccessPointConnectWebSocketMaxReconnectionAttempts(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, false, false)

	// Set an invalid host name to simulate connection failure
	sysAp.config.Hostname = invalidHost

	// set up the error handler
	errorCount := 0
	sysAp.onError = func(err error) {
		errorCount++
	}

	// Run ConnectWebSocket with max 2 reconnection attempts
	err := sysAp.ConnectWebSocket(t.Context(), 2, false, 1*time.Hour)

	// Verify error
	if err == nil || err.Error() != logMsgMaxReconnectionExceeded {
		t.Errorf("Expected error %q, got: %v", logMsgMaxReconnectionExceeded, err)
	}

	// Verify the error count (should be 2 failed attempts)
	if errorCount != 2 {
		t.Errorf("Expected error count to be 2, got %d", errorCount)
	}

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgMaxReconnectionExceeded) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgMaxReconnectionExceeded, logOutput)
	}
}

// TestSystemAccessPointWebSocketMessageLoopTextMessage tests the webSocketMessageLoop method for text messages.
func TestSystemAccessPointWebSocketMessageLoopTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ws, buf, _ := setupSysApWebSocket(t, true, false)
	webSocketMessageChannel := make(chan []byte, 10)
	messageReceivedChannel := make(chan struct{}, 1)

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte(testMessageValid),
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	go func() {
		err := ws.webSocketMessageLoop(ctx, messageReceivedChannel, webSocketMessageChannel, conn)
		if err == nil {
			t.Error(expectedErrorGotNil)
		}
	}()

	// Wait for the context to be done
	message := <-webSocketMessageChannel
	cancel()
	<-ctx.Done()
	close(webSocketMessageChannel)
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

	// Check if the message is valid
	if string(message) != testMessageValid {
		t.Errorf("Expected message 'valid message', got: %s", string(message))
	}

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgReceivedTextMessage) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgReceivedTextMessage, logOutput)
	}
}

// TestSystemAccessPointWebSocketMessageLoopNonTextMessage tests the webSocketMessageLoop method for non-text messages.
func TestSystemAccessPointWebSocketMessageLoopNonTextMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ws, buf, _ := setupSysApWebSocket(t, true, false)
	webSocketMessageChannel := make(chan []byte, 10)
	messageReceivedChannel := make(chan struct{}, 1)
	ws.sysAp.onError = func(err error) {
		if strings.Contains(err.Error(), errMsgNoMoreMessages) {
			cancel()
		} else {
			t.Errorf(errMsgUnexpectedErrorGot, err)
		}
	}

	// Mock a non-text message
	nonTextMessage := []byte{0x00, 0x01, 0x02}

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.BinaryMessage,
		r:           nonTextMessage,
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	go func() {
		err := ws.webSocketMessageLoop(ctx, messageReceivedChannel, webSocketMessageChannel, conn)
		if err == nil {
			t.Error(expectedErrorGotNil)
		}
	}()

	// Wait for the context to be done
	<-ctx.Done()
	close(webSocketMessageChannel)
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgReceivedNonText) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgReceivedNonText, logOutput)
	}
}

// TestSystemAccessPointWebSocketMessageLoopMissingChannel tests the webSocketMessageLoop method for missing channels.
func TestSystemAccessPointWebSocketMessageLoopMissingChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ws, buf, _ := setupSysApWebSocket(t, true, false)
	messageReceivedChannel := make(chan struct{}, 1)

	// Mock a WebSocket connection
	conn := &MockConn{
		messageType: websocket.TextMessage,
		r:           []byte(testMessageValid),
		err:         nil,
	}

	// Run the message loop in a separate goroutine
	err := ws.webSocketMessageLoop(ctx, messageReceivedChannel, nil, conn)
	if err == nil {
		t.Error(expectedErrorGotNil)
	}
	// Check if the error is due to the missing channel
	if !strings.Contains(err.Error(), errMsgConnectionChannelNil) {
		t.Errorf("Expected error %q, got: %v", errMsgConnectionChannelNil, err)
	}

	// Wait for the context to be done
	cancel()
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, errMsgConnectionChannelNil) {
		t.Errorf(errMsgExpectedLogOutputToContain, errMsgConnectionChannelNil, logOutput)
	}
}

// TestSystemAccessPointWebSocketKeepaliveLoopMissingChannel tests the webSocketKeepaliveLoop method for missing channels.
func TestSystemAccessPointWebSocketKeepaliveLoopMissingChannel(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)

	// Mock a WebSocket connection
	conn := &MockConn{
		err: nil,
	}

	// Run the keepalive loop in a separate goroutine
	ws.webSocketKeepaliveLoop(nil, conn, 30*time.Second)

	// Wait for the context to be done
	ws.waitGroup.Wait()

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgMessageReceivedChannelNil) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgMessageReceivedChannelNil, logOutput)
	}
}

// TestSystemAccessPointWebSocketKeepaliveLoopSendPing tests the webSocketKeepaliveLoop method for sending a ping message.
func TestSystemAccessPointWebSocketKeepaliveLoopSendPing(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)
	messageReceivedChannel := make(chan struct{}, 1)

	// Mock a WebSocket connection with thread-safe access
	conn := &MockConn{
		err: errors.New("test error"),
		writeMessages: []struct {
			messageType int
			data        []byte
			deadline    time.Time
		}{},
		mu: &sync.Mutex{},
	}

	// Run the keepalive loop in a separate goroutine
	go func() {
		ws.webSocketKeepaliveLoop(messageReceivedChannel, conn, 250*time.Millisecond)
	}()

	time.Sleep(150 * time.Millisecond)
	conn.mu.Lock()
	writeCount := len(conn.writeMessages)
	conn.mu.Unlock()
	if writeCount != 0 {
		t.Errorf("Expected write message count to be 0, got: %d", writeCount)
	}
	time.Sleep(150 * time.Millisecond)

	// Wait for the context to be done
	close(messageReceivedChannel)
	ws.waitGroup.Wait()

	// Check the log output safely
	logOutput := buf.String()
	if !strings.Contains(logOutput, logMsgKeepalivePing) {
		t.Errorf(errMsgExpectedLogOutputToContain, logMsgKeepalivePing, logOutput)
	}

	// Check if the ping message was sent safely
	conn.mu.Lock()
	writeCount = len(conn.writeMessages)
	conn.mu.Unlock()
	if writeCount != 1 {
		t.Errorf("Expected write message count to be 1, got: %d", writeCount)
	}
}

// TestSystemAccessPointGetWebSocketUrlWithoutTls tests the getWebSocketUrl method of SystemAccessPointWebSocket without TLS.
func TestSystemAccessPointGetWsUrlWithoutTls(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, false, false)

	actual := ws.getWebSocketUrl()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != wsURLHTTP {
		t.Errorf("Expected URL '%s', got '%s'", wsURLHTTP, actual)
	}
}

// TestSystemAccessPointGetWebSocketUrlWithTls tests the getWebSocketUrl method of SystemAccessPointWebSocket with TLS.
func TestSystemAccessPointGetWsUrlWithTls(t *testing.T) {
	ws, buf, _ := setupSysApWebSocket(t, true, false)

	actual := ws.getWebSocketUrl()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != wsURLHTTPS {
		t.Errorf("Expected URL '%s', got '%s'", wsURLHTTPS, actual)
	}
}

// TestSystemAccessPointCalculateBackoffDuration tests the calculateBackoffDuration method of SystemAccessPoint.
func TestSystemAccessPointCalculateBackoffDuration(t *testing.T) {
	// Test exponential backoff calculation
	testCases := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 2 * time.Second},   // 1s * 2^1 = 2s
		{2, 4 * time.Second},   // 1s * 2^2 = 4s
		{3, 8 * time.Second},   // 1s * 2^3 = 8s
		{4, 16 * time.Second},  // 1s * 2^4 = 16s
		{5, 30 * time.Second},  // Capped at 30s
		{6, 30 * time.Second},  // Capped at 30s
		{10, 30 * time.Second}, // Capped at 30s
	}

	for _, tc := range testCases {
		result := calculateBackoffDuration(tc.attempt)
		if result != tc.expected {
			t.Errorf("For attempt %d, expected %v, got %v", tc.attempt, tc.expected, result)
		}
	}
}

type MockConn struct {
	messageRead   bool
	messageType   int
	r             []byte
	err           error
	writeMessages []struct {
		messageType int
		data        []byte
		deadline    time.Time
	}
	mu *sync.Mutex
}

func (m *MockConn) ReadMessage() (int, []byte, error) {
	if m.messageRead {
		return -1, nil, fmt.Errorf("no more messages")
	}

	m.messageRead = true
	return m.messageType, m.r, m.err
}

func (m *MockConn) WriteControl(messageType int, data []byte, deadline time.Time) error {
	if m.mu != nil {
		m.mu.Lock()
		defer m.mu.Unlock()
	}
	m.writeMessages = append(m.writeMessages, struct {
		messageType int
		data        []byte
		deadline    time.Time
	}{
		messageType: messageType,
		data:        data,
		deadline:    deadline,
	})
	return m.err
}
