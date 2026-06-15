//go:build integration

package cmd

import (
	"bytes"
	"os"
	"testing"
)

func skipIfNoKey(t *testing.T) {
	t.Helper()
	if os.Getenv("KOMODOR_API_KEY") == "" {
		t.Skip("KOMODOR_API_KEY not set")
	}
}

func TestIntegrationClustersListLive(t *testing.T) {
	skipIfNoKey(t)
	var buf bytes.Buffer
	root := NewRootCmd()
	out, err := runCmd(root, &buf, "clusters", "list")
	if err != nil {
		t.Fatalf("clusters list failed: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output")
	}
}

func TestIntegrationAPIKeyValidateLive(t *testing.T) {
	skipIfNoKey(t)
	var buf bytes.Buffer
	root := NewRootCmd()
	_, err := runCmd(root, &buf, "apikey", "validate")
	if err != nil {
		t.Fatalf("apikey validate failed: %v", err)
	}
}
