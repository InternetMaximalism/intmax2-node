package backup_balance

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_balance.go -package=mocks -source=post_backup_balance.go

type UCPostBackupBalance struct {
	Message string `json:"message"`
}

const NUM_TRANSFERS_IN_TX uint = 64
const INSUFFICIENT_FLAGS_LEN uint = NUM_TRANSFERS_IN_TX / 32

type InsufficientFlags struct {
	Limbs [INSUFFICIENT_FLAGS_LEN]uint32
}

type PublicState struct {
	BlockTreeRoot       goldenposeidon.PoseidonHashOut
	PrevAccountTreeRoot goldenposeidon.PoseidonHashOut
	AccountTreeRoot     goldenposeidon.PoseidonHashOut
	DepositTreeRoot     [32]byte
	BlockHash           [32]byte
	BlockNumber         uint32
}

type BalancePublicInputs struct {
	PublicKey               *big.Int                       `json:"pubkey"`
	PrivateCommitment       goldenposeidon.PoseidonHashOut `json:"private_commitment"`
	LastTxHash              goldenposeidon.PoseidonHashOut `json:"last_tx_hash"`
	LastTxInsufficientFlags InsufficientFlags              `json:"last_tx_insufficient_flags"`
	PublicState             PublicState                    `json:"public_state"`
}

type EncryptedPlonky2Proof struct {
	Proof                 string `json:"proof"`
	EncryptedPublicInputs string `json:"publicInputs"`
}

type UCPostBackupBalanceInput struct {
	User                  string                `json:"user"`
	DecodeUser            *intMaxAcc.PublicKey  `json:"-"`
	BlockNumber           uint32                `json:"blockNumber"`
	EncryptedBalanceProof EncryptedPlonky2Proof `json:"encryptedBalanceProof"`
	EncryptedBalanceData  string                `json:"encryptedBalanceData"`
	EncryptedTxs          []string              `json:"encryptedTxs"`
	EncryptedTransfers    []string              `json:"encryptedTransfers"`
	EncryptedDeposits     []string              `json:"encryptedDeposits"`
	Signature             string                `json:"signature"`
}

// UseCasePostBackupBalance describes PostBackupBalance contract.
type UseCasePostBackupBalance interface {
	Do(ctx context.Context, input *UCPostBackupBalanceInput) (*UCPostBackupBalance, error)
}

func MakeMessage(
	user intMaxAcc.Address,
	blockNumber uint32,
	balanceProof []byte,
	encryptedBalancePublicInputs []byte,
	encryptedBalanceData []byte,
	encryptedTxs [][]byte,
	encryptedTransfers [][]byte,
	encryptedDeposits [][]byte,
) []ffg.Element {
	const numAddressBytes = 32
	buf := finite_field.NewBuffer(make([]ffg.Element, 0))
	finite_field.WriteFixedSizeBytes(buf, user.Bytes(), numAddressBytes)
	err := finite_field.WriteUint64(buf, uint64(blockNumber))
	// blockNumber is uint32, so it should be safe to cast to uint64
	if err != nil {
		panic(err)
	}
	finite_field.WriteBytes(buf, balanceProof)
	finite_field.WriteBytes(buf, encryptedBalancePublicInputs)
	finite_field.WriteBytes(buf, encryptedBalanceData)

	err = finite_field.WriteUint64(buf, uint64(len(encryptedTxs)))
	if err != nil {
		panic(err)
	}
	for _, tx := range encryptedTxs {
		finite_field.WriteBytes(buf, tx)
	}
	err = finite_field.WriteUint64(buf, uint64(len(encryptedTransfers)))
	if err != nil {
		panic(err)
	}
	for _, transfer := range encryptedTransfers {
		finite_field.WriteBytes(buf, transfer)
	}
	err = finite_field.WriteUint64(buf, uint64(len(encryptedDeposits)))
	if err != nil {
		panic(err)
	}
	for _, deposit := range encryptedDeposits {
		finite_field.WriteBytes(buf, deposit)
	}

	return buf.Inner()
}
