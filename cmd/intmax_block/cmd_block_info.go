package intmax_block

import (
	"context"
	"encoding/json"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

func infoCmd(
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
) *cobra.Command {
	const (
		use   = "info"
		short = "Returns the INTMAX block info"

		emptyKey        = ""
		int0Key         = 0
		blockHashKey    = "hash"
		blockHashDesc   = "the block hash value. use as --hash \"0xa8f3e5ac6ed846a9429afcc94368eb7c49f1103d84110bd6c9b65d58b8e6574a\""
		blockNumberKey  = "number"
		blockNumberDesc = "the block number value. use as --number \"1\" (default is zero and value must more then zero)"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var blockHash string
	cmd.PersistentFlags().StringVar(&blockHash, blockHashKey, emptyKey, blockHashDesc)

	var blockNumber int64
	cmd.PersistentFlags().Int64Var(&blockNumber, blockNumberKey, int0Key, blockNumberDesc)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if blockNumber < int0Key {
			const msg = "Fatal: number must more then zero\n"
			_, _ = fmt.Fprintf(os.Stderr, msg)
			os.Exit(1)
		} else if blockNumber == int0Key && strings.EqualFold(strings.TrimSpace(blockHash), emptyKey) {
			const msg = "Fatal: hash or number must setup\n"
			_, _ = fmt.Fprintf(os.Stderr, msg)
			os.Exit(1)
		}

		resp, err := newCommands().GetINTMAXBlockInfo(cfg, log).Do(ctx, args, blockHash, uint64(blockNumber))
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}

		const (
			successKey         = "success"
			dataBlockNumberKey = "data.blockNumber"
			blockNotFound      = `{
  "success": false,
  "error": {
    "code": 404,
    "message": "block not found."
  }
}`
		)

		if gjson.GetBytes(resp, successKey).Bool() {
			if blockNumber > int0Key && !strings.EqualFold(strings.TrimSpace(blockHash), emptyKey) {
				if gjson.GetBytes(resp, dataBlockNumberKey).Int() != blockNumber {
					resp = json.RawMessage(blockNotFound)
				}
			}
		}

		_, _ = os.Stdout.WriteString(string(resp))
	}

	return &cmd
}
