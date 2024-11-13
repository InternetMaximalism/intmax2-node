package db_app

import (
	"context"
	"encoding/json"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	mFL "intmax2-node/internal/sql_filter/models"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	"intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

type SQLDb interface {
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
	) (*models.Block, error)
	Block(proposalBlockID string) (*models.Block, error)
	BlockByTxRoot(txRoot string) (*models.Block, error)
	UpdateBlockStatus(proposalBlockID string, blockHash string, blockNumber uint32) error
	GetUnprocessedBlocks() ([]*models.Block, error)
}

type Tokens interface {
	CreateToken(
		tokenIndex, tokenAddress string,
		tokenID *uint256.Int,
	) (*models.Token, error)
	TokenByIndex(tokenIndex string) (*models.Token, error)
	TokenByTokenInfo(tokenAddress, tokenID string) (*models.Token, error)
}

type Signatures interface {
	CreateSignature(signature, proposalBlockID string) (*models.Signature, error)
	SignatureByID(signatureID string) (*models.Signature, error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash, signatureID string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
		txTreeRoot string,
		proposalBlockID string,
	) (*models.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*models.TxMerkleProofs, error)
	TxMerkleProofsByTxHash(txHash string) (*models.TxMerkleProofs, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*models.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*models.EventBlockNumber, error)
	EventBlockNumbersByEventNames(eventNames []string) ([]*models.EventBlockNumber, error)
}

type EventBlockNumbersForValidityProver interface {
	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*models.EventBlockNumberForValidityProver, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*models.EventBlockNumberForValidityProver, error)
}

type Balances interface {
	CreateBalance(userAddress, tokenAddress, balance string) (*models.Balance, error)
	UpdateBalanceByID(balanceID, balance string) (*models.Balance, error)
	BalanceByID(id string) (*models.Balance, error)
	BalanceByUserAndTokenIndex(userAddress, tokenIndex string) (*models.Balance, error)
	BalanceByUserAndTokenInfo(userAddress, tokenAddress, tokenID string) (*models.Balance, error)
}

type Withdrawals interface {
	CreateWithdrawal(
		id string,
		transferData *models.TransferData,
		transferMerkleProof *models.TransferMerkleProof,
		transaction *models.Transaction,
		txMerkleProof *models.TxMerkleProof,
		transferHash string,
		blockNumber int64,
		blockHash string,
		enoughBalanceProof *models.EnoughBalanceProof,
	) (*models.Withdrawal, error)
	UpdateWithdrawalsStatus(ids []string, status models.WithdrawalStatus) error
	WithdrawalByID(id string) (*models.Withdrawal, error)
	WithdrawalsByHashes(transferHashes []string) (*[]models.Withdrawal, error)
	WithdrawalsByStatus(status models.WithdrawalStatus, limit *int) (*[]models.Withdrawal, error)
}

type BackupTransfers interface {
	CreateBackupTransfer(
		recipient, encryptedTransferHash, encryptedTransfer string,
		blockNumber int64,
	) (*models.BackupTransfer, error)
	GetBackupTransfer(condition string, value string) (*models.BackupTransfer, error)
	GetBackupTransferByRecipientAndTransferDoubleHash(
		recipient, transferDoubleHash string,
	) (*models.BackupTransfer, error)
	GetBackupTransfers(condition string, value interface{}) ([]*models.BackupTransfer, error)
	GetBackupTransfersByRecipient(
		recipient string,
		pagination models.PaginationOfListOfBackupTransfersInput,
		sorting mFL.Sorting, orderBy mFL.OrderBy,
		filters mFL.FiltersList,
	) (
		paginator *models.PaginationOfListOfBackupTransfers,
		listDBApp models.ListOfBackupTransfer,
		err error,
	)
}

type BackupTransactions interface {
	CreateBackupTransaction(
		sender, encryptedTxHash, encryptedTx, signature string,
		blockNumber int64,
	) (*models.BackupTransaction, error)
	GetBackupTransaction(condition string, value string) (*models.BackupTransaction, error)
	GetBackupTransactionBySenderAndTxDoubleHash(sender, txDoubleHash string) (*models.BackupTransaction, error)
	GetBackupTransactions(condition string, value interface{}) ([]*models.BackupTransaction, error)
	GetBackupTransactionsBySender(
		sender string,
		pagination models.PaginationOfListOfBackupTransactionsInput,
		sorting mFL.Sorting, orderBy mFL.OrderBy,
		filters mFL.FiltersList,
	) (
		paginator *models.PaginationOfListOfBackupTransactions,
		listDBApp models.ListOfBackupTransaction,
		err error,
	)
}

type BackupDeposits interface {
	CreateBackupDeposit(
		recipient, depositHash, encryptedDeposit string,
		blockNumber int64,
	) (*models.BackupDeposit, error)
	GetBackupDepositByRecipientAndDepositDoubleHash(
		recipient, depositDoubleHash string,
	) (*models.BackupDeposit, error)
	GetBackupDeposit(conditions []string, values []interface{}) (*models.BackupDeposit, error)
	GetBackupDeposits(condition string, value interface{}) ([]*models.BackupDeposit, error)
	GetBackupDepositsByRecipient(
		recipient string,
		pagination models.PaginationOfListOfBackupDepositsInput,
		sorting mFL.Sorting, orderBy mFL.OrderBy,
		filters mFL.FiltersList,
	) (
		paginator *models.PaginationOfListOfBackupDeposits,
		listDBApp models.ListOfBackupDeposit,
		err error,
	)
}

