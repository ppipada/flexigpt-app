package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// Sensitive keys to filter.
var sensitiveKeys = []string{"authorization", "key"}

// Define a context key type to avoid collisions.
type contextKey string

const debugHTTPResponseKey = contextKey("DebugHTTPResponse")

// DebugHTTPResponse wraps http.Response and includes additional debug information.
type DebugHTTPResponse struct {
	RequestDetails  *APIRequestDetails
	ResponseDetails *APIResponseDetails
	ErrorDetails    *APIErrorDetails
}

// FilterSensitiveInfo recursively filters out sensitive keys from a data structure.
// It supports nested maps and slices.
func FilterSensitiveInfo(data map[string]any) map[string]any {
	filteredData := make(map[string]any)
	for key, value := range data {
		if containsSensitiveKey(key) {
			// Mask the sensitive value.
			filteredData[key] = "***"
		} else {
			// Recursively process nested data structures.
			filteredData[key] = deepCopyAndFilter(value)
		}
	}
	return filteredData
}

// deepCopyAndFilter recursively traverses the data structure,
// filtering sensitive keys from maps and processing slices.
func deepCopyAndFilter(value any) any {
	switch v := value.(type) {
	case map[string]any:
		// Process nested map.
		return FilterSensitiveInfo(v)
	case []any:
		// Process each element in the slice.
		newSlice := make([]any, len(v))
		for i, elem := range v {
			newSlice[i] = deepCopyAndFilter(elem)
		}
		return newSlice
	default:
		// Return the value as is for other data types.
		return v
	}
}

// containsSensitiveKey checks if a key contains any sensitive keywords.
func containsSensitiveKey(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, sensitiveKey := range sensitiveKeys {
		if strings.Contains(lowerKey, sensitiveKey) {
			return true
		}
	}
	return false
}

func generateCurlCommand(config *APIRequestDetails) string {
	var curlCommand strings.Builder

	// Add HTTP method.
	if config.Method != nil {
		curlCommand.WriteString("curl -X " + strings.ToUpper(*config.Method) + " ")
	}

	// Add URL.
	if config.URL != nil {
		curlCommand.WriteString("\"" + *config.URL + "\" ")
	}

	// Add headers.
	if config.Headers != nil {
		for key, value := range config.Headers {
			curlCommand.WriteString(fmt.Sprintf("-H \"%s: %v\" ", key, value))
		}
	}

	// Add data.
	if config.Data != nil {
		dataBytes, err := json.Marshal(config.Data)
		if err == nil {
			curlCommand.WriteString(fmt.Sprintf("-d '%s' ", string(dataBytes)))
		}
	}

	return curlCommand.String()
}

// captureRequestDetails captures details of the HTTP request.
func captureRequestDetails(req *http.Request) *APIRequestDetails {
	headers := make(map[string]any)
	for key, values := range req.Header {
		headers[key] = strings.Join(values, ", ")
	}

	var data map[string]any
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		// Reset body for further use.
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		_ = json.Unmarshal(bodyBytes, &data)
	}

	url := req.URL.String()
	method := req.Method

	apireq := &APIRequestDetails{
		URL:     &url,
		Method:  &method,
		Headers: FilterSensitiveInfo(headers),
		Data:    FilterSensitiveInfo(data),
	}

	curlcmd := generateCurlCommand(apireq)
	apireq.CurlCommand = &curlcmd

	return apireq
}

// captureResponseDetails captures details of the HTTP response.
func CaptureResponseDetails(
	resp *http.Response,
) *APIResponseDetails {
	headers := make(map[string]any)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	var data map[string]any
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			// Attempt to unmarshal the body into 'data'.
			_ = json.Unmarshal(bodyBytes, &data)
		}
		// Reconstruct the body so it can be read again later.
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	return &APIResponseDetails{
		Status:  resp.StatusCode,
		Headers: FilterSensitiveInfo(headers),
		Data:    FilterSensitiveInfo(data),
	}
}

type loggingReadCloser struct {
	io.ReadCloser
	buf       *bytes.Buffer
	debugResp *DebugHTTPResponse
	logMode   bool
}

func (lc *loggingReadCloser) Read(p []byte) (int, error) {
	n, err := lc.ReadCloser.Read(p)
	if n > 0 {
		lc.buf.Write(p[:n])
	}
	return n, err
}

