package transaction_test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"intmax2-node/configs"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	"intmax2-node/internal/pow"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/internal/use_cases/transaction"
	ucTransaction "intmax2-node/pkg/use_cases/transaction"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type SampleTxRequest struct {
	desc    string
	input   *transaction.UCTransactionInput
	prepare func()
	err     error
}

const txRequestCase1 = `{
    "sender":"0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3",
    "transfersHash":"0xe35508e23eed79e9f9c1c446c6429a3cb1a43aa86edac916f5790b8bfce468b7",
    "nonce":1,
    "powNonce":"0x244c",
    "expiration":"2024-07-11T16:31:28.651829+02:00",
    "signature":"0x27fbfe686763ba4c215741dccc4e4500a0d9297291d20cf0b9fa404f7470a4c1096d6f35c01187e2d5a70fa89a9d7e5a5cbdb2eb508cbf629bee5dce736a0cfc0cdc2282b36ca35ae50f345d0240e38c84a10174c5605c8d84de90041d28f6fa2d7c264d813eff2f8b8ff3d4ac93c8ec82c2070387403f9de4a73fc5b4c5302d"
}`

const txRequestCase2 = `{
    "sender":"0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3",
    "transfersHash":"0xe35508e23eed79e9f9c1c446c6429a3cb1a43aa86edac916f5790b8bfce468b7",
    "nonce":1,
	"powNonce":"0x244c",
    "expiration":"2024-07-11T16:31:28.651829+02:00",
	"signature":"0x056f6ff941a9ee7fc7acbd725e423d632652271a26bde5c668994125770b67f22fdb0bb5cf0154eae3937d784d2c6ff8b4c8a44dd3ec074e34c0a074a609f2142e40317f19ef2982bb776859c6d4553971f5955f0d8de41fb3402c37b99395f40d01eace7a16650fef3d8c96aba5c0e474a544261cf1301a655b76d9e647645b"
}`

const txRequestCase3 = `{
    "sender":"0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3",
    "transfersHash":"0xe35508e23eed79e9f9c1c446c6429a3cb1a43aa86edac916f5790b8bfce468b7",
    "nonce":1,
    "powNonce":"0xc1",
    "expiration":"2024-07-11T16:31:28.651829+02:00",
    "signature":"0x27fbfe686763ba4c215741dccc4e4500a0d9297291d20cf0b9fa404f7470a4c1096d6f35c01187e2d5a70fa89a9d7e5a5cbdb2eb508cbf629bee5dce736a0cfc0cdc2282b36ca35ae50f345d0240e38c84a10174c5605c8d84de90041d28f6fa2d7c264d813eff2f8b8ff3d4ac93c8ec82c2070387403f9de4a73fc5b4c5302d"
}`

func TestDecodeTransactionRequest(t *testing.T) {
	transactionRequest := &transaction.UCTransactionInput{}
	err := json.Unmarshal([]byte(txRequestCase2), transactionRequest)
	assert.NoError(t, err)
}

