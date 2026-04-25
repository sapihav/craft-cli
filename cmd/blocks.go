package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ashrafali/craft-cli/internal/models"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var blocksCmd = &cobra.Command{
	Use:   "blocks",
	Short: "Manage document blocks",
	Long: `Manage blocks within Craft documents - get, add, update, delete, and move blocks.

Examples:
  craft blocks get BLOCK_ID                           # Get a specific block
  craft blocks add PAGE_ID --markdown "Hello"         # Add block at end
  craft blocks add PAGE_ID -m "Red heading" --text-style h1 --color "#ef052a"
  craft blocks add PAGE_ID --type line --line-style strong
  craft blocks add PAGE_ID --json '[{"type":"text","textStyle":"h1","markdown":"Heading"}]'
  craft blocks update BLOCK_ID --color "#0400ff" --font serif
  craft blocks update --json '[{"id":"ID","textStyle":"h2"}]'
  craft blocks delete BLOCK_ID                        # Delete a block
  craft blocks move BLOCK_ID --to PAGE_ID --pos end   # Move block`,
}

var blocksGetCmd = &cobra.Command{
	Use:   "get [block-id]",
	Short: "Get a specific block",
	Long: `Retrieve a specific block by ID or daily note date.

When --date is provided, fetches the daily note for that date instead of a block ID.
Use --depth to control how many levels of children to include.
Use --metadata to include created/modified timestamps.

Examples:
  craft blocks get BLOCK_ID                        # Get a specific block
  craft blocks get BLOCK_ID --depth 0              # Block only, no children
  craft blocks get BLOCK_ID --metadata             # Include metadata
  craft blocks get --date today                    # Today's daily note
  craft blocks get --date yesterday                # Yesterday's daily note
  craft blocks get --date 2024-01-15               # Specific date
  craft blocks get --date today --depth 1          # Daily note, 1 level deep`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		var block *models.Block

		if blockDate != "" {
			block, err = client.GetBlockByDate(blockDate, blockDepth, blockMetadata)
		} else {
			if len(args) == 0 {
				return fmt.Errorf("block-id is required when not using --date")
			}
			block, err = client.GetBlockWithOptions(args[0], blockDepth, blockMetadata)
		}
		if err != nil {
			return err
		}

		format := getOutputFormat()
		switch format {
		case FormatStructured, FormatJSON, FormatCompact:
			return outputJSON(block)
		case FormatCraft:
			var sb strings.Builder
			renderBlockCraft(&sb, block, 0)
			fmt.Print(sb.String())
			return nil
		case FormatRich:
			var sb strings.Builder
			renderBlockRich(&sb, block, 0)
			fmt.Print(sb.String())
			return nil
		default:
			fmt.Println(block.Markdown)
			return nil
		}
	},
}

var (
	blockMarkdown   string
	blockPosition   string
	blockSiblingID  string
	blockTargetPage string
	blockDate       string
	blockDepth      int
	blockMetadata   bool

	// JSON mode flags
	blockJSON  string
	blockStdin bool

	// Styling flags (shared between add and update)
	blockType             string
	blockTextStyle        string
	blockListStyle        string
	blockDecorations      string
	blockColor            string
	blockFont             string
	blockTextAlignment    string
	blockIndentationLevel string
	blockLineStyle        string
	blockLanguage         string
	blockRawCode          string
	blockURL              string
	blockAltText          string
	blockFileName         string
	blockTitle            string
	blockDescription      string
	blockLayout           string
	blockBlockLayout      string
	blockCardLayout       string
	blockTaskState        string
	blockScheduleDate     string
	blockDeadlineDate     string
)

