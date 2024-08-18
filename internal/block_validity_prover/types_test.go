package block_validity_prover_test

import (
	"encoding/json"
	"fmt"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/block_validity_prover"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestBlockHash(t *testing.T) {
	prevBlockHash := common.Hash{}
	depositRoot := common.HexToHash("0xb6155ab566bbd2e341525fd88c43b4d69572bf4afe7df45cd74d6901a172e41c")
	signatureHash := common.Hash{}
	postedBlock := block_post_service.PostedBlock{
		PrevBlockHash: prevBlockHash,
		BlockNumber:   0,
		DepositRoot:   depositRoot,
		SignatureHash: signatureHash,
	}

	currentHash := postedBlock.Hash()
	require.Equal(t, currentHash.String(), "0x545cac70c52cf8589c16de1eb85e264d51e18adb15ac810db3f44efa190a1074")
}

func TestPostBackupBalance(t *testing.T) {
	publicState := new(block_validity_prover.PublicState).Genesis()
	s, err := json.Marshal(publicState)
	require.NoError(t, err)
	fmt.Printf("publicState: %s\n", s)
}

func TestBlockWitness(t *testing.T) {
	blockWitness := new(block_validity_prover.BlockWitness).Genesis()
	s, err := json.Marshal(blockWitness)
	require.NoError(t, err)
	fmt.Printf("blockWitness: %s\n", s)
}

func TestValidityTransitionWitness(t *testing.T) {
	validityTransitionWitness := new(block_validity_prover.ValidityTransitionWitness).Genesis()
	s, err := json.Marshal(validityTransitionWitness)
	require.NoError(t, err)
	fmt.Printf("validityTransitionWitness: %s\n", s)
}

func TestValidityWitness(t *testing.T) {
	validityWitness := new(block_validity_prover.ValidityWitness).Genesis()
	s, err := json.Marshal(validityWitness)
	require.NoError(t, err)
	fmt.Printf("validityWitness: %s\n", s)
}

func TestAccountIdPacked(t *testing.T) {
	accountIDs := make([]uint64, 128)
	for i := 0; i < 10; i++ {
		accountIDs[i] = uint64(i) * 100
	}

	packed := new(block_validity_prover.AccountIdPacked).Pack(accountIDs)
	unpacked := packed.Unpack()
	require.Equal(t, accountIDs, unpacked)
}
