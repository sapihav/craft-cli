package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	moveTargetFolder   string
	moveTargetLocation string
)

var moveCmd = &cobra.Command{
	Use:   "move [document-id]",
	Short: "Move a document to a folder or location",
	Long: `Move a document to a different folder or special location.

Locations:
  unsorted   - Move to unsorted documents
  trash      - Move to trash

Examples:
  craft move abc123 --to-folder def456    # Move to folder
  craft move abc123 --to-location unsorted # Move to unsorted
  craft move abc123 --to-location trash    # Move to trash`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateResourceID(args[0], "document ID"); err != nil {
			return err
		}
		if moveTargetFolder == "" && moveTargetLocation == "" {
			return fmt.Errorf("either --to-folder or --to-location is required")
		}
		if moveTargetFolder != "" {
			if err := validateResourceID(moveTargetFolder, "target folder ID"); err != nil {
				return err
			}
		}

		if isDryRun() {
			target := map[string]interface{}{"id": args[0]}
			if moveTargetFolder != "" {
				target["destination_folder"] = moveTargetFolder
			} else {
				target["destination_location"] = moveTargetLocation
			}
			return dryRunOutput("move", target)
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		docID := args[0]
		if err := client.MoveDocument(docID, moveTargetFolder, moveTargetLocation); err != nil {
			return err
		}

		if !isQuiet() {
			if moveTargetFolder != "" {
				fmt.Printf("Document %s moved to folder %s\n", docID, moveTargetFolder)
			} else {
				fmt.Printf("Document %s moved to %s\n", docID, moveTargetLocation)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
	moveCmd.Flags().StringVar(&moveTargetFolder, "to-folder", "", "Target folder ID")
	moveCmd.Flags().StringVar(&moveTargetLocation, "to-location", "", "Target location (unsorted, trash)")
}
