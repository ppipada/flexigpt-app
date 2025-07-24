package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	goruntime "runtime"
	"strings"
)

const FrontendPathPrefix = "/frontend/dist"

var DIRPages = []string{"/agents", "/chats", "/settings", "/404"}

func EmbeddedFSWalker(assets embed.FS) {
	_ = fs.WalkDir(assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		slog.Info("Embedded walk", "Path", path)
		return nil
	})
}

func LogStackTrace() {
	// Create a buffer to hold the stack trace.
	buf := make([]byte, 1024)
	// Capture the stack trace.
	n := goruntime.Stack(buf, false)
	// Log the stack trace.
	slog.Info("Stack", "Trace", string(buf[:n]))
}

func URLCleanerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Clean the URL using getActualURL.
		cleanedURL := getActualURL(req.URL.Path)

		// Update the request URL path.
		req.URL.Path = cleanedURL

		// Call the next handler.
		next.ServeHTTP(w, req)
	})
}

func getActualURL(origurl string) string {
	callurl := origurl
	if !strings.HasPrefix(callurl, FrontendPathPrefix) {
		return callurl
	}

	// Handle if it's a page request.
	if strings.HasSuffix(callurl, "/") {
		return callurl[len(FrontendPathPrefix):] + "index.html"
	}

	for _, d := range DIRPages {
		durl := FrontendPathPrefix + d
		if callurl == durl {
			return callurl[len(FrontendPathPrefix):] + "/index.html"
		}
	}

	return callurl[len(FrontendPathPrefix):]
}
