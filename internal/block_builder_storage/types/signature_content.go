package types

import (
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/finite_field"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/ffg"
)

type SignatureContent struct {
	IsRegistrationBlock bool                `json:"isRegistrationBlock"`
	TxTreeRoot          intMaxTypes.Bytes32 `json:"txTreeRoot"`
	SenderFlag          intMaxTypes.Bytes16 `json:"senderFlag"`
	PublicKeyHash       intMaxTypes.Bytes32 `json:"pubkeyHash"`
	AccountIDHash       intMaxTypes.Bytes32 `json:"accountIdHash"`
	AggPublicKey        intMaxTypes.FlatG1  `json:"aggPubkey"`
	AggSignature        intMaxTypes.FlatG2  `json:"aggSignature"`
	MessagePoint        intMaxTypes.FlatG2  `json:"messagePoint"`
}

func NewSignatureContentFromBlockContent(blockContent *intMaxTypes.BlockContent) *SignatureContent {
	isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType

	publicKeys := make([]intMaxTypes.Uint256, len(blockContent.Senders))
	accountIDs := make([]uint64, len(blockContent.Senders))
	senderFlagBytes := [Int16Key]byte{}
	for i, sender := range blockContent.Senders {
		publicKey := new(intMaxTypes.Uint256).FromBigInt(sender.PublicKey.BigInt())
		publicKeys[i] = *publicKey
		accountIDs[i] = sender.AccountID
		var flag uint8 = 0
		if sender.IsSigned {
			flag = 1
		}
		senderFlagBytes[i/Int8Key] |= flag << (Int8Key - 1 - i%Int8Key)
	}

	signatureContent := SignatureContent{
		IsRegistrationBlock: isRegistrationBlock,
		TxTreeRoot:          intMaxTypes.Bytes32{},
		SenderFlag:          intMaxTypes.Bytes16{},
		PublicKeyHash:       GetPublicKeysHash(publicKeys),
		AccountIDHash:       GetAccountIDsHash(accountIDs),
		AggPublicKey:        intMaxTypes.FlattenG1Affine(blockContent.AggregatedPublicKey.Pk),
		AggSignature:        intMaxTypes.FlattenG2Affine(blockContent.AggregatedSignature),
		MessagePoint:        intMaxTypes.FlattenG2Affine(blockContent.MessagePoint),
	}
	copy(signatureContent.TxTreeRoot[:], intMaxTypes.CommonHashToUint32Slice(blockContent.TxTreeRoot))
	signatureContent.SenderFlag.FromBytes(senderFlagBytes[:])

	return &signatureContent
}

func (sc *SignatureContent) Set(other *SignatureContent) *SignatureContent {
	sc.IsRegistrationBlock = other.IsRegistrationBlock
	sc.TxTreeRoot = other.TxTreeRoot
	sc.SenderFlag = other.SenderFlag
	sc.PublicKeyHash = other.PublicKeyHash
	sc.AccountIDHash = other.AccountIDHash
	sc.AggPublicKey = other.AggPublicKey
	sc.AggSignature = other.AggSignature
	sc.MessagePoint = other.MessagePoint

	return sc
}

func (sc *SignatureContent) ToFieldElementSlice() []ffg.Element {
	buf := finite_field.NewBuffer(make([]ffg.Element, 0))
	var isRegistrationBlock uint32 = 0
	if sc.IsRegistrationBlock {
		isRegistrationBlock = 1
	}
	finite_field.WriteUint32(buf, isRegistrationBlock)
	for _, d := range sc.TxTreeRoot {
		finite_field.WriteUint32(buf, d)
	}
	for _, d := range sc.SenderFlag {
		finite_field.WriteUint32(buf, d)
	}
	for _, d := range sc.PublicKeyHash {
		finite_field.WriteUint32(buf, d)
	}
	for _, d := range sc.AccountIDHash {
		finite_field.WriteUint32(buf, d)
	}
	for i := range sc.AggPublicKey {
		coord := sc.AggPublicKey[i].ToFieldElementSlice()
		for j := range coord {
			finite_field.WriteGoldilocksField(buf, &coord[j])
		}
	}
	for i := range sc.AggSignature {
		coord := sc.AggSignature[i].ToFieldElementSlice()
		for j := range coord {
			finite_field.WriteGoldilocksField(buf, &coord[j])
		}
	}
	for i := range sc.MessagePoint {
		coord := sc.MessagePoint[i].ToFieldElementSlice()
		for j := range coord {
			finite_field.WriteGoldilocksField(buf, &coord[j])
		}
	}

	return buf.Inner()
}

