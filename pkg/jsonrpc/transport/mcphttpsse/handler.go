package mcphttpsse

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/flexigpt/flexiui/pkg/jsonrpc/humaadapter"
	jsonrpcReqResp "github.com/flexigpt/flexiui/pkg/jsonrpc/reqresp"
	"github.com/google/uuid"
)

const (
	SSEEndpoint     = "/sse"
	JSONRPCEndpoint = "/jsonrpc"
)

// SSETransport manages SSE connections and messages.
type SSETransport struct {
	sessionMap map[string]sse.Sender
	mu         sync.Mutex
}

// NewSSETransport creates a new SSE server transport.
func NewSSETransport(endpoint string) *SSETransport {
	return &SSETransport{
		sessionMap: make(map[string]sse.Sender),
	}
}

// RegisterRoutes registers the SSE and POST message handlers with the Huma API.
func (s *SSETransport) Register(
	api huma.API,
	methodMap map[string]jsonrpcReqResp.IMethodHandler,
	notificationMap map[string]jsonrpcReqResp.INotificationHandler,
) {
	// Define the mapping between event names and message types.
	messageTypes := map[string]any{
		"message": jsonrpcReqResp.BatchRequest{
			Body: &jsonrpcReqResp.BatchItem[jsonrpcReqResp.UnionRequest]{},
		},
		"endpoint": "", // For the initial endpoint event.
	}

	// Register the SSE endpoint.
	sse.Register(api, huma.Operation{
		OperationID: "sse-connection",
		Method:      http.MethodGet,
		Path:        SSEEndpoint,
		Summary:     "Establishes an SSE connection",
	}, messageTypes, s.handleSSEConnection)

	// Get default operation
	op := humaadapter.GetDefaultOperation()
	op.Path = JSONRPCEndpoint
	// Register the methods
	humaadapter.Register(api, op, methodMap, notificationMap, nil, nil)
}

// handleSSEConnection handles the initial SSE connection request.
func (s *SSETransport) handleSSEConnection(
	ctx context.Context,
	input *struct{},
	send sse.Sender,
) {
	// Generate a unique session ID.
	u, err := uuid.NewV7()
	if err != nil {
		// TODO: Write error and close
		return
	}
	sessionID := u.String()

	// Store the sender associated with the session ID.
	s.mu.Lock()
	s.sessionMap[sessionID] = send
	s.mu.Unlock()

	// Log the new connection.
	log.Printf("New SSE connection established. Session ID: %s", sessionID)

	// Send the endpoint event to the client with the session ID.
	// The event is deduced from value type in the handler
	err = send.Data(fmt.Sprintf("%s?sessionId=%s", JSONRPCEndpoint, sessionID))
	if err != nil {
		// TODO: Write error and close
		return
	}

	// Wait until the context is done (client disconnects).
	<-ctx.Done()

	// Clean up when the client disconnects.
	s.mu.Lock()
	delete(s.sessionMap, sessionID)
	s.mu.Unlock()
	log.Printf("SSE connection closed. Session ID: %s", sessionID)
}
