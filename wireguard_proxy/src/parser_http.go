package main

import (
        "bufio"
        "bytes"
        "io"
        "net/http"
)

// ParsedHTTP holds the structured metadata needed for your policy engine and audit logger
type ParsedHTTP struct {
        *ConnectionContext
        Method  string            `json:"method"`
        URL     string            `json:"url"`
        Headers map[string]string `json:"headers"`
        Body    string            `json:"body_snippet,omitempty"`
}

// ParseAndInspectHTTP reads an incoming HTTP request from a buffered connection,
// parses it, extracts the required fields, and returns BOTH the metadata structure
// and the original *http.Request (with body restored) so you can modify and forward it.
func ParseAndInspectHTTP(br *bufio.Reader) (*ParsedHTTP, *http.Request, error) {
        // 1. Parse the raw byte stream into a standard Go http.Request object
        req, err := http.ReadRequest(br)
        if err != nil {
                return nil, nil, err
        }

        // 2. Flatten headers into a simple key-value map for logging/policy inspection
        headers := make(map[string]string)
        for key, values := range req.Header {
                if len(values) > 0 {
                        headers[key] = values[0]
                }
        }

        // 3. Read a snippet of the body (e.g., first 1KB) for policy auditing
        var bodySnippet string
        if req.Body != nil {
                bodyBytes, err := io.ReadAll(io.LimitReader(req.Body, 1024))
                if err == nil && len(bodyBytes) > 0 {
                        bodySnippet = string(bodyBytes)

                        // CRITICAL: Because reading req.Body consumes it, we must replace it
                        // with a fresh reader so the proxy can still forward it to the backend server!
			remainingBody := io.MultiReader(bytes.NewReader(bodyBytes), req.Body)
                        //req.Body = io.NopCloser(bytes.NewReader(remainingBody))
			req.Body = io.NopCloser(remainingBody)
                }
        }

        // Construct the final structured metadata
        parsed := &ParsedHTTP{
                Method:  req.Method,
                URL:     req.URL.String(),
                Headers: headers,
                Body:    bodySnippet,
        }

        // Return both the audit struct and the raw request object for forwarding/modification
        return parsed, req, nil
}
