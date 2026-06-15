package cmd

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/phillipedwards/komodor-cli/internal/client"
)

func TestServicesSearchReturnsTable(t *testing.T) {
	mock := &mockClient{
		postServicesSearch: func(ctx context.Context, body client.SearchServicesBody, reqEditors ...client.RequestEditorFn) (*client.PostApiV2ServicesSearchResponse, error) {
			return &client.PostApiV2ServicesSearchResponse{
				HTTPResponse: &http.Response{StatusCode: 200},
				JSON200: &client.SearchServicesResponse{
					Data: client.SearchServicesData{
						Services: []client.SingleService{
							{Service: "my-svc", Cluster: "prod", Namespace: "default", Kind: "Deployment", Link: "https://app.komodor.com/svc/my-svc"},
							{Service: "another-svc", Cluster: "staging", Namespace: "kube-system", Kind: "StatefulSet", Link: "https://app.komodor.com/svc/another-svc"},
						},
					},
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	out, err := runCmd(root, &buf, "services", "search")
	if err != nil {
		t.Fatalf("services search returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-svc") {
		t.Errorf("expected output to contain %q, got:\n%s", "my-svc", out)
	}
	if !strings.Contains(out, "another-svc") {
		t.Errorf("expected output to contain %q, got:\n%s", "another-svc", out)
	}
}

func TestServicesSearchEmpty(t *testing.T) {
	mock := &mockClient{
		postServicesSearch: func(ctx context.Context, body client.SearchServicesBody, reqEditors ...client.RequestEditorFn) (*client.PostApiV2ServicesSearchResponse, error) {
			return &client.PostApiV2ServicesSearchResponse{
				HTTPResponse: &http.Response{StatusCode: 200},
				JSON200: &client.SearchServicesResponse{
					Data: client.SearchServicesData{
						Services: []client.SingleService{},
					},
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	_, err := runCmd(root, &buf, "services", "search")
	if err != nil {
		t.Fatalf("services search returned unexpected error on empty result: %v", err)
	}
}

func TestServicesSearchAPIError(t *testing.T) {
	mock := &mockClient{
		postServicesSearch: func(ctx context.Context, body client.SearchServicesBody, reqEditors ...client.RequestEditorFn) (*client.PostApiV2ServicesSearchResponse, error) {
			return &client.PostApiV2ServicesSearchResponse{
				HTTPResponse: &http.Response{StatusCode: 403},
				Body:         []byte(`{"error":{"message":"Forbidden"}}`),
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	_, err := runCmd(root, &buf, "services", "search")
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error to contain %q, got: %v", "403", err)
	}
}
