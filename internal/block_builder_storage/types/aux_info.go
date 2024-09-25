package types

import (
	"encoding/json"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/logger"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// AuxInfo is a structure for recording past tree states.
type AuxInfo struct {
	BlockContent *intMaxTypes.BlockContent
	PostedBlock  *block_post_service.PostedBlock
}

func BlockAuxInfoFromBlockContent(
	log logger.Logger,
	auxInfo *mDBApp.BlockContentWithProof,
) (*AuxInfo, error) {
	const (
		mask0x = "0x%s"
	)

	decodedAggregatedPublicKeyPoint, err := hexutil.Decode(fmt.Sprintf(mask0x, auxInfo.AggregatedPublicKey))
	if err != nil {
		return nil, fmt.Errorf("aggregated public key hex decode error: %w", err)
	}

	aggregatedPublicKeyPoint := new(bn254.G1Affine)
	err = aggregatedPublicKeyPoint.Unmarshal(decodedAggregatedPublicKeyPoint)
	if err != nil {
		return nil, fmt.Errorf("aggregated public key unmarshal error: %w", err)
	}

	var aggregatedPublicKey *intMaxAcc.PublicKey
	aggregatedPublicKey, err = intMaxAcc.NewPublicKey(aggregatedPublicKeyPoint)
	if err != nil {
		return nil, fmt.Errorf("aggregated public key error: %w", err)
	}

	var decodedAggregatedSignature []byte
	decodedAggregatedSignature, err = hexutil.Decode(fmt.Sprintf(mask0x, auxInfo.AggregatedSignature))
	if err != nil {
		return nil, fmt.Errorf("aggregated signature hex decode error: %w", err)
	}

	aggregatedSignature := new(bn254.G2Affine)
	err = aggregatedSignature.Unmarshal(decodedAggregatedSignature)
	if err != nil {
		return nil, fmt.Errorf("aggregated signature unmarshal error: %w", err)
	}

	var decodedMessagePoint []byte
	decodedMessagePoint, err = hexutil.Decode(fmt.Sprintf(mask0x, auxInfo.MessagePoint))
	if err != nil {
		return nil, fmt.Errorf("aggregated message point hex decode error: %w", err)
	}

	messagePoint := new(bn254.G2Affine)
	err = messagePoint.Unmarshal(decodedMessagePoint)
	if err != nil {
		return nil, fmt.Errorf("message point unmarshal error: %w", err)
	}

	var columnSenders []intMaxTypes.ColumnSender
	err = json.Unmarshal(auxInfo.Senders, &columnSenders)
	if err != nil {
		return nil, fmt.Errorf("senders unmarshal error: %w", err)
	}

	senders := make([]intMaxTypes.Sender, len(columnSenders))
	for i, sender := range columnSenders {
		var publicKey *intMaxAcc.PublicKey
		publicKey, err = intMaxAcc.NewPublicKeyFromAddressHex(sender.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("public key unmarshal decode error: %w", err)
		}

		senders[i] = intMaxTypes.Sender{
			AccountID: sender.AccountID,
			PublicKey: publicKey,
			IsSigned:  sender.IsSigned,
		}
	}

	var senderType string
	if auxInfo.IsRegistrationBlock {
		senderType = intMaxTypes.PublicKeySenderType
	} else {
		senderType = intMaxTypes.AccountIDSenderType
	}

	blockContent := intMaxTypes.BlockContent{
		TxTreeRoot:          common.HexToHash(fmt.Sprintf(mask0x, auxInfo.TxRoot)),
		AggregatedPublicKey: aggregatedPublicKey,
		AggregatedSignature: aggregatedSignature,
		MessagePoint:        messagePoint,
		Senders:             senders,
		SenderType:          senderType,
	}

	postedBlock := block_post_service.PostedBlock{
		BlockNumber:   auxInfo.BlockNumber,
		PrevBlockHash: common.HexToHash(fmt.Sprintf(mask0x, auxInfo.PrevBlockHash)),
		DepositRoot:   common.HexToHash(fmt.Sprintf(mask0x, auxInfo.DepositRoot)),
		SignatureHash: common.HexToHash(fmt.Sprintf(mask0x, auxInfo.SignatureHash)), // TODO: Calculate from blockContent
	}

	if blockHash := postedBlock.Hash(); blockHash.Hex() != fmt.Sprintf(mask0x, auxInfo.BlockHash) {
		log.Errorf("postedBlock: %v", postedBlock)
		log.Errorf("blockHash: %s != %s", blockHash.Hex(), auxInfo.BlockHash)
		panic("block hash mismatch")
	}

	return &AuxInfo{
		PostedBlock:  &postedBlock,
		BlockContent: &blockContent,
	}, nil
}
