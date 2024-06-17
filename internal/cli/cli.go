package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

func Run(ctx context.Context, cmd ...*cobra.Command) error {
	const app = "app"
	rootCmd := &cobra.Command{Use: app}

	// switch off usage message on run without args
	rootCmd.Run = func(cmd *cobra.Command, args []string) {}

	// add exit on help
	helpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		helpFunc(c, s)
		const code = -1
		os.Exit(code)
	})

	// add commands
	rootCmd.AddCommand(cmd...)

	return rootCmd.ExecuteContext(ctx)
}
