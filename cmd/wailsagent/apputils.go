package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	goruntime "runtime"
	"strings"
)

const FRONTEND_PATH_PREFIX = "/frontend/build"

var DIR_PAGES = []string{"/agents", "/chats", "/settings", "/404"}

func getActualURL(origurl string) string {
	callurl := origurl
	if !strings.HasPrefix(callurl, FRONTEND_PATH_PREFIX) {
		return callurl
	}

	// percentIndex := strings.Index(callurl, "%")
	// if percentIndex != -1 {
	// 	callurl = callurl[:percentIndex]
	// }

	// qIndex := strings.Index(callurl, "?")
	// if qIndex != -1 {
	// 	callurl = callurl[:qIndex]
	// }

	// Handle if it's a page request
	if strings.HasSuffix(callurl, "/") {
		return callurl[len(FRONTEND_PATH_PREFIX):] + "index.html"
	}

	for _, d := range DIR_PAGES {
		durl := FRONTEND_PATH_PREFIX + d
		if callurl == durl {
			return callurl[len(FRONTEND_PATH_PREFIX):] + "/index.html"
		}
	}

	return callurl[len(FRONTEND_PATH_PREFIX):]
}

func LogStackTrace() {
	// Create a buffer to hold the stack trace
	buf := make([]byte, 1024)
	// Capture the stack trace
	n := goruntime.Stack(buf, false)
	// Log the stack trace
	slog.Info("Stack", "Trace", string(buf[:n]))
}

func URLCleanerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Clean the URL using getActualURL
		// if req.URL.Path == "/agents" {
		// 	LogStackTrace()
		// }
		cleanedURL := getActualURL(req.URL.Path)

		// Update the request URL path
		req.URL.Path = cleanedURL

		// Call the next handler
		next.ServeHTTP(w, req)
	})
}

func EmbeddedFSWalker(assets embed.FS) {
	_ = fs.WalkDir(assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		slog.Info("Embedded walk", "Path", path)
		return nil
	})
}
