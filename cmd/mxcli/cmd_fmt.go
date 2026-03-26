// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/mendixlabs/mxcli/mdl/formatter"
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt <file.mdl>",
	Short: "Format an MDL file",
	Long: `Format an MDL script file with consistent styling:
  - Uppercase MDL keywords
  - Normalize indentation (2-space units)
  - Remove trailing whitespace
  - Normalize blank lines

Examples:
  # Format to stdout
  mxcli fmt script.mdl

  # Format in-place
  mxcli fmt script.mdl -w
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		writeInPlace, _ := cmd.Flags().GetBool("write")

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		formatted := formatter.Format(string(data))

		if writeInPlace {
			if err := os.WriteFile(filePath, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Formatted %s\n", filePath)
		} else {
			fmt.Print(formatted)
		}

		return nil
	},
}

func init() {
	fmtCmd.Flags().BoolP("write", "w", false, "Write result to source file instead of stdout")
}