func (lc *loggingReadCloser) Close() error {
	err := lc.ReadCloser.Close()
	if err != nil {
		return err
	}
	dataBytes := lc.buf.Bytes()

	// Process the data based on its type.
	var data any
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		// Text data.
		lc.debugResp.ResponseDetails.Data = string(dataBytes)
	} else {
		mapData, ok := data.(map[string]any)
		if ok {
			// JSON data.
			lc.debugResp.ResponseDetails.Data = FilterSensitiveInfo(mapData)
		} else {
			// Text data.
			lc.debugResp.ResponseDetails.Data = string(dataBytes)
		}
	}

	if lc.logMode {
		slog.Debug("Response", "Body", string(dataBytes)+"\n")
	}

	return err
}

// LogTransport is a custom http.RoundTripper that logs requests and responses.
type LogTransport struct {
	Transport http.RoundTripper
	LogMode   bool
}

func getDetailsStr(v any) string {
	// Two spaces for indentation.
	s, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Could not get json of object: %+v", v)
	}
	return string(s)
}

// RoundTrip executes a single HTTP transaction and logs the request and response.
func (t *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCtx := req.Context()
	debugResp, ok := GetDebugHTTPResponse(reqCtx)
	if !ok || debugResp == nil {
		// Allocate a pointer for processing.
		// This is not going to be available in consumer anycase, but is present for processing sakes only.
		debugResp = &DebugHTTPResponse{}
	}

	// Capture request details.
	reqDetails := captureRequestDetails(req)
	debugResp.RequestDetails = reqDetails

	// Log request details if LogMode is enabled.
	if t.LogMode {
		slog.Debug("Roundtripper", "Request Details", getDetailsStr(reqDetails))
	}

	// Perform the request.
	resp, err := t.Transport.RoundTrip(req)

	// Capture response details.
	var respDetails *APIResponseDetails
	if resp != nil {
		// Capture headers.
		headers := make(map[string]any)
		for key, values := range resp.Header {
			headers[key] = strings.Join(values, ", ")
		}

		// Initialize response details.
		respDetails = &APIResponseDetails{
			Status:  resp.StatusCode,
			Headers: FilterSensitiveInfo(headers),
		}
		debugResp.ResponseDetails = respDetails

		// Wrap the response body.
		buffer := new(bytes.Buffer)
		resp.Body = &loggingReadCloser{
			ReadCloser: resp.Body,
			buf:        buffer,
			debugResp:  debugResp,
			logMode:    t.LogMode,
		}
	}

	// Capture error details if an error occurred.
	var errorDetails *APIErrorDetails
	if err != nil {
		errorDetails = &APIErrorDetails{
			Message:         err.Error(),
			RequestDetails:  reqDetails,
			ResponseDetails: respDetails,
		}
		debugResp.ErrorDetails = errorDetails
	}

	// Log response details if LogMode is enabled.
	if t.LogMode {
		if respDetails != nil {
			slog.Debug("Roundtripper", "Response Details", getDetailsStr(respDetails))
		}
		if errorDetails != nil {
			slog.Debug("Roundtripper", "Error Details", getDetailsStr(errorDetails))
		}
	}

	// Return the response and error.
	return resp, err
}

// NewDebugHTTPClient creates a new HTTP client with logging capabilities.
func NewDebugHTTPClient(logMode bool) *http.Client {
	return &http.Client{
		Transport: &LogTransport{
			Transport: http.DefaultTransport,
			LogMode:   logMode,
		},
	}
}

func AddDebugResponseToCtx(ctx context.Context) context.Context {
	debugResp := &DebugHTTPResponse{}
	// Create a context with the DebugHTTPResponse.
	return context.WithValue(ctx, debugHTTPResponseKey, debugResp)
}

// Helper function to retrieve DebugHTTPResponse from context.
func GetDebugHTTPResponse(ctx context.Context) (*DebugHTTPResponse, bool) {
	debugResp, ok := ctx.Value(debugHTTPResponseKey).(*DebugHTTPResponse)
	return debugResp, ok
}

func PrintJSON(v any) {
	p, err := json.MarshalIndent(v, "", "")
	if err == nil {
		slog.Info("Request params", "JSON", string(p))
	}
}
