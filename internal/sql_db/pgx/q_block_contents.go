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
) (*mDBApp.BlockContentWithProof, error) {
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

	var bDBApp *mDBApp.BlockContentWithProof
	bDBApp, err = p.BlockContent(s.BlockContentID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) CreateValidityProof(
	blockContentID string,
	validityProof []byte,
) (*mDBApp.BlockProof, error) {
	const (
		q = `INSERT INTO block_validity_proofs (
			 block_content_id ,validity_proof
			 ) VALUES ($1, $2)`
	)

	_, err := p.exec(p.ctx, q, blockContentID, validityProof)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.BlockProof
	bDBApp, err = p.BlockValidityProof(blockContentID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) BlockContent(blockContentID string) (*mDBApp.BlockContentWithProof, error) {
	const (
		q = `SELECT
             bc.id ,bc.block_hash ,bc.prev_block_hash ,bc.deposit_root ,bc.signature_hash
			 ,bc.is_registration_block ,bc.senders ,bc.tx_tree_root ,bc.aggregated_public_key
			 ,bc.aggregated_signature ,bc.message_point ,bc.created_at ,bc.block_number
			 ,bp.validity_proof
             FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bc.id = $1`
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
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

// TODO: pagination
func (p *pgx) ScanBlockHashAndSenders() (blockHashAndSendersMap map[uint32]mDBApp.BlockHashAndSenders, lastBlockNumber uint32, err error) {
	const (
		q = `SELECT block_number, block_hash, deposit_root, senders FROM block_contents
			 ORDER BY block_number ASC`
	)

	rows, err := p.query(p.ctx, q)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	blockHashAndSendersMap = make(map[uint32]mDBApp.BlockHashAndSenders)
	lastBlockNumber = uint32(0)
	for rows.Next() {
		var blockNumber uint32
		var blockHash string
		var depositTreeRoot string
		var sendersJSON []byte
		err = rows.Scan(&blockNumber, &blockHash, &depositTreeRoot, &sendersJSON)
		if err != nil {
			return nil, 0, err
		}

		lastBlockNumber = blockNumber

		var senders []intMaxTypes.ColumnSender
		err = json.Unmarshal(sendersJSON, &senders)
		if err != nil {
			return nil, 0, err
		}

		blockHashAndSendersMap[blockNumber] = mDBApp.BlockHashAndSenders{
			BlockHash:       blockHash,
			Senders:         senders,
			DepositTreeRoot: depositTreeRoot,
		}
	}

	return blockHashAndSendersMap, lastBlockNumber, nil
}

func (p *pgx) BlockValidityProof(blockContentID string) (*mDBApp.BlockProof, error) {
	const (
		q = `SELECT
			 block_content_id ,validity_proof
			 FROM block_validity_proofs
			 WHERE block_content_id = $1`
	)

	var tmp models.BlockProof
	err := errPgx.Err(p.queryRow(p.ctx, q, blockContentID).
		Scan(
			&tmp.BlockContentID,
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := &mDBApp.BlockProof{
		BlockContentID: tmp.BlockContentID,
		ValidityProof:  tmp.ValidityProof,
	}

	return bDBApp, nil
}

func (p *pgx) LastBlockNumberGeneratedValidityProof() (uint32, error) {
	const (
		q = `SELECT
			 bc.block_number
			 FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bp.validity_proof IS NOT NULL
			 ORDER BY bc.block_number DESC
			 LIMIT 1`
	)

	var blockNumber int64
	err := errPgx.Err(p.queryRow(p.ctx, q).Scan(&blockNumber))
	if err != nil {
		return 0, err
	}

	return uint32(blockNumber), nil
}

func (p *pgx) LastBlockValidityProof() (*mDBApp.BlockContentWithProof, error) {
	const (
		q = `SELECT
			 bc.id ,bc.block_hash ,bc.prev_block_hash ,bc.deposit_root ,bc.signature_hash
			 ,bc.is_registration_block ,bc.senders ,bc.tx_tree_root ,bc.aggregated_public_key
			 ,bc.aggregated_signature ,bc.message_point ,bc.created_at ,bc.block_number
			 ,bp.validity_proof
			 FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bp.validity_proof IS NOT NULL
			 ORDER BY bc.block_number DESC
			 LIMIT 1`
	)

	var tmp models.BlockContent
	err := errPgx.Err(p.queryRow(p.ctx, q).
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
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error) {
	const (
		q = `SELECT
             bc.id ,bc.block_hash ,bc.prev_block_hash ,bc.deposit_root ,bc.signature_hash
			 ,bc.is_registration_block ,bc.senders ,bc.tx_tree_root ,bc.aggregated_public_key
			 ,bc.aggregated_signature ,bc.message_point ,bc.created_at ,bc.block_number
			 ,bp.validity_proof
             FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bc.block_number = $1`
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
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) BlockContentByTxRoot(txRoot string) (*mDBApp.BlockContentWithProof, error) {
	const (
		q = `SELECT
             bc.id ,bc.block_hash ,bc.prev_block_hash ,bc.deposit_root ,bc.signature_hash
			 ,bc.is_registration_block ,bc.senders ,bc.tx_tree_root ,bc.aggregated_public_key
			 ,bc.aggregated_signature ,bc.message_point ,bc.created_at ,bc.block_number
			 ,bp.validity_proof
             FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bc.tx_tree_root = $1`
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
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) blockContentToDBApp(tmp *models.BlockContent) *mDBApp.BlockContentWithProof {
	blockNumber := uint32(tmp.BlockNumber)
	m := mDBApp.BlockContentWithProof{
		BlockContent: mDBApp.BlockContent{
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
		},
		ValidityProof: tmp.ValidityProof,
	}

	return &m
}
