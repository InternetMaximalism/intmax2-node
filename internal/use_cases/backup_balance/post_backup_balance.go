package backup_balance

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/block_signature"
	"math/big"

	"github.com/iden3/go-iden3-crypto/ffg"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_balance.go -package=mocks -source=post_backup_balance.go

const (
	numTransfersInTx     = 1 << intMaxTree.TRANSFER_TREE_HEIGHT
	insufficientFlagsLen = numTransfersInTx / 32
	uint256LimbSize      = 8
	int32Key             = 32
)

type UCPostBackupBalance struct {
	Message string `json:"message"`
}

type InsufficientFlags struct {
	Limbs [insufficientFlagsLen]uint32
}

func (flags *InsufficientFlags) FromFieldElementSlice(value []ffg.Element) *InsufficientFlags {
	for i, x := range value {
		y := x.ToUint64Regular()
		if y >= uint64(1)<<int32Key {
			panic("overflow")
		}
		flags.Limbs[i] = uint32(y)
	}

	return flags
}

type BalancePublicInputs struct {
	PublicKey               *big.Int                           `json:"pubkey"`
	PrivateCommitment       goldenposeidon.PoseidonHashOut     `json:"privateCommitment"`
	LastTxHash              goldenposeidon.PoseidonHashOut     `json:"lastTxHash"`
	LastTxInsufficientFlags InsufficientFlags                  `json:"lastTxInsufficientFlags"`
	PublicState             *block_validity_prover.PublicState `json:"publicState"`
}

func (pis *BalancePublicInputs) Equal(other *BalancePublicInputs) bool {
	if pis.PublicKey.Cmp(other.PublicKey) != 0 {
		return false
	}
	if !pis.PrivateCommitment.Equal(&other.PrivateCommitment) {
		return false
	}
	if !pis.LastTxHash.Equal(&other.LastTxHash) {
		return false
	}
	if pis.LastTxInsufficientFlags != other.LastTxInsufficientFlags {
		return false
	}
	if !pis.PublicState.Equal(other.PublicState) {
		return false
	}
	return true
}

func VerifyEnoughBalanceProof(enoughBalanceProof *block_signature.Plonky2Proof) (*BalancePublicInputs, error) {
	publicInputs := make([]ffg.Element, len(enoughBalanceProof.PublicInputs))
	for i, publicInput := range enoughBalanceProof.PublicInputs {
		publicInputs[i].SetUint64(publicInput)
	}
	decodedPublicInputs := new(BalancePublicInputs).FromPublicInputs(publicInputs)
	err := decodedPublicInputs.Verify()
	if err != nil {
		return nil, err
	}

	// TODO: Verify verifier data in public inputs.

	// TODO: Verify enough balance proof by using Balance Validity Prover.
	return decodedPublicInputs, nil
}

func (pis *BalancePublicInputs) FromPublicInputs(publicInputs []ffg.Element) *BalancePublicInputs {
	const startPrivateCommitmentIndex = uint256LimbSize
	const lastTxHashIndex = startPrivateCommitmentIndex + goldenposeidon.NUM_HASH_OUT_ELTS
	const lastTxInsufficientFlagsIndex = lastTxHashIndex + goldenposeidon.NUM_HASH_OUT_ELTS
	const publicStateIndex = lastTxInsufficientFlagsIndex + insufficientFlagsLen
	const endIndex = publicStateIndex + block_validity_prover.PublicStateLimbSize
	if len(publicInputs) != endIndex {
		panic("Invalid public inputs length")
	}

	publicKey := new(intMaxTypes.Uint256).FromFieldElementSlice(publicInputs[0:startPrivateCommitmentIndex])
	privateCommitment := new(goldenposeidon.PoseidonHashOut)
	copy(privateCommitment.Elements[:], publicInputs[startPrivateCommitmentIndex:lastTxHashIndex])
	lastTxHash := new(goldenposeidon.PoseidonHashOut)
	copy(lastTxHash.Elements[:], publicInputs[lastTxHashIndex:lastTxInsufficientFlagsIndex])
	lastTxInsufficientFlags := new(InsufficientFlags).FromFieldElementSlice(publicInputs[lastTxInsufficientFlagsIndex:publicStateIndex])
	publicState := new(block_validity_prover.PublicState).FromFieldElementSlice(publicInputs[publicStateIndex:endIndex])

	return &BalancePublicInputs{
		PublicKey:               publicKey.BigInt(),
		PrivateCommitment:       *privateCommitment,
		LastTxHash:              *lastTxHash,
		LastTxInsufficientFlags: *lastTxInsufficientFlags,
		PublicState:             publicState,
	}
}

func (pis *BalancePublicInputs) Verify() error {
	return nil
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
