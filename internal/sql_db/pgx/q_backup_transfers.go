package pgx

import (
	"database/sql"
	"errors"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	"intmax2-node/internal/sql_filter"
	mFL "intmax2-node/internal/sql_filter/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	errorsDB "intmax2-node/pkg/sql_db/errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	libPGX "github.com/jackc/pgx/v5"
)

func (p *pgx) CreateBackupTransfer(
	recipient, encryptedTransferHash, encryptedTransfer string,
	senderLastBalanceProofBody, senderBalanceTransitionProofBody []byte,
	blockNumber int64,
) (*mDBApp.BackupTransfer, error) {
	const query = `
	    INSERT INTO backup_transfers
        (id, recipient, transfer_double_hash, encrypted_transfer, sender_last_balance_proof_body, sender_balance_transition_proof_body, block_number, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(query, id, recipient, encryptedTransferHash, encryptedTransfer, senderLastBalanceProofBody, senderBalanceTransitionProofBody, blockNumber, createdAt)
	if err != nil {
		return nil, err
	}

	return p.GetBackupTransfer("id", id)
}

func (p *pgx) GetBackupTransfer(condition, value string) (*mDBApp.BackupTransfer, error) {
	const baseQuery = `
        SELECT id, recipient, transfer_double_hash, encrypted_transfer, sender_last_balance_proof_body, sender_balance_transition_proof_body, block_number, created_at
        FROM backup_transfers
        WHERE %s = $1
`
	query := fmt.Sprintf(baseQuery, condition)

	var b models.BackupTransfer
	err := errPgx.Err(p.queryRow(p.ctx, query, value).
		Scan(
			&b.ID,
			&b.Recipient,
			&b.TransferDoubleHash,
			&b.EncryptedTransfer,
			&b.SenderLastBalanceProofBody,
			&b.SenderBalanceTransitionProofBody,
			&b.BlockNumber,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, err
	}
	transfer := p.backupTransferToDBApp(&b)
	return &transfer, nil
}

func (p *pgx) GetBackupTransferByRecipientAndTransferDoubleHash(
	recipient, transferDoubleHash string,
) (*mDBApp.BackupTransfer, error) {
	const (
		q = `
        SELECT id, recipient, transfer_double_hash, encrypted_transfer, block_number, created_at
        FROM backup_transfers
        WHERE recipient = $1 AND transfer_double_hash = $2 `
	)

	var b models.BackupTransfer
	err := errPgx.Err(p.queryRow(p.ctx, q, recipient, transferDoubleHash).
		Scan(
			&b.ID,
			&b.Recipient,
			&b.TransferDoubleHash,
			&b.EncryptedTransfer,
			&b.BlockNumber,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, err
	}
	transfer := p.backupTransferToDBApp(&b)
	return &transfer, nil
}

func (p *pgx) GetBackupTransfers(condition string, value interface{}) ([]*mDBApp.BackupTransfer, error) {
	const baseQuery = `
        SELECT id, recipient, transfer_double_hash, encrypted_transfer, block_number, created_at
        FROM backup_transfers
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)
	var transfers []*mDBApp.BackupTransfer
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) error {
		var b models.BackupTransfer
		err := rows.Scan(&b.ID, &b.Recipient, &b.TransferDoubleHash, &b.EncryptedTransfer, &b.BlockNumber, &b.CreatedAt)
		if err != nil {
			return err
		}
		transfer := p.backupTransferToDBApp(&b)
		transfers = append(transfers, &transfer)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transfers, nil
}

