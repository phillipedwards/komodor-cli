package cmd

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/phillipedwards/komodor-cli/internal/client"
)

func TestAPIKeyValidateSuccess(t *testing.T) {
	mock := &mockClient{
		validateAPIKey: func(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ApikeyValidateResponse, error) {
			return &client.GetApiV2ApikeyValidateResponse{
				HTTPResponse: &http.Response{StatusCode: 200},
				JSON200:      &client.ApiKeyValidationResponse{Valid: true},
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	_, err := runCmd(root, &buf, "apikey", "validate")
	if err != nil {
		t.Fatalf("apikey validate returned unexpected error: %v", err)
	}
}

func TestAPIKeyValidateFailure(t *testing.T) {
	mock := &mockClient{
		validateAPIKey: func(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ApikeyValidateResponse, error) {
			return &client.GetApiV2ApikeyValidateResponse{
				HTTPResponse: &http.Response{StatusCode: 401},
				Body:         []byte(`{"error":{"message":"Unauthorized"}}`),
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	_, err := runCmd(root, &buf, "apikey", "validate")
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to contain %q, got: %v", "401", err)
	}
}
