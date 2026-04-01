package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CommandSchema describes a CLI command for machine consumption
type CommandSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Usage       string          `json:"usage"`
	Flags       []FlagSchema    `json:"flags,omitempty"`
	Subcommands []CommandSchema `json:"subcommands,omitempty"`
	Examples    []string        `json:"examples,omitempty"`
	Safety      *SafetyInfo     `json:"safety,omitempty"`
}

// FlagSchema describes a command flag
type FlagSchema struct {
	Name     string `json:"name"`
	Short    string `json:"short,omitempty"`
	Type     string `json:"type"`
	Default  string `json:"default,omitempty"`
	Required bool   `json:"required"`
	Desc     string `json:"description"`
}

// SafetyInfo describes the safety characteristics of a command
type SafetyInfo struct {
	ReadOnly    bool `json:"readonly"`
	Destructive bool `json:"destructive"`
	Idempotent  bool `json:"idempotent"`
	DryRun      bool `json:"supports_dry_run"`
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Output machine-readable CLI schema as JSON",
	Long: `Output a structured JSON manifest of all craft-cli commands,
flags, types, and safety metadata. Designed for AI agent introspection.

An agent can call 'craft schema' once to discover all available
commands without parsing --help text.

Examples:
  craft schema                    # Full schema as JSON
  craft schema --command list     # Schema for a specific command
  craft schema --commands-only    # Just command names and descriptions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		commandFilter, _ := cmd.Flags().GetString("command")
		commandsOnly, _ := cmd.Flags().GetBool("commands-only")

		schema := buildSchema(rootCmd)

		if commandFilter != "" {
			for _, sc := range schema.Subcommands {
				if sc.Name == commandFilter {
					return outputSchemaJSON(sc)
				}
			}
			return fmt.Errorf("unknown command: %s", commandFilter)
		}

		if commandsOnly {
			type briefCmd struct {
				Name string `json:"name"`
				Desc string `json:"description"`
			}
			var cmds []briefCmd
			for _, sc := range schema.Subcommands {
				cmds = append(cmds, briefCmd{Name: sc.Name, Desc: sc.Description})
			}
			return outputSchemaJSON(cmds)
		}

		return outputSchemaJSON(schema)
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
	schemaCmd.Flags().String("command", "", "Show schema for a specific command only")
	schemaCmd.Flags().Bool("commands-only", false, "Output only command names and descriptions")
}

func buildSchema(cmd *cobra.Command) CommandSchema {
	schema := CommandSchema{
		Name:        cmd.Name(),
		Description: cmd.Short,
		Usage:       cmd.UseLine(),
	}

	// Collect examples
	if cmd.Example != "" {
		schema.Examples = schemaParseExamples(cmd.Example)
	}

	// Collect local flags
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		fs := FlagSchema{
			Name: "--" + f.Name,
			Type: f.Value.Type(),
			Desc: f.Usage,
		}
		if f.Shorthand != "" {
			fs.Short = "-" + f.Shorthand
		}
		if f.DefValue != "" && f.DefValue != "false" {
			fs.Default = f.DefValue
		}
		schema.Flags = append(schema.Flags, fs)
	})

	// Add safety metadata based on command name
	schema.Safety = inferSafety(cmd.Name())

	// Collect subcommands
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && sub.Name() != "help" && sub.Name() != "schema" {
			subSchema := buildSchema(sub)
			schema.Subcommands = append(schema.Subcommands, subSchema)
		}
	}

	return schema
}

func inferSafety(name string) *SafetyInfo {
	switch name {
	case "list", "get", "search", "info", "connection", "version", "folders", "tasks", "collections", "llm", "schema":
		return &SafetyInfo{ReadOnly: true, Destructive: false, Idempotent: true, DryRun: false}
	case "create":
		return &SafetyInfo{ReadOnly: false, Destructive: false, Idempotent: false, DryRun: true}
	case "update", "move":
		return &SafetyInfo{ReadOnly: false, Destructive: false, Idempotent: true, DryRun: true}
	case "delete", "clear":
		return &SafetyInfo{ReadOnly: false, Destructive: true, Idempotent: true, DryRun: true}
	default:
		return &SafetyInfo{ReadOnly: false, Destructive: false, Idempotent: false, DryRun: true}
	}
}

func schemaParseExamples(examples string) []string {
	var result []string
	for _, line := range strings.Split(examples, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func outputSchemaJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
