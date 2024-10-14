// nolint:dupl
package transaction

import (
	"fmt"
	txWithdrawalTransfersList "intmax2-node/internal/use_cases/tx_withdrawal_transfers_list"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txWithdrawalTransfersListCmd(b *Transaction) *cobra.Command {
	const (
		use   = "list"
		short = "Get transfers list"

		emptyKey                         = ""
		userPrivateKeyKey                = "private-key"
		userPrivateDesc                  = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		filterNameKey                    = "filterName"
		filterNameDesc                   = "specify the filter name. use as --filterName \"block_number\" (support value: \"block_number\", \"start_backup_time\")"
		filterConditionKey               = "filterCondition"
		filterConditionDesc              = "specify the filter condition. use as --filterCondition \"is\" (support values: \"lessThan\", \"lessThanOrEqualTo\", \"is\" (only for \"block_number\"), \"greaterThanOrEqualTo\", \"greaterThan\")" // nolint:lll
		filterValueKey                   = "filterValue"
		filterValueDesc                  = "specify the value of filter. use as --filterValue \"1\" (examples: for \"block_number\" = \"1\"; for \"start_backup_time\" = \"2024-06-10T22:00:00.123Z\")"
		sortingKey                       = "sorting"
		sortingDesc                      = "specify the sorting. use as --sorting \"desc\" (support values: \"asc\", \"desc\")"
		sortingDef                       = "desc"
		paginationDirectionKey           = "paginationDirection"
		paginationDirectionDesc          = "specify the direction pagination. use as --paginationDirection \"next\" (support values: \"next\", \"prev\")"
		paginationDirectionDef           = "next"
		paginationLimitKey               = "paginationLimit"
		paginationLimitDesc              = "specify the limit for pagination without decimals. use as --paginationLimit \"100\""
		paginationLimitDef               = "100"
		paginationCursorBlockNumberKey   = "paginationCursorBlockNumber"
		paginationCursorBlockNumberDesc  = "specify the BlockNumber cursor. use as --paginationCursorBlockNumber \"1\" (more then \"0\")"
		paginationCursorSortingValueKey  = "paginationCursorSortingValue"
		paginationCursorSortingValueDesc = "specify the SortingValue cursor. use as --paginationCursorSortingValue \"1\" (more then \"0\")"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var filterName string
	cmd.PersistentFlags().StringVar(&filterName, filterNameKey, emptyKey, filterNameDesc)

	var filterCondition string
	cmd.PersistentFlags().StringVar(&filterCondition, filterConditionKey, emptyKey, filterConditionDesc)

	var filterValue string
	cmd.PersistentFlags().StringVar(&filterValue, filterValueKey, emptyKey, filterValueDesc)

	var sorting string
	cmd.PersistentFlags().StringVar(&sorting, sortingKey, sortingDef, sortingDesc)

	var paginationDirection string
	cmd.PersistentFlags().StringVar(&paginationDirection, paginationDirectionKey, paginationDirectionDef, paginationDirectionDesc)

	var paginationLimit string
	cmd.PersistentFlags().StringVar(&paginationLimit, paginationLimitKey, paginationLimitDef, paginationLimitDesc)

	var paginationCursorBlockNumber string
	cmd.PersistentFlags().StringVar(&paginationCursorBlockNumber, paginationCursorBlockNumberKey, emptyKey, paginationCursorBlockNumberDesc)

	var paginationCursorSortingValue string
	cmd.PersistentFlags().StringVar(&paginationCursorSortingValue, paginationCursorSortingValueKey, emptyKey, paginationCursorSortingValueDesc)

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDesc)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		resp, err := newCommands().WithdrawalTransfersList(
			b.Config, b.Log, b.SB,
		).Do(
			b.Context, &txWithdrawalTransfersList.UCTxWithdrawalTransfersListInput{
				Sorting: sorting,
				Pagination: &txWithdrawalTransfersList.UCTxWithdrawalTransfersListPagination{
					Direction: paginationDirection,
					Limit:     paginationLimit,
					Cursor: &txWithdrawalTransfersList.UCTxWithdrawalTransfersListPaginationCursor{
						BlockNumber:  paginationCursorBlockNumber,
						SortingValue: paginationCursorSortingValue,
					},
				},
				Filter: &txWithdrawalTransfersList.UCTxWithdrawalTransfersListFilter{
					Name:      filterName,
					Condition: filterCondition,
					Value:     filterValue,
				},
			}, utils.RemoveZeroX(userEthPrivateKey),
		)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString(string(resp))
	}

	return &cmd
}
