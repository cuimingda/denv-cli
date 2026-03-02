package cmd

import (
	"fmt"
	"time"

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
			operationStart := time.Now()

			doingf(cmd, "prepare install plan (force=%t, dry-run=%t)", force, dryRun)
			operations, err := svc.BuildInstallOperations(force)
			if err != nil {
				return err
			}
			verbosef(cmd, "planned %d install operations (dry-run=%t)", len(operations), dryRun)
			doingf(cmd, "prepared %d operations in %s", len(operations), time.Since(operationStart))

			if dryRun {
				doingf(cmd, "showing installation plan without execution")
				for _, operation := range operations {
					verbosef(cmd, "dry-run operation: %s", operation.String())
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Would run: %s\n", operation.String()); err != nil {
						return err
					}
				}
				verbosef(cmd, "dry-run plan generation completed in %s", time.Since(operationStart))
				return nil
			}

			doingf(cmd, "start executing %d install operations", len(operations))
			for idx, operation := range operations {
				doingf(cmd, "start: %s", operation.String())
				verbosef(cmd, "executing operation %d/%d: %s", idx+1, len(operations), operation.String())
				start := time.Now()
				if err := svc.RunInstallOperation(cmd.OutOrStdout(), operation); err != nil {
					return err
				}
				verbosef(cmd, "operation %d/%d completed in %s", idx+1, len(operations), time.Since(start))
			}

			verbosef(cmd, "install completed in %s", time.Since(operationStart))
			_, outErr := fmt.Fprintln(cmd.OutOrStdout(), "install done")
			return outErr
		},
	}

	cmd.Flags().Bool("force", false, "install even if the tool already exists")
	cmd.Flags().Bool("dry-run", false, "show planned install operations only")
	return cmd
}
