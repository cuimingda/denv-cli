package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
	return NewInstallCmdWithService(NewCLIContext().Service)
}

func NewInstallCmdWithService(svc CommandService) *cobra.Command {
	if svc == nil {
		svc = NewCLIContext().Service
	}

	longHelp := `Install all supported developer tools.
Supported tools:
- php  -> brew install php
- python3 -> brew install python3
- node -> brew install node
- go -> brew install go
- curl -> brew install curl
- gh -> brew install gh
- git -> brew install git
- ffmpeg -> brew install ffmpeg
- tree -> brew install tree`

	cmd := &cobra.Command{
		Use:     "install",
		Args:    cobra.NoArgs,
		Short:   "Install supported developer tools",
		Long:    longHelp,
		Example: "  denv install",
		RunE: func(cmd *cobra.Command, _ []string) error {
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			operations, err := svc.BuildInstallOperations(force)
			if err != nil {
				return err
			}

			if dryRun {
				for _, operation := range operations {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Would run: %s\n", operation.String()); err != nil {
						return err
					}
				}
				return nil
			}

			for _, operation := range operations {
				if err := svc.RunInstallOperation(cmd.OutOrStdout(), operation); err != nil {
					return err
				}
			}

			_, outErr := fmt.Fprintln(cmd.OutOrStdout(), "install done")
			return outErr
		},
	}

	cmd.Flags().Bool("force", false, "install even if the tool already exists")
	cmd.Flags().Bool("dry-run", false, "show planned install operations only")
	return cmd
}