var blocksAddCmd = &cobra.Command{
	Use:   "add [page-id]",
	Short: "Add a block to a document",
	Long: `Add a new block to a document at a specified position.

Three input modes:
  1. Flags:  --markdown "text" with optional styling flags
  2. JSON:   --json '[{"type":"text","markdown":"..."}]'
  3. Stdin:  echo '[...]' | craft blocks add PAGE_ID --stdin

Positions:
  start  - Add at the beginning of the page
  end    - Add at the end of the page (default)
  before - Add before a sibling block (requires --sibling)
  after  - Add after a sibling block (requires --sibling)

When --date is provided, adds the block to the daily note for that date.
The page-id argument is not required when using --sibling or --date.

Styling Examples:
  craft blocks add PAGE_ID --markdown "Hello world"
  craft blocks add PAGE_ID --markdown "# Title" --text-style h1 --color "#ef052a"
  craft blocks add PAGE_ID --type line --line-style strong
  craft blocks add PAGE_ID --markdown "Note" --decorations callout --color "#00ca85"
  craft blocks add PAGE_ID --markdown "Centered" --text-alignment center --font serif
  craft blocks add PAGE_ID --type code --language python --raw-code "print('hi')"
  craft blocks add --date today --markdown "Daily log" --list-style bullet

JSON Examples:
  craft blocks add PAGE_ID --json '[{"type":"text","textStyle":"h1","markdown":"# Heading"}]'
  craft blocks add PAGE_ID --json '{"type":"line","lineStyle":"strong"}'
  echo '[{"type":"text","markdown":"piped"}]' | craft blocks add PAGE_ID --stdin`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getAPIClient()
		if err != nil {
			return err
		}

		var blocks []map[string]interface{}

		switch {
		case blockStdin:
			blocks, err = parseBlocksFromStdin()
			if err != nil {
				return err
			}
		case blockJSON != "":
			blocks, err = parseBlocksJSON(blockJSON)
			if err != nil {
				return err
			}
		default:
			block := buildBlockFromFlags(cmd)
			if len(block) == 0 {
				return fmt.Errorf("provide --markdown, --json, --stdin, or a block type like --type line")
			}
			blocks = []map[string]interface{}{block}
		}

		position, err := buildAddPosition(cmd, args)
		if err != nil {
			return err
		}

		result, err := client.AddBlocksJSON(blocks, position)
		if err != nil {
			return err
		}

		if isQuiet() {
			for _, b := range result {
				fmt.Println(b.ID)
			}
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) {
			return outputJSON(result)
		}
		for _, b := range result {
			fmt.Printf("Block created: %s\n", b.ID)
		}
		return nil
	},
}

