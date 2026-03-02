package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func isVerbose(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}

	if flag := cmd.Flags().Lookup("verbose"); flag != nil {
		if value, err := cmd.Flags().GetBool("verbose"); err == nil {
			return value
		}
	}

	root := cmd.Root()
	if root == nil {
		return false
	}

	if flag := root.PersistentFlags().Lookup("verbose"); flag != nil {
		if value, err := root.PersistentFlags().GetBool("verbose"); err == nil {
			return value
		}
	}

	if flag := root.Flags().Lookup("verbose"); flag != nil {
		if value, err := root.Flags().GetBool("verbose"); err == nil {
			return value
		}
	}

	return false
}

func verbosef(cmd *cobra.Command, format string, args ...any) {
	if !isVerbose(cmd) {
		return
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[verbose] "+format+"\n", args...)
}

func doingf(cmd *cobra.Command, format string, args ...any) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[doing] "+format+"\n", args...)
}
