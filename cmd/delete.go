package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <document-id>",
	Short: "Move a document to trash",
	Long: `Soft-delete a document by moving it to Craft trash.

This uses DELETE /documents (you can restore via documents/move).

Use --dry-run to preview what would be deleted without making changes.

Examples:
  craft delete abc123
  craft delete abc123 --dry-run    # Preview without deleting
  craft delete abc123 -q           # Silent delete`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docID := args[0]
		if err := validateResourceID(docID, "document-id"); err != nil {
			return err
		}

		if isDryRun() {
			client, err := getAPIClient()
			if err != nil {
				return err
			}
			doc, err := client.GetDocument(docID)
			if err != nil {
				return fmt.Errorf("document not found: %s", docID)
			}
			return dryRunOutput("delete", map[string]interface{}{
				"id": doc.ID, "title": doc.Title, "reversible": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.DeleteDocument(docID); err != nil {
			return err
		}

		outputDeleted(docID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
