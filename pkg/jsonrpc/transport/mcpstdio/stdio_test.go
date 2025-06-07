package mcpstdio

import (
	"errors"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/ppipada/flexigpt-app/pkg/jsonrpc/transport/helpers_test"
	stdioNet "github.com/ppipada/flexigpt-app/pkg/jsonrpc/transport/mcpstdio/net"
)

type StdIOJSONRPCClient struct {
	// Replace with actual stdio client type.
	client *stdioNet.Client
}

func (c *StdIOJSONRPCClient) Send(reqBytes []byte) ([]byte, error) {
	return c.client.Send(reqBytes)
}

func NewStdIOClient(t *testing.T) *StdIOJSONRPCClient {
	handler := SetupStdIOTransport()
	clientReader, serverWriter := io.Pipe()
	serverReader, clientWriter := io.Pipe()

	server := GetServer(serverReader, serverWriter, handler)
	// Start the server in a goroutine.
	go func() {
		err := server.Serve()
		if err != nil && errors.Is(err, net.ErrClosed) {
			// Log the error if it's not due to server shutdown.
			fmt.Printf("Server error: %v\n", err)
		}
	}()
	t.Cleanup(func() {
		_ = server.Shutdown(t.Context())
	})
	client := GetClient(clientReader, clientWriter)
	return &StdIOJSONRPCClient{
		client: client,
	}
}

func getClient(t *testing.T) helpers_test.JSONRPCClient {
	return NewStdIOClient(t)
}

func TestValidSingleRequests(t *testing.T) {
	helpers_test.TestValidSingleRequests(t, getClient(t))
}

func TestInvalidSingleRequests(t *testing.T) {
	helpers_test.TestInvalidSingleRequests(t, getClient(t))
}

func TestNotifications(t *testing.T) {
	helpers_test.TestNotifications(t, getClient(t))
}

func TestBatchRequests(t *testing.T) {
	helpers_test.TestBatchRequests(t, getClient(t))
}
