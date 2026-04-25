package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
)

// readSchemaInput accepts either inline JSON or a `@filename` reference and
// returns the parsed JSON value. Used by `collections create --schema` and
// `collections schema update --schema`.
func readSchemaInput(input string) (interface{}, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, fmt.Errorf("schema input is empty")
	}
	raw := []byte(trimmed)
	if strings.HasPrefix(trimmed, "@") {
		path := strings.TrimSpace(trimmed[1:])
		if path == "" {
			return nil, fmt.Errorf("--schema @filename: filename is empty")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file %q: %w", path, err)
		}
		raw = data
	}
	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("invalid schema JSON: %w", err)
	}
	return parsed, nil
}

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Manage collections",
	Long: `Manage Craft collections (databases) - list, view schema, and manage items.

Examples:
  craft collections list                                  # List all collections
  craft collections list --document ID                    # List collections in document
  craft collections schema COLLECTION_ID                  # Get collection schema
  craft collections items COLLECTION_ID                   # List items in collection
  craft collections add COLLECTION_ID --title "New Item"  # Add item
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"key":"value"}'
  craft collections delete COLLECTION_ID --item ITEM_ID   # Delete item`,
}

var (
	collectionDocumentID    string
	collectionSchemaFormat  string
	collectionItemDepth     int
	collectionItemTitle     string
	collectionItemProps     string
	collectionAllowNew      bool
	collectionItemID        string
	collectionCreateName    string
	collectionCreateDesc    string
	collectionCreateIcon    string
	collectionSchemaInput   string
)

var collectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all collections",
	Long: `List all collections in your Craft space.

Use --document to filter collections by document ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collections, err := client.GetCollections(collectionDocumentID)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == FormatJSON {
			return outputJSON(collections)
		}
		return outputCollections(collections.Items, format)
	},
}

var collectionsSchemaCmd = &cobra.Command{
	Use:   "schema [collection-id]",
	Short: "Get collection schema",
	Long: `Get the schema for a collection. Output is always JSON.

Schema formats:
  schema  - Standard schema format (default)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if err := validateResourceID(collectionID, "collection ID"); err != nil {
			return err
		}
		schema, err := client.GetCollectionSchema(collectionID, collectionSchemaFormat)
		if err != nil {
			return err
		}

		return outputJSON(schema)
	},
}

var collectionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new collection in a document",
	Long: `Create a new collection (database) inside a Craft document.

Required flags:
  --doc       The document ID where the collection will be created
  --name      The display name of the collection

Optional flags:
  --description   Short description shown under the name
  --icon          Emoji or icon identifier
  --schema        Initial schema as inline JSON or @file (e.g. @schema.json)

Examples:
  craft collections create --doc DOC_ID --name "Tasks"
  craft collections create --doc DOC_ID --name "Tasks" --icon "✅" --description "All tasks"
  craft collections create --doc DOC_ID --name "Tasks" --schema @schema.json
  craft collections create --doc DOC_ID --name "Tasks" --schema '{"properties":[]}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateResourceID(collectionDocumentID, "document ID"); err != nil {
			return err
		}

		var schemaPayload interface{}
		if collectionSchemaInput != "" {
			parsed, err := readSchemaInput(collectionSchemaInput)
			if err != nil {
				return err
			}
			schemaPayload = parsed
		}

		req := &models.CreateCollectionRequest{
			DocumentID:  collectionDocumentID,
			Name:        collectionCreateName,
			Description: collectionCreateDesc,
			Icon:        collectionCreateIcon,
			Schema:      schemaPayload,
		}

		if isDryRun() {
			return dryRunOutput("create collection", map[string]interface{}{
				"method":      "POST",
				"path":        "/collections",
				"document_id": req.DocumentID,
				"name":        req.Name,
				"description": req.Description,
				"icon":        req.Icon,
				"schema":      req.Schema,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		col, err := client.CreateCollection(req)
		if err != nil {
			return err
		}

		if isQuiet() {
			fmt.Println(col.ID)
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) {
			return outputJSON(col)
		}

		fmt.Printf("Collection created: %s (ID: %s)\n", col.Name, col.ID)
		return nil
	},
}

var collectionsSchemaUpdateCmd = &cobra.Command{
	Use:   "update <collection-id>",
	Short: "Replace a collection's schema",
	Long: `Replace the schema for a collection.

The --schema flag is required. It accepts either inline JSON or @filename.
The schema is parsed client-side before sending to fail fast on bad input.

