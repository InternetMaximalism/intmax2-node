package pgx

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"intmax2-node/internal/intmax_block_content"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/holiman/uint256"
)

func (p *pgx) CreateBlockContent(
	postedBlock *intmax_block_content.PostedBlock,
	blockContent *intMaxTypes.BlockContent,
	l2BlockNumber *uint256.Int,
	l2BlockHash common.Hash,
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
	valueL2BlockNumber, _ := l2BlockNumber.Value()

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
		BlockNumberL2:       l2BlockNumber,
		BlockHashL2:         sql.NullString{Valid: true, String: l2BlockHash.String()},
		CreatedAt:           time.Now().UTC(),
	}

	const (
		q = `INSERT INTO block_contents (
             id ,block_hash ,prev_block_hash ,deposit_root ,signature_hash
			 ,is_registration_block ,senders ,tx_tree_root ,aggregated_public_key
			 ,aggregated_signature ,message_point ,created_at ,block_number, block_number_l2, block_hash_l2
             ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) `
	)

	_, err = p.exec(p.ctx, q,
		s.BlockContentID, s.BlockHash, s.PrevBlockHash, s.DepositRoot, s.SignatureHash,
		s.IsRegistrationBlock, s.Senders, s.TxRoot, s.AggregatedPublicKey,
		s.AggregatedSignature, s.MessagePoint, s.CreatedAt, s.BlockNumber, valueL2BlockNumber, l2BlockHash.String())
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

func (p *pgx) BlockContentIDByL2BlockNumber(l2BlockNumber string) (bcID string, err error) {
	const (
		emptyKey = ""

		q = `SELECT id FROM block_contents WHERE block_number_l2 = $1`
	)

	err = errPgx.Err(p.queryRow(p.ctx, q, l2BlockNumber).
		Scan(
			&bcID,
		))
	if err != nil {
		return emptyKey, err
	}

	return bcID, nil
}

