package migrator

import (
	"context"
	"intmax2-node/internal/logger"
	"os"

	"github.com/spf13/cobra"
)

func NewMigratorCmd(ctx context.Context, log logger.Logger, db SQLDriverApp) *cobra.Command {
	const (
		use   = "migrate --action \"<up|down|1|-1>\""
		short = "Execute migration"
		long  = "Execute migrations stored at binary\n" +
			"Actions:\n" +
			"up - migrate all steps Up\n" +
			"down - migrate all steps Down\n" +
			"number - amount of steps to migrate (if > 0 - migrate number steps up, if < 0 migrate number steps down)"
		emptyKey          = ""
		actionKey         = "action"
		actionDescription = "action flag. use as --action \"<up|down|1|-1>\""
		maskMigrate       = "migrate: %+v"
		maskStepDown      = "%d steps done"
		code0             = 0
		module            = "module"
		migrate           = "migrate"
	)
	cmd := cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	var action string
	cmd.PersistentFlags().StringVar(&action, actionKey, emptyKey, actionDescription)

	cmd.Run = func(_ *cobra.Command, args []string) {
		log = log.WithFields(logger.Fields{module: migrate})
		if action == emptyKey {
			const msg = "action flag is not set"
			log.Fatalf(maskMigrate, msg)
		}

		n, err := db.Migrator(ctx, action)
		if err != nil {
			log.Fatalf(maskMigrate, err)
		}
		log.Infof(maskStepDown, n)
		os.Exit(code0)
	}

	return &cmd
}
