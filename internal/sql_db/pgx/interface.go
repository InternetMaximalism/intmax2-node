package pgx

import (
	"context"
	"encoding/json"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	mFL "intmax2-node/internal/sql_filter/models"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type PGX interface {
	GenericCommands
	ServiceCommands
	Blocks
	Tokens
	Signatures
	TxMerkleProofs
	EventBlockNumbers
	Balances
	Withdrawals
	BackupTransfers
	BackupTransactions
	BackupDeposits
	BackupSenderProofs
	CtrlEventBlockNumbersJobs
	EventBlockNumbersErrors
	EventBlockNumbersForValidityProver
	Senders
	Accounts
	BlockParticipants
	BackupBalances
	Deposits
	BlockContents
	CtrlProcessingJobs
	GasPriceOracle
	L2BatchIndex
	RelationshipL2BatchIndexAndBlockContent
	EthereumCounterparties
}

type GenericCommands interface {
	Begin(ctx context.Context) (interface{}, error)
	Rollback()
	Commit() error
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Migrator(ctx context.Context, command string) (step int, err error)
	Check(ctx context.Context) health.Health
}

type Blocks interface {
	CreateBlock(
		builderPublicKey, txRoot, aggregatedSignature, aggregatedPublicKey string, senders []intMaxTypes.ColumnSender,
		senderType uint,
		options []byte,
	) (*mDBApp.Block, error)
	Block(proposalBlockID string) (*mDBApp.Block, error)
	BlockByTxRoot(txRoot string) (*mDBApp.Block, error)
	UpdateBlockStatus(proposalBlockID string, blockHash string, blockNumber uint32) error
	GetUnprocessedBlocks() ([]*mDBApp.Block, error)
}

type Tokens interface {
	CreateToken(
		tokenIndex, tokenAddress string,
		tokenID *uint256.Int,
	) (*mDBApp.Token, error)
	TokenByIndex(tokenIndex string) (*mDBApp.Token, error)
	TokenByTokenInfo(tokenAddress, tokenID string) (*mDBApp.Token, error)
}

type Signatures interface {
	CreateSignature(signature, proposalBlockID string) (*mDBApp.Signature, error)
	SignatureByID(signatureID string) (*mDBApp.Signature, error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash, signatureID string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
		txTreeRoot string,
		proposalBlockID string,
	) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByTxHash(txHash string) (*mDBApp.TxMerkleProofs, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*mDBApp.EventBlockNumber, error)
	EventBlockNumbersByEventNames(eventNames []string) ([]*mDBApp.EventBlockNumber, error)
}

type Balances interface {
	CreateBalance(userAddress, tokenAddress, balance string) (*mDBApp.Balance, error)
	UpdateBalanceByID(balanceID, balance string) (*mDBApp.Balance, error)
	BalanceByID(id string) (*mDBApp.Balance, error)
	BalanceByUserAndTokenIndex(userAddress, tokenIndex string) (*mDBApp.Balance, error)
	BalanceByUserAndTokenInfo(userAddress, tokenAddress, tokenID string) (*mDBApp.Balance, error)
}

type Withdrawals interface {
	CreateWithdrawal(
		id string,
		transferData *mDBApp.TransferData,
		transferMerkleProof *mDBApp.TransferMerkleProof,
		transaction *mDBApp.Transaction,
		txMerkleProof *mDBApp.TxMerkleProof,
		transferHash string,
		blockNumber int64,
		blockHash string,
		enoughBalanceProof *mDBApp.EnoughBalanceProof,
	) (*mDBApp.Withdrawal, error)
	UpdateWithdrawalsStatus(ids []string, status mDBApp.WithdrawalStatus) error
	WithdrawalByID(id string) (*mDBApp.Withdrawal, error)
	WithdrawalsByHashes(transferHashes []string) (*[]mDBApp.Withdrawal, error)
	WithdrawalsByStatus(status mDBApp.WithdrawalStatus, limit *int) (*[]mDBApp.Withdrawal, error)
}

type BackupTransfers interface {
	CreateBackupTransfer(
		recipient, encryptedTransferHash, encryptedTransfer string,
		blockNumber int64,
	) (*mDBApp.BackupTransfer, error)
	GetBackupTransfer(condition string, value string) (*mDBApp.BackupTransfer, error)
	GetBackupTransferByRecipientAndTransferDoubleHash(
		recipient, transferDoubleHash string,
	) (*mDBApp.BackupTransfer, error)
	GetBackupTransfers(condition string, value interface{}) ([]*mDBApp.BackupTransfer, error)
	GetBackupTransfersByRecipient(
		recipient string,
		pagination mDBApp.PaginationOfListOfBackupTransfersInput,
		sorting mFL.Sorting, orderBy mFL.OrderBy,
		filters mFL.FiltersList,
	) (
		paginator *mDBApp.PaginationOfListOfBackupTransfers,
		listDBApp mDBApp.ListOfBackupTransfer,
		err error,
	)
}

type BackupTransactions interface {
	CreateBackupTransaction(
		sender, encryptedTxHash, encryptedTx, signature string,
		blockNumber int64,
	) (*mDBApp.BackupTransaction, error)
	GetBackupTransaction(condition string, value string) (*mDBApp.BackupTransaction, error)
	GetBackupTransactionBySenderAndTxDoubleHash(sender, txDoubleHash string) (*mDBApp.BackupTransaction, error)
	GetBackupTransactions(condition string, value interface{}) ([]*mDBApp.BackupTransaction, error)
	GetBackupTransactionsBySender(
		sender string,
		pagination mDBApp.PaginationOfListOfBackupTransactionsInput,
		sorting mFL.Sorting, orderBy mFL.OrderBy,
		filters mFL.FiltersList,
	) (
		paginator *mDBApp.PaginationOfListOfBackupTransactions,
		listDBApp mDBApp.ListOfBackupTransaction,
		err error,
	)
}

type BackupDeposits interface {
	CreateBackupDeposit(
		recipient, depositHash, encryptedDeposit string,
		blockNumber int64,
	) (*mDBApp.BackupDeposit, error)
	GetBackupDepositByRecipientAndDepositDoubleHash(
		recipient, depositDoubleHash string,
	) (*mDBApp.BackupDeposit, error)
	GetBackupDeposit(conditions []string, values []interface{}) (*mDBApp.BackupDeposit, error)
	GetBackupDeposits(condition string, value interface{}) ([]*mDBApp.BackupDeposit, error)
	GetBackupDepositsByRecipient(
		recipient string,
		pagination mDBApp.PaginationOfListOfBackupDepositsInput,
		sorting mFL.Sorting, orderBy mFL.OrderBy,
		filters mFL.FiltersList,
	) (
		paginator *mDBApp.PaginationOfListOfBackupDeposits,
		listDBApp mDBApp.ListOfBackupDeposit,
		err error,
	)
}

type BackupSenderProofs interface {
	CreateBackupSenderProof(
		lastBalanceProofBody, balanceTransitionProofBody []byte,
		enoughBalanceProofBodyHash string,
	) (*mDBApp.BackupSenderProof, error)
	GetBackupSenderProofsByHashes(enoughBalanceProofBodyHashes []string) ([]*mDBApp.BackupSenderProof, error)
}

type CtrlEventBlockNumbersJobs interface {
	CreateCtrlEventBlockNumbersJobs(eventName string) error
	CtrlEventBlockNumbersJobs(eventName string) (*mDBApp.CtrlEventBlockNumbersJobs, error)
}

type EventBlockNumbersErrors interface {
	UpsertEventBlockNumbersErrors(
		eventName string,
		blockNumber *uint256.Int,
		options []byte,
		updErr error,
	) error
	EventBlockNumbersErrors(
		eventName string,
		blockNumber *uint256.Int,
	) (*mDBApp.EventBlockNumbersErrors, error)
}

type EventBlockNumbersForValidityProver interface {
	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
}

type Senders interface {
	CreateSenders(
		address, publicKey string,
	) (*mDBApp.Sender, error)
	SenderByID(id string) (*mDBApp.Sender, error)
	SenderByAddress(address string) (*mDBApp.Sender, error)
	SenderByPublicKey(publicKey string) (*mDBApp.Sender, error)
}

type Accounts interface {
	CreateAccount(senderID string) (*mDBApp.Account, error)
	AccountBySenderID(senderID string) (*mDBApp.Account, error)
	AccountBySender(publicKey *intMaxAcc.PublicKey) (*mDBApp.Account, error)
	AccountByAccountID(accountID *uint256.Int) (*mDBApp.Account, error)
	ResetSequenceByAccounts() error
	DelAllAccounts() error
}

type BlockParticipants interface {
	CreateBlockParticipant(
		blockNumber uint32,
		senderId string,
	) (*mDBApp.BlockParticipant, error)
	BlockParticipantByBlockNumberAndSenderID(
		blockNumber uint32,
		senderId string,
	) (*mDBApp.BlockParticipant, error)
}

type BackupBalances interface {
	CreateBackupBalance(
		user, encryptedBalanceProof, encryptedBalanceData, signature string,
		encryptedTxs, encryptedTransfers, encryptedDeposits []string,
		blockNumber int64,
	) (*mDBApp.BackupBalance, error)
	GetBackupBalance(conditions []string, values []interface{}) (*mDBApp.BackupBalance, error)
	GetBackupBalances(condition string, value interface{}) ([]*mDBApp.BackupBalance, error)
	GetLatestBackupBalanceByUserAddress(user string, limit int64) ([]*mDBApp.BackupBalance, error)
}

type Deposits interface {
	CreateDeposit(depositLeaf intMaxTree.DepositLeaf, depositID uint32, sender string) (*mDBApp.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	UpdateSenderByDepositID(depositID uint32, sender string) error
	Deposit(ID string) (*mDBApp.Deposit, error)
	DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error)
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	DepositsListByDepositHash(depositHash ...common.Hash) ([]*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchNextDepositIndex() (uint32, error)
}

type BlockContents interface {
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
		l2BlockNumber *uint256.Int,
		l2BlockHash common.Hash,
	) (*mDBApp.BlockContentWithProof, error)
	BlockContentUpdDepositLeavesCounterByBlockNumber(
		blockNumber, depositLeavesCounter uint32,
	) error
	BlockContentIDByL2BlockNumber(l2BlockNumber string) (bcID string, err error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByBlockHash(blockHash string) (*mDBApp.BlockContentWithProof, error)
	BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	BlockContentListByTxRoot(txRoot ...common.Hash) ([]*mDBApp.BlockContentWithProof, error)
	ScanBlockHashAndSenders() (blockHashAndSendersMap map[uint32]mDBApp.BlockHashAndSenders, lastBlockNumber uint32, err error)
	CreateValidityProof(blockHash common.Hash, validityProof []byte) (*mDBApp.BlockProof, error)
	LastBlockValidityProof() (*mDBApp.BlockContentWithProof, error)
	LastBlockNumberGeneratedValidityProof() (uint32, error)
	LastPostedBlockNumber() (uint32, error)
}

type CtrlProcessingJobs interface {
	CreateCtrlProcessingJobs(name string, options json.RawMessage) error
	CtrlProcessingJobs(name string) (*mDBApp.CtrlProcessingJobs, error)
	CtrlProcessingJobsByMaskName(mask string) (*mDBApp.CtrlProcessingJobs, error)
	UpdatedAtOfCtrlProcessingJobByName(name string, updatedAt time.Time) (err error)
	DeleteCtrlProcessingJobByName(name string) (err error)
}

type GasPriceOracle interface {
	CreateGasPriceOracle(name string, value *uint256.Int) error
	GasPriceOracle(name string) (*mDBApp.GasPriceOracle, error)
}

type L2BatchIndex interface {
	CreateL2BatchIndex(batchIndex *uint256.Int) (err error)
	L2BatchIndex(batchIndex *uint256.Int) (*mDBApp.L2BatchIndex, error)
	UpdOptionsOfBatchIndex(batchIndex *uint256.Int, options json.RawMessage) (err error)
	UpdL1VerifiedBatchTxHashOfBatchIndex(batchIndex *uint256.Int, hash string) (err error)
}

type RelationshipL2BatchIndexAndBlockContent interface {
	CreateRelationshipL2BatchIndexAndBlockContentID(
		batchIndex *uint256.Int,
		blockContentID string,
	) (err error)
	RelationshipL2BatchIndexAndBlockContentsByBlockContentID(
		blockContentID string,
	) (*mDBApp.RelationshipL2BatchIndexBlockContents, error)
}

type EthereumCounterparties interface {
	CreateEthereumCounterparty(
		address string,
	) (*mDBApp.EthereumCounterparty, error)
	EthereumCounterpartyByAddress(address string) (*mDBApp.EthereumCounterparty, error)
}
