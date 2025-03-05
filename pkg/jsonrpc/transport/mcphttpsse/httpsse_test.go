package mcphttpsse

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flexigpt/flexiui/pkg/jsonrpc/transport/helpers_test"
)

type HTTPJSONRPCClient struct {
	client *http.Client
	url    string
}

func NewHTTPClient(t *testing.T) *HTTPJSONRPCClient {
	handler := SetupSSETransport()
	server := httptest.NewUnstartedServer(handler)
	server.Start()
	t.Cleanup(server.Close) // Ensure server closes after test
	client := server.Client()
	url := server.URL + JSONRPCEndpoint
	return &HTTPJSONRPCClient{
		client: client,
		url:    url,
	}
}

func (c *HTTPJSONRPCClient) Send(reqBytes []byte) ([]byte, error) {
	// Create a new HTTP request with context.
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		c.url,
		bytes.NewReader(reqBytes),
	)
	if err != nil {
		return nil, err
	}

	// Set the content type header.
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP request.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func getClient(t *testing.T) helpers_test.JSONRPCClient {
	return NewHTTPClient(t)
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