func (p *pgx) GetBackupTransfersByRecipient(
	recipient string,
	pagination mDBApp.PaginationOfListOfBackupTransfersInput,
	sorting mFL.Sorting, orderBy mFL.OrderBy,
	filters mFL.FiltersList,
) (
	paginator *mDBApp.PaginationOfListOfBackupTransfers,
	listDBApp mDBApp.ListOfBackupTransfer,
	err error,
) {
	var (
		q = `
SELECT id, recipient, transfer_double_hash, encrypted_transfer, block_number, created_at
FROM backup_transfers
WHERE recipient = @recipient %s
`
	)

	sorting = mFL.Sorting(strings.TrimSpace(string(sorting)))
	if sorting == "" {
		sorting = mFL.SortingDESC
	}

	var (
		cursor       string
		orderByValue string
	)
	switch orderBy {
	case mFL.DateCreate:
		const createdAtKey = "created_at"
		orderByValue = createdAtKey
		if pagination.Cursor != nil {
			cursor = time.Unix(0, pagination.Cursor.SortingValue.Int64()).UTC().Format(time.RFC3339Nano)
		}
	default:
		orderBy = mFL.DateCreate
		const startedAtKey = "created_at"
		orderByValue = startedAtKey
		if pagination.Cursor != nil {
			cursor = time.Unix(0, pagination.Cursor.SortingValue.Int64()).UTC().Format(time.RFC3339Nano)
		}
	}

	wParams := make(libPGX.NamedArgs)
	wParams["recipient"] = recipient

	var where string
	if len(filters) > 0 {
		var (
			fc     sql_filter.SQLFilter
			params map[string]interface{}
		)
		where, params = fc.FilterDataToWhereQuery(filters)
		where = fc.PrepareWhereString(where, false)
		where = fmt.Sprintf("AND (%s)", where)

		for pKey := range params {
			wParams[pKey] = params[pKey]
		}
	}

	var revers bool
	if pagination.Cursor != nil {
		rID := pagination.Cursor.BN
		cond := mFL.LessSymbol
		if sorting == mFL.SortingDESC && pagination.Direction == mFL.DirectionNext {
			cond = mFL.LessSymbol
		} else if sorting == mFL.SortingDESC && pagination.Direction == mFL.DirectionPrev {
			sorting = mFL.SortingASC
			cond = mFL.MoreSymbol
			revers = true
		} else if sorting == mFL.SortingASC && pagination.Direction == mFL.DirectionNext {
			cond = mFL.MoreSymbol
		} else if sorting == mFL.SortingASC && pagination.Direction == mFL.DirectionPrev {
			sorting = mFL.SortingDESC
			cond = mFL.LessSymbol
			revers = true
		}
		if revers && sorting == mFL.SortingASC ||
			sorting == mFL.SortingASC && pagination.Direction == mFL.DirectionNext {
			where += fmt.Sprintf(
				"AND ((block_number, %s) %s ('%s', '%s') AND %s %s '%s')",
				orderByValue, cond, rID, cursor, orderByValue, cond, cursor)
		} else {
			where += fmt.Sprintf(
				"AND ((block_number, %s) %s ('%s', '%s'))",
				orderByValue, cond, rID, cursor)
		}
	}

	q += fmt.Sprintf(" ORDER BY block_number %s, %s %s", sorting, orderByValue, sorting)

	q += fmt.Sprintf(" FETCH FIRST %d ROWS ONLY ", pagination.Offset)

	var rows *sql.Rows
	rows, err = p.query(p.ctx, fmt.Sprintf(q, where), wParams)
	if err != nil {
		err = errPgx.Err(err)
		if errors.Is(err, errorsDB.ErrNotFound) {
			return nil, nil, nil
		}

		return nil, nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var list models.ListOfBackupTransfer
	for rows.Next() {
		var b models.BackupTransfer
		err = rows.Scan(
			&b.ID,
			&b.Recipient,
			&b.TransferDoubleHash,
			&b.EncryptedTransfer,
			&b.BlockNumber,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, nil, err
		}
		list = append(list, b)
	}

	listDBApp = make(mDBApp.ListOfBackupTransfer, len(list))
	if revers {
		for key := len(list) - 1; key >= 0; key-- {
			listDBApp[len(list)-1-key] = p.backupTransferToDBApp(&list[key])
		}
	} else {
		for key := range list {
			listDBApp[key] = p.backupTransferToDBApp(&list[key])
		}
	}

	paginator = &mDBApp.PaginationOfListOfBackupTransfers{
		Offset: pagination.Offset,
	}

	if list != nil {
		const (
			int0Key = 0
			int1Key = 1
		)

		startV := int0Key
		endV := len(list) - int1Key
		if revers {
			startV = len(list) - int1Key
			endV = int0Key
		}

		paginator.Cursor = &mDBApp.CursorListOfBackupTransfers{
			Prev: &mDBApp.CursorBaseOfListOfBackupTransfers{
				BN: new(big.Int).SetUint64(list[startV].BlockNumber),
			},
			Next: &mDBApp.CursorBaseOfListOfBackupTransfers{
				BN: new(big.Int).SetUint64(list[endV].BlockNumber),
			},
		}

		switch orderBy { // nolint:gocritic
		case mFL.DateCreate:
			paginator.Cursor.Prev.SortingValue = new(big.Int).SetInt64(list[startV].CreatedAt.UTC().UnixNano())
			paginator.Cursor.Next.SortingValue = new(big.Int).SetInt64(list[endV].CreatedAt.UTC().UnixNano())
		}
	}

	return paginator, listDBApp, nil
}

func (p *pgx) backupTransferToDBApp(b *models.BackupTransfer) mDBApp.BackupTransfer {
	return mDBApp.BackupTransfer{
		ID:                               b.ID,
		Recipient:                        b.Recipient,
		TransferDoubleHash:               b.TransferDoubleHash.String,
		EncryptedTransfer:                b.EncryptedTransfer,
		SenderLastBalanceProofBody:       b.SenderLastBalanceProofBody,
		SenderBalanceTransitionProofBody: b.SenderBalanceTransitionProofBody,
		BlockNumber:                      b.BlockNumber,
		CreatedAt:                        b.CreatedAt,
	}
}
