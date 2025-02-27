//go:build integration

package httponly

import "testing"

// go test -v -tags=integration -run TestRunServer -count=1 ./pkg/mcpsdk/transport/example &
func TestRunServer(t *testing.T) {
	// Start the server
	StartHTTPServer()
}

func StartHTTPServer() {
	cli := GetHTTPServerCLI()
	cli.Run()
}
