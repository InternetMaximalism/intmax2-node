package transaction

import (
	"fmt"
	txDepositListIncoming "intmax2-node/internal/use_cases/tx_deposits_list_incoming"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txDepositListIncomingCmd(b *Transaction) *cobra.Command {
	const (
		use   = "incoming"
		short = "Get deposit list (incoming)"

		emptyKey                         = ""
		userPrivateKeyKey                = "private-key"
		userPrivateDesc                  = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		filterNameKey                    = "filterName"
		filterNameDesc                   = "specify the filter name. use as --filterName \"block_number\" (support value: \"block_number\")"
		filterConditionKey               = "filterCondition"
		filterConditionDesc              = "specify the filter condition. use as --filterCondition \"is\" (support values: \"lessThan\", \"lessThanOrEqualTo\", \"is\", \"greaterThanOrEqualTo\", \"greaterThan\")" // nolint:lll
		filterValueKey                   = "filterValue"
		filterValueDesc                  = "specify the value of filter. use as --filterValue \"1\""
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
		resp, err := newCommands().ReceiverDepositsListIncoming(
			b.Config, b.Log, b.SB,
		).Do(
			b.Context, &txDepositListIncoming.UCTxDepositsListIncomingInput{
				Sorting: sorting,
				Pagination: &txDepositListIncoming.UCTxDepositsListIncomingPagination{
					Direction: paginationDirection,
					Limit:     paginationLimit,
					Cursor: &txDepositListIncoming.UCTxDepositsListIncomingPaginationCursor{
						BlockNumber:  paginationCursorBlockNumber,
						SortingValue: paginationCursorSortingValue,
					},
				},
				Filter: &txDepositListIncoming.UCTxDepositsListIncomingFilter{
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