var blocksUpdateCmd = &cobra.Command{
	Use:   "update [block-id]",
	Short: "Update a block's content or styling",
	Long: `Update an existing block's content, styling, or both.

Three input modes:
  1. Flags:  BLOCK_ID --markdown "text" with optional styling flags
  2. JSON:   --json '[{"id":"ID","color":"#ff0000"}]'
  3. Stdin:  echo '[...]' | craft blocks update --stdin

Flag Examples:
  craft blocks update BLOCK_ID --markdown "New text"
  craft blocks update BLOCK_ID --color "#ff0000" --font serif
  craft blocks update BLOCK_ID --text-style h1 --decorations callout
  craft blocks update BLOCK_ID --text-alignment center

JSON Examples:
  craft blocks update --json '[{"id":"ID","textStyle":"h2","color":"#0400ff"}]'
  craft blocks update --json '{"id":"ID","markdown":"Updated"}'
  echo '[{"id":"ID","color":"#00ca85"}]' | craft blocks update --stdin`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var blocks []map[string]interface{}
		var err error

		switch {
		case blockStdin:
			blocks, err = parseBlocksFromStdin()
			if err != nil {
				return err
			}
			for _, b := range blocks {
				if _, ok := b["id"]; !ok {
					return fmt.Errorf("each block in JSON must have an \"id\" field for update")
				}
			}
		case blockJSON != "":
			blocks, err = parseBlocksJSON(blockJSON)
			if err != nil {
				return err
			}
			for _, b := range blocks {
				if _, ok := b["id"]; !ok {
					return fmt.Errorf("each block in JSON must have an \"id\" field for update")
				}
			}
		default:
			if len(args) == 0 {
				return fmt.Errorf("block-id argument is required when not using --json or --stdin")
			}
			block := buildUpdateFromFlags(cmd, args[0])
			if len(block) <= 1 { // only "id" present
				return fmt.Errorf("at least one property to update is required (e.g. --markdown, --color, --text-style)")
			}
			blocks = []map[string]interface{}{block}
		}

		if isDryRun() {
			ids := []string{}
			for _, b := range blocks {
				if id, ok := b["id"].(string); ok {
					ids = append(ids, id)
				}
			}
			return dryRunOutput("update blocks", map[string]interface{}{
				"block_ids": ids, "count": len(blocks),
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		if err := client.UpdateBlocksJSON(blocks); err != nil {
			return err
		}

		if !isQuiet() {
			for _, b := range blocks {
				if id, ok := b["id"].(string); ok {
					fmt.Printf("Block %s updated\n", id)
				}
			}
		}
		return nil
	},
}

var blocksDeleteCmd = &cobra.Command{
	Use:   "delete [block-id]",
	Short: "Delete a block",
	Long:  "Delete a specific block from a document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDryRun() {
			return dryRunOutput("delete block", map[string]interface{}{
				"id": args[0], "destructive": true,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		blockID := args[0]
		if err := validateResourceID(blockID, "block ID"); err != nil {
			return err
		}
		if err := client.DeleteBlock(blockID); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Block %s deleted\n", blockID)
		}
		return nil
	},
}

var blocksMoveCmd = &cobra.Command{
	Use:   "move [block-id]",
	Short: "Move a block to a new location",
	Long: `Move a block to a different page or position.

Examples:
  craft blocks move BLOCK_ID --to PAGE_ID --position end
  craft blocks move BLOCK_ID --to PAGE_ID --position start`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if blockTargetPage == "" {
			return fmt.Errorf("--to is required")
		}
		if blockPosition == "" {
			blockPosition = "end"
		}

		if isDryRun() {
			return dryRunOutput("move block", map[string]interface{}{
				"id": args[0], "target_page": blockTargetPage, "position": blockPosition,
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		blockID := args[0]
		if err := validateResourceID(blockID, "block ID"); err != nil {
			return err
		}
		if err := validateResourceID(blockTargetPage, "target page ID"); err != nil {
			return err
		}
		if err := client.MoveBlock(blockID, blockTargetPage, blockPosition); err != nil {
			return err
		}

		if !isQuiet() {
			fmt.Printf("Block %s moved to %s\n", blockID, blockTargetPage)
		}
		return nil
	},
}

var blocksRevertCmd = &cobra.Command{
	Use:   "revert [block-id]",
	Short: "Revert a block to its previous state",
	Long: `Revert a block change, restoring its previous state.

Mirrors the MCP blocks_revert tool. Returns the post-revert block when the
API echoes one. Use --dry-run to preview the planned mutation.

Examples:
  craft blocks revert BLOCK_ID
  craft blocks revert BLOCK_ID --dry-run
  craft blocks revert BLOCK_ID --quiet`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		blockID := args[0]
		if err := validateResourceID(blockID, "block ID"); err != nil {
			return err
		}

		if isDryRun() {
			return dryRunOutput("revert block", map[string]interface{}{
				"id":     blockID,
				"method": "POST",
				"path":   fmt.Sprintf("/blocks/%s/revert", blockID),
			})
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		block, err := client.RevertBlock(blockID)
		if err != nil {
			return err
		}

		if isQuiet() {
			return nil
		}

		format := getOutputFormat()
		if isJSONFormat(format) && block != nil {
			return outputJSON(block)
		}
		fmt.Printf("Block %s reverted\n", blockID)
		return nil
	},
}

var (
	blockImageOut    string
	blockImageFormat string
)

// stdoutIsTTY reports whether stdout is a terminal. Overridable for tests.
var stdoutIsTTY = func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

var blocksImageCmd = &cobra.Command{
	Use:   "image [block-id]",
	Short: "Fetch the binary image for an image block",
	Long: `Fetch the binary image content for an image block (MCP image_view parity).

Default: writes raw image bytes to stdout (suitable for piping to file -,
ImageMagick, etc.). Refuses to write binary to a TTY without --out, to avoid
clobbering the terminal.

Examples:
  craft blocks image BLOCK_ID --out photo.png
  craft blocks image BLOCK_ID --image-format jpeg --out photo.jpg
  craft blocks image BLOCK_ID > photo.png
  craft blocks image BLOCK_ID --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		blockID := args[0]
		if err := validateResourceID(blockID, "block ID"); err != nil {
			return err
		}

		switch blockImageFormat {
		case "", "png", "jpeg", "jpg", "webp", "gif":
			// allowed
		default:
			return fmt.Errorf("invalid --image-format %q (allowed: png, jpeg, webp, gif)", blockImageFormat)
		}

		if isDryRun() {
			return dryRunOutput("fetch block image", map[string]interface{}{
				"id":     blockID,
				"method": "GET",
				"path":   fmt.Sprintf("/blocks/%s/image", blockID),
				"format": blockImageFormat,
				"out":    blockImageOut,
			})
		}

		// Refuse to dump binary into an interactive terminal.
		if blockImageOut == "" && stdoutIsTTY() {
			return fmt.Errorf("refusing to write binary image to a TTY; use --out FILE or redirect stdout")
		}

		client, err := getAPIClient()
		if err != nil {
			return err
		}

		data, contentType, err := client.GetBlockImage(blockID, blockImageFormat)
		if err != nil {
			return err
		}

		if blockImageOut != "" {
			if err := os.WriteFile(blockImageOut, data, 0o644); err != nil {
				return fmt.Errorf("failed to write %s: %w", blockImageOut, err)
			}
			if isQuiet() {
				return nil
			}
			result := map[string]interface{}{
				"result": map[string]interface{}{
					"path":         blockImageOut,
					"content_type": contentType,
					"size_bytes":   len(data),
				},
			}
			return outputJSON(result)
		}

		_, err = os.Stdout.Write(data)
		return err
	},
}

// ========== Helper Functions ==========

// parseBlocksJSON parses a JSON string into a slice of block maps.
// Accepts either a JSON array or a single JSON object.
func parseBlocksJSON(jsonStr string) ([]map[string]interface{}, error) {
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" {
		return nil, fmt.Errorf("empty JSON input")
	}

	// Try array first
	if jsonStr[0] == '[' {
		var blocks []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &blocks); err != nil {
			return nil, fmt.Errorf("invalid JSON array: %w", err)
		}
		return blocks, nil
	}

	// Try single object
	var block map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &block); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return []map[string]interface{}{block}, nil
}

// parseBlocksFromStdin reads JSON block data from stdin.
func parseBlocksFromStdin() ([]map[string]interface{}, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}
	return parseBlocksJSON(string(data))
}

// buildAddPosition constructs the position map from command flags and args.
func buildAddPosition(cmd *cobra.Command, args []string) (map[string]interface{}, error) {
	pos := make(map[string]interface{})

	position := blockPosition
	if position == "" {
		position = "end"
	}
	pos["position"] = position

	if blockSiblingID != "" {
		if position != "before" && position != "after" {
			return nil, fmt.Errorf("--position must be 'before' or 'after' when using --sibling")
		}
		pos["siblingId"] = blockSiblingID
		return pos, nil
	}

	if blockDate != "" {
		pos["date"] = blockDate
		return pos, nil
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("page-id is required when not using --sibling or --date")
	}
	pos["pageId"] = args[0]
	return pos, nil
}

// buildBlockFromFlags constructs a block map from individual CLI flags.
func buildBlockFromFlags(cmd *cobra.Command) map[string]interface{} {
	block := make(map[string]interface{})

	// Type defaults to "text" only if we have content
	if cmd.Flags().Changed("type") {
		block["type"] = blockType
	}

	if cmd.Flags().Changed("markdown") {
		block["markdown"] = blockMarkdown
	}

	addStylingToMap(cmd, block)

	// If no type was set but we have some properties, default to "text"
	if _, hasType := block["type"]; !hasType && len(block) > 0 {
		block["type"] = "text"
	}

	return block
}

// buildUpdateFromFlags constructs an update block map from CLI flags and the block ID.
func buildUpdateFromFlags(cmd *cobra.Command, blockID string) map[string]interface{} {
	block := make(map[string]interface{})
	block["id"] = blockID

	if cmd.Flags().Changed("markdown") {
		block["markdown"] = blockMarkdown
	}

	addStylingToMap(cmd, block)

	return block
}

// addStylingToMap adds styling properties to a block map based on which flags were changed.
func addStylingToMap(cmd *cobra.Command, block map[string]interface{}) {
	if cmd.Flags().Changed("text-style") {
		block["textStyle"] = blockTextStyle
	}
	if cmd.Flags().Changed("list-style") {
		block["listStyle"] = blockListStyle
	}
	if cmd.Flags().Changed("decorations") {
		parts := strings.Split(blockDecorations, ",")
		var trimmed []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				trimmed = append(trimmed, p)
			}
		}
		if len(trimmed) > 0 {
			block["decorations"] = trimmed
		}
	}
	if cmd.Flags().Changed("color") {
		block["color"] = blockColor
	}
	if cmd.Flags().Changed("font") {
		block["font"] = blockFont
	}
	if cmd.Flags().Changed("text-alignment") {
		block["textAlignment"] = blockTextAlignment
	}
	if cmd.Flags().Changed("indentation-level") {
		if level, err := strconv.Atoi(blockIndentationLevel); err == nil {
			block["indentationLevel"] = level
		}
	}
	if cmd.Flags().Changed("line-style") {
		block["lineStyle"] = blockLineStyle
	}
	if cmd.Flags().Changed("language") {
		block["language"] = blockLanguage
	}
	if cmd.Flags().Changed("raw-code") {
		block["rawCode"] = blockRawCode
	}
	if cmd.Flags().Changed("url") {
		block["url"] = blockURL
	}
	if cmd.Flags().Changed("alt-text") {
		block["altText"] = blockAltText
	}
	if cmd.Flags().Changed("file-name") {
		block["fileName"] = blockFileName
	}
	if cmd.Flags().Changed("title") {
		block["title"] = blockTitle
	}
	if cmd.Flags().Changed("description") {
		block["description"] = blockDescription
	}
	if cmd.Flags().Changed("layout") {
		block["layout"] = blockLayout
	}
	if cmd.Flags().Changed("block-layout") {
		block["blockLayout"] = blockBlockLayout
	}
	if cmd.Flags().Changed("card-layout") {
		block["cardLayout"] = blockCardLayout
	}

	// Task info: build only if any task flag is set
	taskInfo := make(map[string]interface{})
	if cmd.Flags().Changed("task-state") {
		taskInfo["state"] = blockTaskState
	}
	if cmd.Flags().Changed("schedule-date") {
		taskInfo["scheduleDate"] = blockScheduleDate
	}
	if cmd.Flags().Changed("deadline-date") {
		taskInfo["deadlineDate"] = blockDeadlineDate
	}
	if len(taskInfo) > 0 {
		block["taskInfo"] = taskInfo
	}
}

// registerStylingFlags adds all styling flags to a command.
// includeType controls whether --type is registered (add yes, update no since type is immutable).
func registerStylingFlags(cmd *cobra.Command, includeType bool) {
	if includeType {
		cmd.Flags().StringVar(&blockType, "type", "text", "Block type: text, page, code, line, richUrl, image, file")
	}
	cmd.Flags().StringVar(&blockTextStyle, "text-style", "", "Text style: h1, h2, h3, h4, caption, body, page, card")
	cmd.Flags().StringVar(&blockListStyle, "list-style", "", "List style: none, bullet, numbered, task, toggle")
	cmd.Flags().StringVar(&blockDecorations, "decorations", "", "Decorations (comma-separated): callout, quote")
	cmd.Flags().StringVar(&blockColor, "color", "", "Block color as #RRGGBB hex (e.g. #ef052a)")
	cmd.Flags().StringVar(&blockFont, "font", "", "Font: system, serif, mono, rounded")
	cmd.Flags().StringVar(&blockTextAlignment, "text-alignment", "", "Text alignment: left, center, right, justify")
	cmd.Flags().StringVar(&blockIndentationLevel, "indentation-level", "", "Indentation level: 0-5")
	cmd.Flags().StringVar(&blockLineStyle, "line-style", "", "Line/divider style: strong, regular, light, extraLight, pageBreak")
	cmd.Flags().StringVar(&blockLanguage, "language", "", "Code block language (e.g. python, javascript, math_formula)")
	cmd.Flags().StringVar(&blockRawCode, "raw-code", "", "Raw code content for code blocks")
	cmd.Flags().StringVar(&blockURL, "url", "", "URL for richUrl, image, video, or file blocks")
	cmd.Flags().StringVar(&blockAltText, "alt-text", "", "Alt text for image/video blocks")
	cmd.Flags().StringVar(&blockFileName, "file-name", "", "File name for file blocks")
	cmd.Flags().StringVar(&blockTitle, "title", "", "Title for richUrl blocks")
	cmd.Flags().StringVar(&blockDescription, "description", "", "Description for richUrl blocks")
	cmd.Flags().StringVar(&blockLayout, "layout", "", "Layout for richUrl: small, regular, card")
	cmd.Flags().StringVar(&blockBlockLayout, "block-layout", "", "Layout for file blocks: small, regular, card")
	cmd.Flags().StringVar(&blockCardLayout, "card-layout", "", "Card layout: small, square, regular, large")
	cmd.Flags().StringVar(&blockTaskState, "task-state", "", "Task state: todo, done, canceled")
	cmd.Flags().StringVar(&blockScheduleDate, "schedule-date", "", "Task schedule date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&blockDeadlineDate, "deadline-date", "", "Task deadline date (YYYY-MM-DD)")
}

func init() {
	rootCmd.AddCommand(blocksCmd)

	blocksCmd.AddCommand(blocksGetCmd)
	blocksGetCmd.Flags().StringVar(&blockDate, "date", "", "Daily note date (today, tomorrow, yesterday, YYYY-MM-DD)")
	blocksGetCmd.Flags().IntVar(&blockDepth, "depth", -1, "Max depth (-1 for all, 0 for block only)")
	blocksGetCmd.Flags().BoolVar(&blockMetadata, "metadata", false, "Include metadata (created/modified info)")

	blocksCmd.AddCommand(blocksAddCmd)
	blocksAddCmd.Flags().StringVarP(&blockMarkdown, "markdown", "m", "", "Markdown content for the block")
	blocksAddCmd.Flags().StringVarP(&blockPosition, "position", "p", "end", "Position: start, end, before, after")
	blocksAddCmd.Flags().StringVar(&blockSiblingID, "sibling", "", "Sibling block ID for relative positioning")
	blocksAddCmd.Flags().StringVar(&blockDate, "date", "", "Daily note date (today, tomorrow, yesterday, YYYY-MM-DD)")
	blocksAddCmd.Flags().StringVar(&blockJSON, "json", "", "Block(s) as JSON (array or single object)")
	blocksAddCmd.Flags().BoolVar(&blockStdin, "stdin", false, "Read block JSON from stdin")
	registerStylingFlags(blocksAddCmd, true)

	blocksCmd.AddCommand(blocksUpdateCmd)
	blocksUpdateCmd.Flags().StringVarP(&blockMarkdown, "markdown", "m", "", "New markdown content")
	blocksUpdateCmd.Flags().StringVar(&blockJSON, "json", "", "Block(s) as JSON with \"id\" fields (array or single object)")
	blocksUpdateCmd.Flags().BoolVar(&blockStdin, "stdin", false, "Read block JSON from stdin")
	registerStylingFlags(blocksUpdateCmd, false)

	blocksCmd.AddCommand(blocksDeleteCmd)

	blocksCmd.AddCommand(blocksMoveCmd)
	blocksMoveCmd.Flags().StringVar(&blockTargetPage, "to", "", "Target page ID")
	blocksMoveCmd.Flags().StringVarP(&blockPosition, "position", "p", "end", "Position: start, end")
	blocksMoveCmd.MarkFlagRequired("to")

	blocksCmd.AddCommand(blocksRevertCmd)

	blocksCmd.AddCommand(blocksImageCmd)
	blocksImageCmd.Flags().StringVar(&blockImageOut, "out", "", "Write image bytes to FILE (prints JSON envelope to stdout)")
	blocksImageCmd.Flags().StringVar(&blockImageFormat, "image-format", "", "Preferred image format: png, jpeg, webp, gif (default: image/*)")
}
