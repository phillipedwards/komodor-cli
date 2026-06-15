package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/phillipedwards/komodor-cli/internal/client"
)

// apiError formats an API error response for display.
func apiError(statusCode int, body []byte) error {
	var wrapper client.ErrorWrapper
	if err := json.Unmarshal(body, &wrapper); err == nil && wrapper.Error.Message != "" {
		return fmt.Errorf("API error %d: %s", statusCode, wrapper.Error.Message)
	}
	return fmt.Errorf("API error %d: %s", statusCode, http.StatusText(statusCode))
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func boolPtr(b bool) *bool { return &b }

func strSlicePtr(ss []string) *[]string {
	if len(ss) == 0 {
		return nil
	}
	return &ss
}

func joinStrings(ss []string) string {
	return strings.Join(ss, ", ")
}
