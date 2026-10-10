package apilogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// ToJSON safely marshals any value into a JSON string.
// If v is nil, returns empty string. If v is already a string, returns as-is.
func ToJSON(v any) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// BuildCurlFromRequest creates a copy-pasteable curl command from an *http.Request.
// The request body is read and restored so that subsequent reads in the application continue to work.
func BuildCurlFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}

	var body string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			body = string(bodyBytes)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	headers := make(map[string]string)
	for k, v := range req.Header {
		headers[k] = strings.Join(v, ", ")
	}

	endpoint := ""
	if req.URL != nil {
		endpoint = req.URL.String()
	}

	return BuildCurl(req.Method, endpoint, headers, body)
}

// BuildCurl generates a complete curl command string from method, endpoint, headers, and body.
func BuildCurl(method, endpoint string, headers map[string]string, body string) string {
	if endpoint == "" {
		return ""
	}
	if method == "" {
		method = "GET"
	}

	var parts []string
	parts = append(parts, "curl", "-X", strings.ToUpper(method))
	parts = append(parts, fmt.Sprintf("'%s'", endpoint))

	// Sort header keys for deterministic output
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		val := headers[k]
		escapedVal := strings.ReplaceAll(val, "'", `'\''`)
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", k, escapedVal))
	}

	if body != "" {
		escapedBody := strings.ReplaceAll(body, "'", `'\''`)
		parts = append(parts, fmt.Sprintf("-d '%s'", escapedBody))
	}

	return strings.Join(parts, " ")
}