func TestUseCaseTransaction(t *testing.T) {
	const int3Key = 3
	assert.NoError(t, configs.LoadDotEnv(int3Key))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := configs.New()
	dbApp := NewMockSQLDriverApp(ctrl)
	worker := NewMockWorker(ctrl)

	uc := ucTransaction.New(cfg, dbApp, worker)

	cases := make([]SampleTxRequest, 0)
	case1 := SampleTxRequest{
		desc:  "Success",
		input: new(transaction.UCTransactionInput),
	}
	err := json.Unmarshal([]byte(txRequestCase1), &case1.input)
	assert.NoError(t, err)
	decodedSender, err := intMaxAcc.NewPublicKeyFromAddressHex(case1.input.Sender)
	assert.NoError(t, err)
	case1.input.DecodeSender = new(intMaxAcc.PublicKey).Set(decodedSender)
	cases = append(cases, case1)

	case2 := SampleTxRequest{
		desc:  "Invalid Signature",
		input: new(transaction.UCTransactionInput),
		err:   errors.New("signature is invalid"),
	}
	err = json.Unmarshal([]byte(txRequestCase2), &case2.input)
	assert.NoError(t, err)
	decodedSender, err = intMaxAcc.NewPublicKeyFromAddressHex(case2.input.Sender)
	assert.NoError(t, err)
	case2.input.DecodeSender = new(intMaxAcc.PublicKey).Set(decodedSender)
	cases = append(cases, case2)

	case3 := SampleTxRequest{
		desc:  "Invalid PoW nonce",
		input: new(transaction.UCTransactionInput),
		err:   errors.New("PoW nonce is invalid"),
	}
	err = json.Unmarshal([]byte(txRequestCase3), &case3.input)
	assert.NoError(t, err)
	decodedSender, err = intMaxAcc.NewPublicKeyFromAddressHex(case3.input.Sender)
	assert.NoError(t, err)
	case3.input.DecodeSender = new(intMaxAcc.PublicKey).Set(decodedSender)
	cases = append(cases, case3)

	transferTree, err := intMaxTree.NewTransferTree(7, nil, new(intMaxTypes.Transfer).SetZero().Hash())
	assert.NoError(t, err)
	transferTreeRoot, _, _ := transferTree.GetCurrentRootCountAndSiblings()

	zeroHash := new(intMaxTypes.PoseidonHashOut).SetZero()
	txTree, err := intMaxTree.NewTxTree(7, []*intMaxTypes.Tx{}, zeroHash)
	assert.NoError(t, err)

	for i := range cases {
		tx, err := intMaxTypes.NewTx(
			&transferTreeRoot,
			cases[i].input.Nonce,
		)
		assert.NoError(t, err)
		txTree.AddLeaf(uint64(i), tx)
	}

	// NOTE: The Block Builder stores the Merkle proofs obtained here in storage
	// and passes the Merkle proof for the user's transaction when
	// a `GET /block/proposed` request is received from the user.
	for i := range cases {
		_, _, err := txTree.ComputeMerkleProof(uint64(i))
		assert.NoError(t, err)
	}

	for i := range cases {
		worker.EXPECT().Receiver(gomock.Any()).Times(1)
		t.Run(cases[i].desc, func(t *testing.T) {
			if cases[i].prepare != nil {
				cases[i].prepare()
			}

			ctx := context.Background()
			if cases[i].err != nil {
				assert.True(t, errors.Is(uc.Do(ctx, cases[i].input), cases[i].err))
			} else {
				assert.NoError(t, uc.Do(ctx, cases[i].input))
			}
		})
	}
}

func MakeSampleTxRequest(cfg *configs.Config) *transaction.UCTransactionInput {
	senderAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(2))
	if err != nil {
		panic(err)
	}
	sender := senderAccount.ToAddress().String()

	recipientAccount, err := intMaxAcc.NewPrivateKey(big.NewInt(4))
	if err != nil {
		panic(err)
	}
	recipientAddress, err := intMaxTypes.NewINTMAXAddress(recipientAccount.ToAddress().Bytes())
	if err != nil {
		panic(err)
	}

	salt := new(intMaxTypes.PoseidonHashOut).SetZero()
	zeroHash := new(intMaxTypes.PoseidonHashOut).SetZero()
	transfer := intMaxTypes.Transfer{
		Recipient:  recipientAddress,
		TokenIndex: 1,
		Amount:     big.NewInt(100),
		Salt:       salt,
	}
	transfers := make([]*intMaxTypes.Transfer, 8)
	transfers[0] = &transfer
	transferTree, err := intMaxTree.NewTransferTree(
		7,
		nil,
		zeroHash,
	)
	if err != nil {
		panic(err)
	}
	transfersHash, _, _ := transferTree.GetCurrentRootCountAndSiblings()
	var nonce uint64 = 1

	expiration := time.Now()

	ctx := context.Background()
	pw := pow.New(cfg.PoW.Difficulty)
	pWorker := pow.NewWorker(cfg.PoW.Workers, pw)
	powNonce, err := pow.NewPoWNonce(pw, pWorker).Nonce(ctx, transfersHash.Marshal())
	if err != nil {
		panic(err)
	}

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	message := crypto.Keccak256(
		transfersHash.Marshal(), nonceBytes, []byte(powNonce), []byte(sender), []byte(expiration.Format(time.RFC3339)),
	)

	signature, err := senderAccount.Sign(finite_field.BytesToFieldElementSlice(message))
	if err != nil {
		panic(err)
	}

	return &transaction.UCTransactionInput{
		Sender:        sender,
		DecodeSender:  senderAccount.Public(),
		TransfersHash: transfersHash.String(),
		Nonce:         nonce,
		PowNonce:      powNonce,
		Expiration:    expiration,
		Signature:     hexutil.Encode(signature.Marshal()),
	}
}