type BackupSenderProofs interface {
	CreateBackupSenderProof(
		lastBalanceProofBody, balanceTransitionProofBody []byte,
		enoughBalanceProofBodyHash string,
	) (*models.BackupSenderProof, error)
	GetBackupSenderProofsByHashes(enoughBalanceProofBodyHashes []string) ([]*models.BackupSenderProof, error)
}

type CtrlEventBlockNumbersJobs interface {
	CreateCtrlEventBlockNumbersJobs(eventName string) error
	CtrlEventBlockNumbersJobs(eventName string) (*models.CtrlEventBlockNumbersJobs, error)
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
	) (*models.EventBlockNumbersErrors, error)
}

type Senders interface {
	CreateSenders(
		address, publicKey string,
	) (*models.Sender, error)
	SenderByID(id string) (*models.Sender, error)
	SenderByAddress(address string) (*models.Sender, error)
	SenderByPublicKey(publicKey string) (*models.Sender, error)
}

type Accounts interface {
	CreateAccount(senderID string) (*models.Account, error)
	AccountBySenderID(senderID string) (*models.Account, error)
	AccountBySender(publicKey *intMaxAcc.PublicKey) (*models.Account, error)
	AccountByAccountID(accountID *uint256.Int) (*models.Account, error)
	ResetSequenceByAccounts() error
	DelAllAccounts() error
}

type BlockParticipants interface {
	CreateBlockParticipant(
		blockNumber uint32,
		senderId string,
	) (*models.BlockParticipant, error)
	BlockParticipantByBlockNumberAndSenderID(
		blockNumber uint32,
		senderId string,
	) (*models.BlockParticipant, error)
}

type BackupBalances interface {
	CreateBackupBalance(
		user, encryptedBalanceProof, encryptedBalanceData, signature string,
		encryptedTxs, encryptedTransfers, encryptedDeposits []string,
		blockNumber int64,
	) (*models.BackupBalance, error)
	GetBackupBalance(conditions []string, values []interface{}) (*models.BackupBalance, error)
	GetBackupBalances(condition string, value interface{}) ([]*models.BackupBalance, error)
	GetLatestBackupBalanceByUserAddress(user string, limit int64) ([]*models.BackupBalance, error)
}

type Deposits interface {
	CreateDeposit(depositLeaf intMaxTree.DepositLeaf, depositID uint32, sender string) (*models.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	UpdateSenderByDepositID(depositID uint32, sender string) error
	Deposit(ID string) (*models.Deposit, error)
	DepositByDepositID(depositID uint32) (*models.Deposit, error)
	DepositByDepositHash(depositHash common.Hash) (*models.Deposit, error)
	DepositsListByDepositHash(depositHash ...common.Hash) ([]*models.Deposit, error)
	ScanDeposits() ([]*models.Deposit, error)
	FetchNextDepositIndex() (uint32, error)
}

type BlockContents interface {
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
		l2BlockNumber *uint256.Int,
		l2BlockHash common.Hash,
	) (*models.BlockContentWithProof, error)
	BlockContentUpdDepositLeavesCounterByBlockNumber(
		blockNumber, depositLeavesCounter uint32,
	) error
	BlockContentIDByL2BlockNumber(l2BlockNumber string) (bcID string, err error)
	BlockContentByBlockNumber(blockNumber uint32) (*models.BlockContentWithProof, error)
	BlockContentByBlockHash(blockHash string) (*models.BlockContentWithProof, error)
	BlockContentByTxRoot(txRoot common.Hash) (*models.BlockContentWithProof, error)
	BlockContentListByTxRoot(txRoot ...common.Hash) ([]*models.BlockContentWithProof, error)
	ScanBlockHashAndSenders() (blockHashAndSendersMap map[uint32]models.BlockHashAndSenders, lastBlockNumber uint32, err error)
	CreateValidityProof(blockHash common.Hash, validityProof []byte) (*models.BlockProof, error)
	LastBlockValidityProof() (*models.BlockContentWithProof, error)
	LastBlockNumberGeneratedValidityProof() (uint32, error)
	LastPostedBlockNumber() (uint32, error)
}

type CtrlProcessingJobs interface {
	CreateCtrlProcessingJobs(name string, options json.RawMessage) error
	CtrlProcessingJobs(name string) (*models.CtrlProcessingJobs, error)
	CtrlProcessingJobsByMaskName(mask string) (*models.CtrlProcessingJobs, error)
	UpdatedAtOfCtrlProcessingJobByName(name string, updatedAt time.Time) (err error)
	DeleteCtrlProcessingJobByName(name string) (err error)
}

type GasPriceOracle interface {
	CreateGasPriceOracle(name string, value *uint256.Int) error
	GasPriceOracle(name string) (*models.GasPriceOracle, error)
}

type L2BatchIndex interface {
	CreateL2BatchIndex(batchIndex *uint256.Int) (err error)
	L2BatchIndex(batchIndex *uint256.Int) (*models.L2BatchIndex, error)
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
	) (*models.RelationshipL2BatchIndexBlockContents, error)
}

type EthereumCounterparties interface {
	CreateEthereumCounterparty(
		address string,
	) (*models.EthereumCounterparty, error)
	EthereumCounterpartyByAddress(address string) (*models.EthereumCounterparty, error)
}
