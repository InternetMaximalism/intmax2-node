package block_validity_prover_test

import (
	"intmax2-node/internal/block_validity_prover"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestBlockHash(t *testing.T) {
	prevBlockHash := common.Hash{}
	depositRoot := common.HexToHash("0xb6155ab566bbd2e341525fd88c43b4d69572bf4afe7df45cd74d6901a172e41c")
	signatureHash := common.Hash{}
	postedBlock := block_validity_prover.PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   0,
		DepositRoot:   depositRoot,
		SignatureHash: signatureHash,
	}

	currentHash := postedBlock.Hash()
	require.Equal(t, currentHash.String(), "0x545cac70c52cf8589c16de1eb85e264d51e18adb15ac810db3f44efa190a1074")
}