func (sc *SignatureContent) Commitment() *intMaxGP.PoseidonHashOut {
	flattenSignatureContent := sc.ToFieldElementSlice()
	commitment := intMaxGP.HashNoPad(flattenSignatureContent)

	return commitment
}

func (sc *SignatureContent) Hash() common.Hash {
	commitment := sc.Commitment()
	result := new(intMaxTypes.Bytes32).FromPoseidonHashOut(commitment)

	return common.Hash(result.Bytes())
}

func (sc *SignatureContent) IsValidFormat(publicKeys []intMaxTypes.Uint256) error {
	if len(publicKeys) != intMaxTypes.NumOfSenders {
		return errors.New("public keys length is invalid")
	}

	// sender flag check
	zeroSenderFlag := intMaxTypes.Bytes16{}
	if sc.SenderFlag == zeroSenderFlag {
		return errors.New("sender flag is zero")
	}

	// public key order check
	curPublicKey := publicKeys[0]
	for i := 1; i < len(publicKeys); i++ {
		publicKey := publicKeys[i]
		if curPublicKey.BigInt().Cmp(publicKey.BigInt()) != 1 && !publicKey.IsDummyPublicKey() {
			return errors.New("public key order is invalid")
		}

		curPublicKey = publicKey
	}

	// public keys order and recovery check
	for _, publicKey := range publicKeys {
		_, err := intMaxAcc.NewPublicKeyFromAddressInt(publicKey.BigInt())
		if err != nil {
			return errors.New("public key recovery check failed")
		}
	}

	// message point check
	txTreeRoot := sc.TxTreeRoot.ToFieldElementSlice()
	messagePointExpected := intMaxGP.HashToG2(txTreeRoot)
	messagePoint := intMaxTypes.NewG2AffineFromFlatG2(&sc.MessagePoint)
	if !messagePointExpected.Equal(messagePoint) {
		// fmt.Printf("messagePointExpected: %v\n", messagePointExpected)
		// fmt.Printf("messagePoint: %v\n", messagePoint)
		return errors.New("message point check failed")
	}

	return nil
}

// VerifyAggregation verifies that the calculation of agg_pubkey matches.
// It is assumed that the format validation has already passed.
func (sc *SignatureContent) VerifyAggregation(publicKey []intMaxTypes.Uint256) error {
	if len(publicKey) != intMaxTypes.NumOfSenders {
		return errors.New("public keys length is invalid")
	}

	aggregatedPublicKey := new(intMaxAcc.PublicKey)
	for i, pubKey := range publicKey {
		senderFlagBit := GetBitFromUint32Slice(sc.SenderFlag[:], i)
		publicKey, err := intMaxAcc.NewPublicKeyFromAddressInt(pubKey.BigInt())
		if err != nil {
			return errors.New("public key recovery check failed")
		}

		publicKeysHash := sc.PublicKeyHash.Bytes()
		if senderFlagBit {
			weightedPublicKey := publicKey.WeightByHash(publicKeysHash)
			aggregatedPublicKey.Add(aggregatedPublicKey, weightedPublicKey)
		}
	}

	aggPublicKey := intMaxTypes.NewG1AffineFromFlatG1(&sc.AggPublicKey)
	if !aggregatedPublicKey.Pk.Equal(aggPublicKey) {
		return errors.New("aggregated public key does not match")
	}

	return nil
}