Examples:
  craft collections schema update COL_ID --schema @schema.json
  craft collections schema update COL_ID --schema '{"properties":[{"key":"k","name":"K","type":"text"}]}'
  craft collections schema update COL_ID --schema @schema.json --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collectionID := args[0]
		if err := validateResourceID(collectionID, "collection ID"); err != nil {
			return err
		}

		parsed, err := readSchemaInput(collectionSchemaInput)
		if err != nil {
			return err
		}

		if isDryRun() {
			return dryRunOutput("update collection schema", map[string]interface{}{
				"method":        "PUT",
				"path":          fmt.Sprintf("/collections/%s/schema", collectionID),
				"collection_id": collectionID,
				"schema":        parsed,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		updated, err := client.UpdateCollectionSchema(collectionID, parsed)
		if err != nil {
			return err
		}

		if isQuiet() {
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) {
			if updated != nil {
				return outputJSON(updated)
			}
			return outputJSON(map[string]interface{}{"ok": true, "collectionId": collectionID})
		}

		fmt.Printf("Schema updated for collection %s\n", collectionID)
		return nil
	},
}

var collectionsItemsCmd = &cobra.Command{
	Use:   "items [collection-id]",
	Short: "List items in a collection",
	Long: `List all items in a collection.

Use --depth to control the depth of nested content returned.
Default depth is -1 (no limit).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if err := validateResourceID(collectionID, "collection ID"); err != nil {
			return err
		}
		items, err := client.GetCollectionItems(collectionID, collectionItemDepth)
		if err != nil {
			return err
		}

		format := getOutputFormat()
		if format == FormatJSON {
			return outputJSON(items)
		}
		return outputCollectionItems(items.Items, format)
	},
}

var collectionsAddCmd = &cobra.Command{
	Use:   "add [collection-id]",
	Short: "Add an item to a collection",
	Long: `Add a new item to a collection.

Examples:
  craft collections add COLLECTION_ID --title "New Item"
  craft collections add COLLECTION_ID --title "Item" --properties '{"Status":"Active","Priority":"High"}'
  craft collections add COLLECTION_ID --title "Item" --properties '{"Tag":"new-value"}' --allow-new-options`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			return dryRunOutput("add collection item", map[string]interface{}{
				"collection_id": args[0], "title": collectionItemTitle,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if err := validateResourceID(collectionID, "collection ID"); err != nil {
			return err
		}

		var props map[string]interface{}
		if collectionItemProps != "" {
			if err := json.Unmarshal([]byte(collectionItemProps), &props); err != nil {
				return fmt.Errorf("invalid --properties JSON: %w", err)
			}
		}

		result, err := client.AddCollectionItem(collectionID, collectionItemTitle, props, collectionAllowNew)
		if err != nil {
			return err
		}

		if isQuiet() {
			if len(result.Items) > 0 {
				fmt.Println(result.Items[0].ID)
			}
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) {
			return outputJSON(result)
		}

		if len(result.Items) > 0 {
			fmt.Printf("Item added: %s (ID: %s)\n", result.Items[0].Title, result.Items[0].ID)
		} else {
			fmt.Println("Item added")
		}
		return nil
	},
}

var collectionsUpdateCmd = &cobra.Command{
	Use:   "update [collection-id]",
	Short: "Update an item in a collection",
	Long: `Update an existing item in a collection.

Examples:
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"Status":"Done"}'
  craft collections update COLLECTION_ID --item ITEM_ID --properties '{"Tag":"new"}' --allow-new-options`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			return dryRunOutput("update collection item", map[string]interface{}{
				"collection_id": args[0], "item_id": collectionItemID,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if err := validateResourceID(collectionID, "collection ID"); err != nil {
			return err
		}
		if err := validateResourceID(collectionItemID, "item ID"); err != nil {
			return err
		}

		var props map[string]interface{}
		if err := json.Unmarshal([]byte(collectionItemProps), &props); err != nil {
			return fmt.Errorf("invalid --properties JSON: %w", err)
		}

		if err := client.UpdateCollectionItem(collectionID, collectionItemID, props, collectionAllowNew); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Item %s updated in collection %s\n", collectionItemID, collectionID)
		}
		return nil
	},
}

var collectionsDeleteCmd = &cobra.Command{
	Use:   "delete [collection-id]",
	Short: "Delete an item from a collection",
	Long:  "Delete an item from a collection by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			return dryRunOutput("delete collection item", map[string]interface{}{
				"collection_id": args[0], "item_id": collectionItemID, "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		collectionID := args[0]
		if err := validateResourceID(collectionID, "collection ID"); err != nil {
			return err
		}
		if err := validateResourceID(collectionItemID, "item ID"); err != nil {
			return err
		}
		if err := client.DeleteCollectionItem(collectionID, collectionItemID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Item %s deleted from collection %s\n", collectionItemID, collectionID)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(collectionsCmd)

	collectionsCmd.AddCommand(collectionsListCmd)
	collectionsListCmd.Flags().StringVar(&collectionDocumentID, "document", "", "Filter by document ID")

	collectionsCmd.AddCommand(collectionsSchemaCmd)
	collectionsSchemaCmd.Flags().StringVar(&collectionSchemaFormat, "schema-format", "schema", "Schema format (default: schema)")

	// `collections schema update <id> --schema @file|json` lives as a child of
	// the existing `collections schema [id]` reader to avoid a breaking change.
	collectionsSchemaCmd.AddCommand(collectionsSchemaUpdateCmd)
	collectionsSchemaUpdateCmd.Flags().StringVar(&collectionSchemaInput, "schema", "", "New schema as inline JSON or @file (required)")
	collectionsSchemaUpdateCmd.MarkFlagRequired("schema")

	collectionsCmd.AddCommand(collectionsCreateCmd)
	collectionsCreateCmd.Flags().StringVar(&collectionDocumentID, "doc", "", "Document ID to create the collection in (required)")
	collectionsCreateCmd.Flags().StringVar(&collectionCreateName, "name", "", "Collection name (required)")
	collectionsCreateCmd.Flags().StringVar(&collectionCreateDesc, "description", "", "Collection description")
	collectionsCreateCmd.Flags().StringVar(&collectionCreateIcon, "icon", "", "Collection icon (emoji)")
	collectionsCreateCmd.Flags().StringVar(&collectionSchemaInput, "schema", "", "Initial schema as inline JSON or @file")
	collectionsCreateCmd.MarkFlagRequired("doc")
	collectionsCreateCmd.MarkFlagRequired("name")

	collectionsCmd.AddCommand(collectionsItemsCmd)
	collectionsItemsCmd.Flags().IntVar(&collectionItemDepth, "depth", -1, "Max depth of nested content (-1 for no limit)")

	collectionsCmd.AddCommand(collectionsAddCmd)
	collectionsAddCmd.Flags().StringVar(&collectionItemTitle, "title", "", "Item title (required)")
	collectionsAddCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Item properties as JSON string")
	collectionsAddCmd.Flags().BoolVar(&collectionAllowNew, "allow-new-options", false, "Allow creating new options for select properties")
	collectionsAddCmd.MarkFlagRequired("title")

	collectionsCmd.AddCommand(collectionsUpdateCmd)
	collectionsUpdateCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to update (required)")
	collectionsUpdateCmd.Flags().StringVar(&collectionItemProps, "properties", "", "Item properties as JSON string (required)")
	collectionsUpdateCmd.Flags().BoolVar(&collectionAllowNew, "allow-new-options", false, "Allow creating new options for select properties")
	collectionsUpdateCmd.MarkFlagRequired("item")
	collectionsUpdateCmd.MarkFlagRequired("properties")

	collectionsCmd.AddCommand(collectionsDeleteCmd)
	collectionsDeleteCmd.Flags().StringVar(&collectionItemID, "item", "", "Item ID to delete (required)")
	collectionsDeleteCmd.MarkFlagRequired("item")
}

// outputCollections prints collections in the specified format
func outputCollections(collections []models.Collection, format string) error {
	switch format {
	case FormatCompact:
		return outputJSON(collections)
	case "table":
		return outputCollectionsTable(collections)
	case "markdown":
		return outputCollectionsMarkdown(collections)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputCollectionsTable prints collections as a table
func outputCollectionsTable(collections []models.Collection) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tNAME\tITEMS\tDOCUMENT")
		fmt.Fprintln(w, "---\t----\t-----\t--------")
	}

	for _, c := range collections {
		docID := c.DocumentID
		if docID == "" {
			docID = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", c.ID, c.Name, c.ItemCount, docID)
	}

	return w.Flush()
}

// outputCollectionsMarkdown prints collections as markdown
func outputCollectionsMarkdown(collections []models.Collection) error {
	fmt.Println("# Collections")
	for _, c := range collections {
		fmt.Printf("## %s\n", c.Name)
		fmt.Printf("- **ID**: %s\n", c.ID)
		fmt.Printf("- **Items**: %d\n", c.ItemCount)
		if c.DocumentID != "" {
			fmt.Printf("- **Document**: %s\n", c.DocumentID)
		}
		fmt.Println()
	}
	return nil
}

// outputCollectionItems prints collection items in the specified format
func outputCollectionItems(items []models.CollectionItem, format string) error {
	switch format {
	case FormatCompact:
		return outputJSON(items)
	case "table":
		return outputCollectionItemsTable(items)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// outputCollectionItemsTable prints collection items as a table
func outputCollectionItemsTable(items []models.CollectionItem) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !hasNoHeaders() {
		fmt.Fprintln(w, "ID\tTITLE\tPROPERTIES")
		fmt.Fprintln(w, "---\t-----\t----------")
	}

	for _, item := range items {
		title := item.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		propCount := len(item.Properties)
		propSummary := fmt.Sprintf("%d properties", propCount)
		if propCount == 0 {
			propSummary = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", item.ID, title, propSummary)
	}

	return w.Flush()
}
