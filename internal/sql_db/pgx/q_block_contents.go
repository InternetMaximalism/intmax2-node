package pgx

import (
	"encoding/hex"
	"encoding/json"
	"intmax2-node/internal/block_post_service"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

func (p *pgx) CreateBlockContent(
	postedBlock *block_post_service.PostedBlock,
	blockContent *intMaxTypes.BlockContent,
) (*mDBApp.BlockContent, error) {
	blockNumber := int64(postedBlock.BlockNumber)
	blockHash := postedBlock.Hash().Hex()[2:]
	prevBlockHash := postedBlock.PrevBlockHash.Hex()[2:]
	depositRoot := postedBlock.DepositRoot.Hex()[2:]
	signatureHash := postedBlock.SignatureHash.Hex()[2:]
	txRoot := common.Hash(blockContent.TxTreeRoot).Hex()[2:]
	aggregatedSignature := hex.EncodeToString(blockContent.AggregatedSignature.Marshal())
	aggregatedPublicKey := hex.EncodeToString(blockContent.AggregatedPublicKey.Marshal())
	messagePoint := hex.EncodeToString(blockContent.MessagePoint.Marshal())
	isRegistrationBlock := blockContent.SenderType == intMaxTypes.PublicKeySenderType

	senders := make([]intMaxTypes.ColumnSender, len(blockContent.Senders))
	for i, sender := range blockContent.Senders {
		senders[i] = intMaxTypes.ColumnSender{
			AccountID: sender.AccountID,
			PublicKey: sender.PublicKey.ToAddress().String(),
			IsSigned:  sender.IsSigned,
		}
	}
	sendersJSON, err := json.Marshal(senders)
	if err != nil {
		return nil, err
	}

	s := models.BlockContent{
		BlockContentID:      uuid.New().String(),
		BlockNumber:         blockNumber,
		BlockHash:           blockHash,
		PrevBlockHash:       prevBlockHash,
		DepositRoot:         depositRoot,
		SignatureHash:       signatureHash,
		IsRegistrationBlock: isRegistrationBlock,
		Senders:             sendersJSON,
		TxRoot:              txRoot,
		AggregatedPublicKey: aggregatedPublicKey,
		AggregatedSignature: aggregatedSignature,
		MessagePoint:        messagePoint,
		CreatedAt:           time.Now().UTC(),
	}

	const (
		q = `INSERT INTO block_contents (
             id ,block_hash ,prev_block_hash ,deposit_root ,signature_hash
			 ,is_registration_block ,senders ,tx_tree_root ,aggregated_public_key
			 ,aggregated_signature ,message_point ,created_at ,block_number
             ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) `
	)

	_, err = p.exec(p.ctx, q,
		s.BlockContentID, s.BlockHash, s.PrevBlockHash, s.DepositRoot, s.SignatureHash,
		s.IsRegistrationBlock, s.Senders, s.TxRoot, s.AggregatedPublicKey,
		s.AggregatedSignature, s.MessagePoint, s.CreatedAt, s.BlockNumber)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.BlockContent
	bDBApp, err = p.BlockContent(s.BlockContentID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) BlockContent(blockContentID string) (*mDBApp.BlockContent, error) {
	const (
		q = `SELECT
             id ,block_hash ,prev_block_hash ,deposit_root ,signature_hash
			 ,is_registration_block ,senders ,tx_tree_root ,aggregated_public_key
			 ,aggregated_signature ,message_point ,created_at ,block_number
             FROM block_contents WHERE id = $1`
	)

	var tmp models.BlockContent
	err := errPgx.Err(p.queryRow(p.ctx, q, blockContentID).
		Scan(
			&tmp.BlockContentID,
			&tmp.BlockHash,
			&tmp.PrevBlockHash,
			&tmp.DepositRoot,
			&tmp.SignatureHash,
			&tmp.IsRegistrationBlock,
			&tmp.Senders,
			&tmp.TxRoot,
			&tmp.AggregatedPublicKey,
			&tmp.AggregatedSignature,
			&tmp.MessagePoint,
			&tmp.CreatedAt,
			&tmp.BlockNumber,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContent, error) {
	const (
		q = `SELECT
             id ,block_hash ,prev_block_hash ,deposit_root ,signature_hash
			 ,is_registration_block ,senders ,tx_tree_root ,aggregated_public_key
			 ,aggregated_signature ,message_point ,created_at ,block_number
             FROM block_contents WHERE block_number = $1`
	)

	var tmp models.BlockContent
	err := errPgx.Err(p.queryRow(p.ctx, q, blockNumber).
		Scan(
			&tmp.BlockContentID,
			&tmp.BlockHash,
			&tmp.PrevBlockHash,
			&tmp.DepositRoot,
			&tmp.SignatureHash,
			&tmp.IsRegistrationBlock,
			&tmp.Senders,
			&tmp.TxRoot,
			&tmp.AggregatedPublicKey,
			&tmp.AggregatedSignature,
			&tmp.MessagePoint,
			&tmp.CreatedAt,
			&tmp.BlockNumber,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) BlockContentByTxRoot(txRoot string) (*mDBApp.BlockContent, error) {
	const (
		q = `SELECT
			 id ,block_hash ,prev_block_hash ,deposit_root ,signature_hash
			 ,is_registration_block ,senders ,tx_tree_root ,aggregated_public_key
			 ,aggregated_signature ,message_point ,created_at ,block_number
			 FROM block_contents WHERE tx_tree_root = $1`
	)

	var tmp models.BlockContent
	err := errPgx.Err(p.queryRow(p.ctx, q, txRoot).
		Scan(
			&tmp.BlockContentID,
			&tmp.BlockHash,
			&tmp.PrevBlockHash,
			&tmp.DepositRoot,
			&tmp.SignatureHash,
			&tmp.IsRegistrationBlock,
			&tmp.Senders,
			&tmp.TxRoot,
			&tmp.AggregatedPublicKey,
			&tmp.AggregatedSignature,
			&tmp.MessagePoint,
			&tmp.CreatedAt,
			&tmp.BlockNumber,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) blockContentToDBApp(tmp *models.BlockContent) *mDBApp.BlockContent {
	blockNumber := uint32(tmp.BlockNumber)
	m := mDBApp.BlockContent{
		BlockContentID:      tmp.BlockContentID,
		BlockNumber:         blockNumber,
		BlockHash:           tmp.BlockHash,
		PrevBlockHash:       tmp.PrevBlockHash,
		DepositRoot:         tmp.DepositRoot,
		SignatureHash:       tmp.SignatureHash,
		IsRegistrationBlock: tmp.IsRegistrationBlock,
		Senders:             tmp.Senders,
		TxRoot:              tmp.TxRoot,
		AggregatedPublicKey: tmp.AggregatedPublicKey,
		AggregatedSignature: tmp.AggregatedSignature,
		MessagePoint:        tmp.MessagePoint,
		CreatedAt:           tmp.CreatedAt,
	}

	return &m
}