// Insert a new validity proof or do nothing if it already exists
func (p *pgx) CreateValidityProof(
	blockHash common.Hash,
	validityProof []byte,
) (*mDBApp.BlockProof, error) {
	fmt.Printf("(CreateValidityProof) blockHash: %s\n", blockHash)
	// 1. If a block_content corresponding to the specified block_hash exists,
	//    and there is no row in block_validity_proofs corresponding to that block_content_id:
	//    A new row is inserted.
	// 2. If a block_content corresponding to the specified block_hash exists,
	//    and there is already a row in block_validity_proofs corresponding to that block_content_id:
	//    No changes are made (the old value is retained).
	// 3. If no block_content corresponding to the specified block_hash exists:
	//    Nothing is inserted.
	const (
		q = `WITH block_content AS (
				 SELECT id
				 FROM block_contents
				 WHERE block_hash = $1
			 )
			 INSERT INTO block_validity_proofs (block_content_id, validity_proof)
			 SELECT block_content.id, $2
			 FROM block_content
			 ON CONFLICT (block_content_id) DO NOTHING;`
	)

	blockHashHex := blockHash.Hex()[2:]

	_, err := p.exec(p.ctx, q, blockHashHex, validityProof)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	bDBApp, err := p.BlockValidityProofByBlockHash(blockHash)
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
			 ,bc.block_number_l2 ,bc.block_hash_l2 ,bc.deposit_leaves_counter
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
			&tmp.BlockNumberL2,
			&tmp.BlockHashL2,
			&tmp.DepositLeavesCounter,
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
		q = `SELECT block_number, block_hash, deposit_root, senders, is_registration_block FROM block_contents
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
		var isRegistrationBlock bool
		err = rows.Scan(&blockNumber, &blockHash, &depositTreeRoot, &sendersJSON, &isRegistrationBlock)
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
			BlockHash:           blockHash,
			Senders:             senders,
			DepositTreeRoot:     depositTreeRoot,
			IsRegistrationBlock: isRegistrationBlock,
		}
	}

	return blockHashAndSendersMap, lastBlockNumber, nil
}

func (p *pgx) BlockValidityProofByBlockHash(blockHash common.Hash) (*mDBApp.BlockProof, error) {
	const (
		q = `SELECT
			 bp.block_content_id ,bp.validity_proof
			 FROM block_validity_proofs bp
			 JOIN block_contents bc ON bp.block_content_id = bc.id
			 WHERE bc.block_hash = $1`
	)

	blockHashHex := blockHash.Hex()[2:]

	var tmp models.BlockProof
	err := errPgx.Err(p.queryRow(p.ctx, q, blockHashHex).
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
		fmt.Printf("(LastBlockNumberGeneratedValidityProof) blockNumber not found")
		return 0, err
	}
	fmt.Printf("(LastBlockNumberGeneratedValidityProof) blockNumber: %d\n", blockNumber)

	return uint32(blockNumber), nil
}

func (p *pgx) LastPostedBlockNumber() (uint32, error) {
	const (
		q = `SELECT
			 bc.block_number
			 FROM block_contents bc
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
			 ,bc.block_number_l2 ,bc.block_hash_l2 ,bc.deposit_leaves_counter
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
			&tmp.BlockNumberL2,
			&tmp.BlockHashL2,
			&tmp.DepositLeavesCounter,
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
			 ,bc.block_number_l2 ,bc.block_hash_l2 ,bc.deposit_leaves_counter
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
			&tmp.BlockNumberL2,
			&tmp.BlockHashL2,
			&tmp.DepositLeavesCounter,
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) BlockContentByBlockHash(blockHash string) (*mDBApp.BlockContentWithProof, error) {
	const (
		q = `SELECT
             bc.id ,bc.block_hash ,bc.prev_block_hash ,bc.deposit_root ,bc.signature_hash
			 ,bc.is_registration_block ,bc.senders ,bc.tx_tree_root ,bc.aggregated_public_key
			 ,bc.aggregated_signature ,bc.message_point ,bc.created_at ,bc.block_number
			 ,bc.block_number_l2 ,bc.block_hash_l2 ,bc.deposit_leaves_counter
			 ,bp.validity_proof
             FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bc.block_hash = $1`
	)

	var tmp models.BlockContent
	err := errPgx.Err(p.queryRow(p.ctx, q, blockHash).
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
			&tmp.BlockNumberL2,
			&tmp.BlockHashL2,
			&tmp.DepositLeavesCounter,
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error) {
	const (
		q = `SELECT
             bc.id ,bc.block_hash ,bc.prev_block_hash ,bc.deposit_root ,bc.signature_hash
			 ,bc.is_registration_block ,bc.senders ,bc.tx_tree_root ,bc.aggregated_public_key
			 ,bc.aggregated_signature ,bc.message_point ,bc.created_at ,bc.block_number
			 ,bc.block_number_l2 ,bc.block_hash_l2 ,bc.deposit_leaves_counter
			 ,bp.validity_proof
             FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bc.tx_tree_root = $1`
	)

	var tmp models.BlockContent
	err := errPgx.Err(p.queryRow(p.ctx, q, txRoot.Hex()[2:]).
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
			&tmp.BlockNumberL2,
			&tmp.BlockHashL2,
			&tmp.DepositLeavesCounter,
			&tmp.ValidityProof,
		))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block content by tx root: %w", err)
	}

	bDBApp := p.blockContentToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) BlockContentListByTxRoot(txRoot ...common.Hash) ([]*mDBApp.BlockContentWithProof, error) {
	const (
		q = `SELECT
             bc.id ,bc.block_hash ,bc.prev_block_hash ,bc.deposit_root ,bc.signature_hash
			 ,bc.is_registration_block ,bc.senders ,bc.tx_tree_root ,bc.aggregated_public_key
			 ,bc.aggregated_signature ,bc.message_point ,bc.created_at ,bc.block_number
			 ,bc.block_number_l2 ,bc.block_hash_l2 ,bc.deposit_leaves_counter
			 ,bp.validity_proof
             FROM block_contents bc
			 LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
			 WHERE bc.tx_tree_root = any($1)
			 ORDER BY bc.block_number ASC `
	)

	txsRoot := make([]string, len(txRoot))
	for key := range txRoot {
		txsRoot[key] = txRoot[key].Hex()[2:]
	}

	rows, err := p.query(p.ctx, q, txsRoot)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var bDBApp []*mDBApp.BlockContentWithProof
	for rows.Next() {
		var tmp models.BlockContent
		err = errPgx.Err(rows.Scan(
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
			&tmp.BlockNumberL2,
			&tmp.BlockHashL2,
			&tmp.DepositLeavesCounter,
			&tmp.ValidityProof,
		))
		if err != nil {
			return nil, err
		}

		bDBApp = append(bDBApp, p.blockContentToDBApp(&tmp))
	}

	return bDBApp, nil
}

func (p *pgx) BlockContentUpdDepositLeavesCounterByBlockNumber(
	blockNumber, depositLeavesCounter uint32,
) error {
	const (
		q = ` UPDATE block_contents SET deposit_leaves_counter = $1
              WHERE deposit_leaves_counter != $2 AND block_number = $3 `
	)

	_, err := p.exec(p.ctx, q, depositLeavesCounter, depositLeavesCounter, blockNumber)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) blockContentToDBApp(tmp *models.BlockContent) *mDBApp.BlockContentWithProof {
	blockNumber := uint32(tmp.BlockNumber)
	m := mDBApp.BlockContentWithProof{
		BlockContent: mDBApp.BlockContent{
			BlockContentID:       tmp.BlockContentID,
			BlockNumber:          blockNumber,
			BlockHash:            tmp.BlockHash,
			PrevBlockHash:        tmp.PrevBlockHash,
			DepositRoot:          tmp.DepositRoot,
			DepositLeavesCounter: uint32(tmp.DepositLeavesCounter),
			SignatureHash:        tmp.SignatureHash,
			IsRegistrationBlock:  tmp.IsRegistrationBlock,
			Senders:              tmp.Senders,
			TxRoot:               tmp.TxRoot,
			AggregatedPublicKey:  tmp.AggregatedPublicKey,
			AggregatedSignature:  tmp.AggregatedSignature,
			MessagePoint:         tmp.MessagePoint,
			BlockNumberL2:        tmp.BlockNumberL2,
			BlockHashL2:          tmp.BlockHashL2.String,
			CreatedAt:            tmp.CreatedAt,
		},
		ValidityProof: tmp.ValidityProof,
	}

	return &m
}
