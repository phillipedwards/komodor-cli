package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

// captureFormatter is a test-only Formatter that writes table output to a bytes.Buffer.
// It satisfies output.Formatter and output.TableData rendering without writing to os.Stdout.
type captureFormatter struct {
	buf *bytes.Buffer
}

func (f *captureFormatter) Print(data any) error {
	td, ok := data.(output.TableData)
	if !ok {
		return fmt.Errorf("captureFormatter: unsupported type %T", data)
	}
	// Write headers
	headers := td.Headers()
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(f.buf, "\t")
		}
		fmt.Fprint(f.buf, h)
	}
	fmt.Fprintln(f.buf)
	// Write rows
	for _, row := range td.Rows() {
		for i, v := range row {
			if i > 0 {
				fmt.Fprint(f.buf, "\t")
			}
			fmt.Fprint(f.buf, v)
		}
		fmt.Fprintln(f.buf)
	}
	return nil
}

// mockClient embeds the interface so unimplemented methods panic at test time.
type mockClient struct {
	client.ClientWithResponsesInterface
	getClusters        func(ctx context.Context, params *client.GetApiV2ClustersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ClustersResponse, error)
	postServicesSearch func(ctx context.Context, body client.SearchServicesBody, reqEditors ...client.RequestEditorFn) (*client.PostApiV2ServicesSearchResponse, error)
	getUsers           func(ctx context.Context, params *client.GetApiV2UsersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2UsersResponse, error)
	validateAPIKey     func(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ApikeyValidateResponse, error)
	getUserClusters    func(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.GetApiV2UserClustersResponse, error)
}

func (m *mockClient) GetApiV2ClustersWithResponse(ctx context.Context, params *client.GetApiV2ClustersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ClustersResponse, error) {
	if m.getClusters != nil {
		return m.getClusters(ctx, params, reqEditors...)
	}
	panic("mockClient.GetApiV2ClustersWithResponse not implemented")
}

func (m *mockClient) PostApiV2ServicesSearchWithResponse(ctx context.Context, body client.SearchServicesBody, reqEditors ...client.RequestEditorFn) (*client.PostApiV2ServicesSearchResponse, error) {
	if m.postServicesSearch != nil {
		return m.postServicesSearch(ctx, body, reqEditors...)
	}
	panic("mockClient.PostApiV2ServicesSearchWithResponse not implemented")
}

func (m *mockClient) GetApiV2UsersWithResponse(ctx context.Context, params *client.GetApiV2UsersParams, reqEditors ...client.RequestEditorFn) (*client.GetApiV2UsersResponse, error) {
	if m.getUsers != nil {
		return m.getUsers(ctx, params, reqEditors...)
	}
	panic("mockClient.GetApiV2UsersWithResponse not implemented")
}

func (m *mockClient) GetApiV2ApikeyValidateWithResponse(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.GetApiV2ApikeyValidateResponse, error) {
	if m.validateAPIKey != nil {
		return m.validateAPIKey(ctx, reqEditors...)
	}
	panic("mockClient.GetApiV2ApikeyValidateWithResponse not implemented")
}

func (m *mockClient) GetApiV2UserClustersWithResponse(ctx context.Context, reqEditors ...client.RequestEditorFn) (*client.GetApiV2UserClustersResponse, error) {
	if m.getUserClusters != nil {
		return m.getUserClusters(ctx, reqEditors...)
	}
	panic("mockClient.GetApiV2UserClustersWithResponse not implemented")
}

// newTestRoot returns the root cobra.Command with a mock client and capture formatter injected.
// It replaces PersistentPreRunE entirely so that auth and config are bypassed.
func newTestRoot(t *testing.T, mock *mockClient, buf *bytes.Buffer) *cobra.Command {
	t.Helper()
	root := NewRootCmd()
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		f := &captureFormatter{buf: buf}
		cmd.SetContext(context.WithValue(
			context.WithValue(cmd.Context(), keyClient, client.ClientWithResponsesInterface(mock)),
			keyFormatter, output.Formatter(f),
		))
		return nil
	}
	return root
}

// runCmd executes a cobra command with the given args, capturing stdout and stderr into a single buffer.
// It returns the combined output string and any execution error.
func runCmd(root *cobra.Command, buf *bytes.Buffer, args ...string) (string, error) {
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}
