package stdio

import (
	"sync"
	
	"github.com/LaurieRhodes/mcp-cli-go/internal/infrastructure/logging"
	"github.com/LaurieRhodes/mcp-cli-go/internal/providers/mcp/messages"
)

// ResponseDispatcher handles routing responses to waiting requests
type ResponseDispatcher struct {
	client       *StdioClient
	pending      map[string]chan *messages.JSONRPCMessage
	pendingMutex sync.RWMutex
	started      bool
	startMutex   sync.Mutex
}

// NewResponseDispatcher creates a new response dispatcher
func NewResponseDispatcher(client *StdioClient) *ResponseDispatcher {
	return &ResponseDispatcher{
		client:  client,
		pending: make(map[string]chan *messages.JSONRPCMessage),
	}
}

// Start begins the dispatcher goroutine (call once)
func (d *ResponseDispatcher) Start() {
	d.startMutex.Lock()
	defer d.startMutex.Unlock()
	
	if d.started {
		return // Already started
	}
	d.started = true
	
	go d.dispatch()
}

// dispatch is the main loop that routes responses
func (d *ResponseDispatcher) dispatch() {
	logging.Debug("Response dispatcher started")
	for msg := range d.client.Read() {
		msgID := msg.ID.String()
		logging.Debug("Dispatcher received message ID: %s", msgID)
		
		d.pendingMutex.RLock()
		ch, exists := d.pending[msgID]
		d.pendingMutex.RUnlock()
		
		if exists {
			logging.Debug("Routing response to waiting request: %s", msgID)
			select {
			case ch <- msg:
				// Success
			default:
				logging.Warn("Failed to send response to channel (full or closed): %s", msgID)
			}
			
			// Clean up
			d.pendingMutex.Lock()
			delete(d.pending, msgID)
			d.pendingMutex.Unlock()
		} else {
			logging.Debug("No pending request for message ID: %s", msgID)
		}
	}
	logging.Debug("Response dispatcher stopped")
}

// RegisterRequest registers a request ID and returns a channel for the response
func (d *ResponseDispatcher) RegisterRequest(requestID string) chan *messages.JSONRPCMessage {
	responseCh := make(chan *messages.JSONRPCMessage, 1)
	
	d.pendingMutex.Lock()
	d.pending[requestID] = responseCh
	d.pendingMutex.Unlock()
	
	logging.Debug("Registered request ID: %s", requestID)
	return responseCh
}

// UnregisterRequest removes a pending request (e.g., on timeout)
func (d *ResponseDispatcher) UnregisterRequest(requestID string) {
	d.pendingMutex.Lock()
	delete(d.pending, requestID)
	d.pendingMutex.Unlock()
	
	logging.Debug("Unregistered request ID: %s", requestID)
}

// GetPendingCount returns the number of pending requests (for debugging)
func (d *ResponseDispatcher) GetPendingCount() int {
	d.pendingMutex.RLock()
	defer d.pendingMutex.RUnlock()
	return len(d.pending)
}
