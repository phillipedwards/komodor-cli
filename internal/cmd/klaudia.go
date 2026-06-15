package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/phillipedwards/komodor-cli/internal/client"
	"github.com/phillipedwards/komodor-cli/internal/output"
)

func newKlaudiaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "klaudia",
		Aliases: []string{"ai"},
		Short:   "Klaudia AI investigation",
	}
	cmd.AddCommand(
		newKlaudiaRcaCmd(),
		newKlaudiaFilesCmd(),
	)
	return cmd
}

func newKlaudiaRcaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rca",
		Short: "Root cause analysis operations",
	}
	cmd.AddCommand(
		newKlaudiaRcaTriggerCmd(),
		newKlaudiaRcaGetCmd(),
	)
	return cmd
}

func newKlaudiaRcaTriggerCmd() *cobra.Command {
	var cluster, name, namespace, kind, issueID string

	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger a Klaudia RCA investigation",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			body := client.KlaudiaRcaRequest{
				ClusterName: cluster,
				Kind:        kind,
				Name:        name,
				Namespace:   namespace,
			}
			if issueID != "" {
				body.IssueId = &issueID
			}

			resp, err := c.TriggerKlaudiaRcaWithResponse(cmd.Context(), body)
			if err != nil {
				return fmt.Errorf("trigger klaudia rca: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			if resp.JSON200 == nil {
				return f.Print(map[string]interface{}{})
			}
			return f.Print(&rcaTriggerTable{r: resp.JSON200})
		},
	}

	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	cmd.Flags().StringVar(&name, "name", "", "Resource name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Resource namespace")
	cmd.Flags().StringVar(&kind, "kind", "", "Resource kind (e.g. Deployment)")
	cmd.Flags().StringVar(&issueID, "issue-id", "", "Optional issue ID")
	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("namespace")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.RegisterFlagCompletionFunc("cluster", completeClusterNames)

	return cmd
}

func newKlaudiaRcaGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <session-id>",
		Short: "Get Klaudia RCA investigation results",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			resp, err := c.GetKlaudiaRcaResultsWithResponse(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("get klaudia rca results: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			if resp.JSON200 == nil {
				return f.Print(map[string]interface{}{})
			}
			return f.Print(resp.JSON200)
		},
	}
}

func newKlaudiaFilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "Manage Klaudia AI files",
	}
	cmd.AddCommand(
		newKlaudiaFilesListCmd(),
		newKlaudiaFilesGetCmd(),
		newKlaudiaFilesUploadCmd(),
		newKlaudiaFilesUpdateCmd(),
		newKlaudiaFilesDeleteCmd(),
	)
	return cmd
}

func newKlaudiaFilesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <type>",
		Short: "List Klaudia AI files (type: blueprint, knowledge-base)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			pType := client.ListKlaudiaFilesParamsType(args[0])

			resp, err := c.ListKlaudiaFilesWithResponse(cmd.Context(), pType)
			if err != nil {
				return fmt.Errorf("list klaudia files: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			if resp.JSON200 == nil {
				return f.Print(&aiFilesTable{})
			}
			list, err := resp.JSON200.AsAIFileListResponse()
			if err != nil {
				return fmt.Errorf("parse file list: %w", err)
			}
			return f.Print(&aiFilesTable{rows: list.Files})
		},
	}
}

func newKlaudiaFilesGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <type> <file-id>",
		Short: "Download a Klaudia AI file to stdout (type: blueprint, knowledge-base)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)

			pType := client.DownloadKlaudiaFileParamsType(args[0])
			fileID, err := parseUUID(args[1])
			if err != nil {
				return err
			}

			resp, err := c.DownloadKlaudiaFileWithResponse(cmd.Context(), pType, fileID)
			if err != nil {
				return fmt.Errorf("download klaudia file: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			_, err = os.Stdout.Write(resp.Body)
			return err
		},
	}
}

func newKlaudiaFilesUploadCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "upload <type>",
		Short: "Upload a Klaudia AI file (type: blueprint, knowledge-base)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			pType := client.UploadKlaudiaFilesParamsType(args[0])
			body, contentType, err := buildMultipartFile("files", filePath)
			if err != nil {
				return err
			}

			resp, err := c.UploadKlaudiaFilesWithBodyWithResponse(cmd.Context(), pType, contentType, body)
			if err != nil {
				return fmt.Errorf("upload klaudia file: %w", err)
			}
			if resp.StatusCode() != 201 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			if resp.JSON201 == nil {
				return f.Print(map[string]interface{}{})
			}
			return f.Print(&aiFilesTable{rows: resp.JSON201.Files})
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to file to upload")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newKlaudiaFilesUpdateCmd() *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "update <type> <file-id>",
		Short: "Update a Klaudia AI file (type: blueprint, knowledge-base)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			pType := client.UpdateKlaudiaFileParamsType(args[0])
			fileID, err := parseUUID(args[1])
			if err != nil {
				return err
			}

			body, contentType, err := buildMultipartFile("file", filePath)
			if err != nil {
				return err
			}

			resp, err := c.UpdateKlaudiaFileWithBodyWithResponse(cmd.Context(), pType, fileID, contentType, body)
			if err != nil {
				return fmt.Errorf("update klaudia file: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			if resp.JSON200 == nil {
				return f.Print(map[string]interface{}{})
			}
			fileData, err := resp.JSON200.AsAIFile()
			if err != nil {
				return f.Print(resp.JSON200)
			}
			return f.Print(&aiFilesTable{rows: []client.AIFile{fileData}})
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Path to replacement file")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newKlaudiaFilesDeleteCmd() *cobra.Command {
	var fileIDs []string

	cmd := &cobra.Command{
		Use:   "delete <type>",
		Short: "Delete Klaudia AI files (type: blueprint, knowledge-base)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := clientFromCtx(cmd)
			f := formatterFromCtx(cmd)

			pType := client.DeleteKlaudiaFilesParamsType(args[0])

			var body client.SchemasAIFileDeleteRequest
			if err := body.FromAIFileDeleteRequest(client.AIFileDeleteRequest{FileIDs: fileIDs}); err != nil {
				return fmt.Errorf("build delete request: %w", err)
			}

			resp, err := c.DeleteKlaudiaFilesWithResponse(cmd.Context(), pType, body)
			if err != nil {
				return fmt.Errorf("delete klaudia files: %w", err)
			}
			if resp.StatusCode() != 200 {
				return apiError(resp.StatusCode(), resp.Body)
			}
			if resp.JSON200 == nil {
				return f.Print(map[string]interface{}{})
			}
			del, err := resp.JSON200.AsAIFileDeleteResponse()
			if err != nil {
				return f.Print(resp.JSON200)
			}
			fmt.Printf("deleted: %s\n", strings.Join(del.DeletedFiles, ", "))
			if len(del.FailedFiles) > 0 {
				fmt.Printf("failed:  %s\n", strings.Join(del.FailedFiles, ", "))
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&fileIDs, "file-id", nil, "File ID to delete (repeatable)")
	_ = cmd.MarkFlagRequired("file-id")

	return cmd
}

// buildMultipartFile builds a multipart/form-data body with a single file field.
func buildMultipartFile(fieldName, filePath string) (io.Reader, string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("read file %s: %w", filePath, err)
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return nil, "", fmt.Errorf("create form field: %w", err)
	}
	if _, err = fw.Write(data); err != nil {
		return nil, "", fmt.Errorf("write form data: %w", err)
	}
	if err = mw.Close(); err != nil {
		return nil, "", err
	}

	return &buf, mw.FormDataContentType(), nil
}

type rcaTriggerTable struct {
	r *client.KlaudiaRcaResponse
}

func (t *rcaTriggerTable) Headers() []string {
	return []string{"SESSION_ID", "SESSION_URL"}
}

func (t *rcaTriggerTable) Rows() [][]string {
	if t.r == nil {
		return nil
	}
	return [][]string{{t.r.SessionId, t.r.SessionUrl}}
}

var _ output.TableData = (*rcaTriggerTable)(nil)

type aiFilesTable struct {
	rows []client.AIFile
}

func (t *aiFilesTable) Headers() []string {
	return []string{"ID", "NAME", "SIZE", "UPLOADED_AT", "CREATED_BY"}
}

func (t *aiFilesTable) Rows() [][]string {
	rows := make([][]string, len(t.rows))
	for i, f := range t.rows {
		rows[i] = []string{
			f.Id,
			f.Name,
			fmt.Sprintf("%d", f.Size),
			f.UploadedAt.Format("2006-01-02 15:04:05"),
			f.CreatedByEmail,
		}
	}
	return rows
}

var _ output.TableData = (*aiFilesTable)(nil)
