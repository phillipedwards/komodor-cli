package cmd

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/phillipedwards/komodor-cli/internal/client"
)

func TestClustersListTable(t *testing.T) {
	mock := &mockClient{
		getClusters: func(ctx context.Context, params *client.GetApiV2ClustersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ClustersResponse, error) {
			return &client.GetApiV2ClustersResponse{
				HTTPResponse: &http.Response{StatusCode: 200},
				JSON200: &client.ClustersResponse{
					Data: client.ClustersData{
						Clusters: []client.SingleCluster{
							{Name: "prod-cluster", ApiServerUrl: "https://k8s.example.com", Tags: map[string]interface{}{}},
							{Name: "staging-cluster", ApiServerUrl: "https://staging.k8s.example.com", Tags: map[string]interface{}{}},
						},
					},
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	out, err := runCmd(root, &buf, "clusters", "list")
	if err != nil {
		t.Fatalf("clusters list returned unexpected error: %v", err)
	}
	if !strings.Contains(out, "prod-cluster") {
		t.Errorf("expected output to contain %q, got:\n%s", "prod-cluster", out)
	}
	if !strings.Contains(out, "staging-cluster") {
		t.Errorf("expected output to contain %q, got:\n%s", "staging-cluster", out)
	}
}

func TestClustersListEmpty(t *testing.T) {
	mock := &mockClient{
		getClusters: func(ctx context.Context, params *client.GetApiV2ClustersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ClustersResponse, error) {
			return &client.GetApiV2ClustersResponse{
				HTTPResponse: &http.Response{StatusCode: 200},
				JSON200: &client.ClustersResponse{
					Data: client.ClustersData{
						Clusters: []client.SingleCluster{},
					},
				},
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	out, err := runCmd(root, &buf, "clusters", "list")
	if err != nil {
		t.Fatalf("clusters list returned unexpected error on empty result: %v", err)
	}
	// Headers should still be present
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected output to contain column header %q, got:\n%s", "NAME", out)
	}
}

func TestClustersListAPIError(t *testing.T) {
	mock := &mockClient{
		getClusters: func(ctx context.Context, params *client.GetApiV2ClustersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ClustersResponse, error) {
			return &client.GetApiV2ClustersResponse{
				HTTPResponse: &http.Response{StatusCode: 403},
				Body:         []byte(`{"error":{"message":"Forbidden"}}`),
			}, nil
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	_, err := runCmd(root, &buf, "clusters", "list")
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error to contain %q, got: %v", "403", err)
	}
}

func TestClustersListNetworkError(t *testing.T) {
	mock := &mockClient{
		getClusters: func(ctx context.Context, params *client.GetApiV2ClustersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ClustersResponse, error) {
			return nil, errors.New("network error")
		},
	}

	var buf bytes.Buffer
	root := newTestRoot(t, mock, &buf)
	_, err := runCmd(root, &buf, "clusters", "list")
	if err == nil {
		t.Fatal("expected error for network failure, got nil")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("expected error to wrap %q, got: %v", "network error", err)
	}
}
