package block_validity_prover_account

import (
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_account_test -source=db_app.go

type SQLDriverApp interface {
	Senders
	Accounts
}

type Senders interface {
	SenderByAddress(address string) (*mDBApp.Sender, error)
}

type Accounts interface {
	AccountBySenderID(senderID string) (*mDBApp.Account, error)
}
