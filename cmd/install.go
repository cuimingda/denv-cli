package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/cuimingda/denv-cli/internal"
	"github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewInstallCmdWithService(ctx.InstallContext)
}

type installCommandService struct {
	service denv.InstallContext
}

func (s installCommandService) BuildInstallQueue(force bool) (denv.InstallQueue, error) {
	return s.service.BuildInstallQueue(force)
}

func (s installCommandService) ExecuteInstallQueue(out io.Writer, queue denv.InstallQueue) error {
	return s.service.ExecuteInstallQueue(out, queue)
}

func NewInstallCmdWithService(service denv.InstallContext) *cobra.Command {
	if service == nil {
		panic("install command requires a non-nil install planner")
	}

	longHelp := denv.InstallLongHelp()
	commandService := installCommandService{service: service}

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
			installQueue, err := commandService.BuildInstallQueue(force)
			if err != nil {
				return err
			}
			doingf(cmd, "prepared %d operations in %s", installQueue.Len(), time.Since(operationStart))

			if dryRun {
				doingf(cmd, "showing installation plan without execution")
				for _, operation := range installQueue.ToOperations() {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Would run: %s\n", operation.String()); err != nil {
						return err
					}
				}
				doingf(cmd, "dry-run plan generation completed in %s", time.Since(operationStart))
				return nil
			}

			doingf(cmd, "start executing %d install operations", installQueue.Len())
			if err := commandService.ExecuteInstallQueue(cmd.OutOrStdout(), installQueue); err != nil {
				return err
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
