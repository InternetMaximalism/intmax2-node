package block_post_service

import (
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

type accountInfo struct {
	dbApp SQLDriverApp
}

func NewAccountInfo(dbApp SQLDriverApp) AccountInfo {
	return &accountInfo{
		dbApp: dbApp,
	}
}

// func (ai *accountInfo) RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSeenBlockNumber uint32) (accID uint64, err error) {
// 	var sender *mDBApp.Sender
// 	sender, err = ai.dbApp.SenderByAddress(pk.ToAddress().String())
// 	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
// 		return 0, errors.Join(ErrSenderByAddressFail, err)
// 	}
// 	if errors.Is(err, errorsDB.ErrNotFound) {
// 		var newSender *mDBApp.Sender
// 		newSender, err = ai.dbApp.CreateSenders(pk.ToAddress().String(), pk.String())
// 		if err != nil {
// 			return 0, errors.Join(ErrCreateSendersFail, err)
// 		}
// 		sender = &mDBApp.Sender{
// 			ID:        newSender.ID,
// 			Address:   newSender.Address,
// 			PublicKey: newSender.PublicKey,
// 			CreatedAt: newSender.CreatedAt,
// 		}

// 		_, err = ai.dbApp.CreateAccount(sender.ID)
// 		if err != nil {
// 			return 0, errors.Join(ErrCreateAccountFail, err)
// 		}
// 	}

// 	account, err := ai.dbApp.AccountBySenderID(sender.ID)
// 	if err != nil && !errors.Is(err, errorsDB.ErrNotFound) {
// 		return 0, errors.Join(ErrAccountBySenderIDFail, err)
// 	}
// 	if errors.Is(err, errorsDB.ErrNotFound) {
// 		_, err = ai.dbApp.CreateAccount(sender.ID)
// 		if err != nil {
// 			return 0, errors.Join(ErrCreateAccountFail, err)
// 		}
// 	}

// 	return account.AccountID.Uint64(), nil
// }

func (ai *accountInfo) PublicKeyByAccountID(blockNumber uint32, accountID uint64) (pk *intMaxAcc.PublicKey, err error) {
	var accID uint256.Int
	accID.SetUint64(accountID)

	var acc *mDBApp.Account
	acc, err = ai.dbApp.AccountByAccountID(&accID)
	if err != nil {
		return nil, errors.Join(ErrAccountByAccountIDFail, err)
	}

	var sender *mDBApp.Sender
	sender, err = ai.dbApp.SenderByID(acc.SenderID)
	if err != nil {
		return nil, errors.Join(ErrSenderByIDFail, err)
	}

	pk, err = intMaxAcc.HexToPublicKey(sender.PublicKey)
	if err != nil {
		return nil, errors.Join(ErrDecodeHexToPublicKeyFail, err)
	}

	return pk, nil
}

func (ai *accountInfo) AccountBySenderAddress(senderAddress string) (accID *uint256.Int, err error) {
	var sender *mDBApp.Sender
	sender, err = ai.dbApp.SenderByAddress(senderAddress)
	if err != nil {
		return nil, errors.Join(ErrSenderByAddressFail, err)
	}

	var acc *mDBApp.Account
	acc, err = ai.dbApp.AccountBySenderID(sender.ID)
	if err != nil {
		return nil, errors.Join(ErrAccountBySenderIDFail, err)
	}

	return acc.AccountID, nil
}
