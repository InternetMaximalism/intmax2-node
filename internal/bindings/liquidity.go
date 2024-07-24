// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// DepositLibDeposit is an auto generated low-level Go binding around an user-defined struct.
type DepositLibDeposit struct {
	RecipientSaltHash [32]byte
	TokenIndex        uint32
	Amount            *big.Int
}

// ITokenDataTokenInfo is an auto generated low-level Go binding around an user-defined struct.
type ITokenDataTokenInfo struct {
	TokenType    uint8
	TokenAddress common.Address
	TokenId      *big.Int
}

// LiquidityMetaData contains all meta data concerning the Liquidity contract.
var LiquidityMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"AccessControlBadConfirmation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"neededRole\",\"type\":\"bytes32\"}],\"name\":\"AccessControlUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"AddressInsufficientBalance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"front\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rear\",\"type\":\"uint256\"}],\"name\":\"DepositIdIsNotInQueue\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidAmount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"depositDataHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"calculatedHash\",\"type\":\"bytes32\"}],\"name\":\"InvalidDepositHash\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidTokenAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidTokenInfo\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidValue\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidWithdrawalAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlyRecipientCanCancelDeposit\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReentrancyGuardReentrantCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderIsNotScrollMessenger\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"upToDepositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastDepositId\",\"type\":\"uint256\"}],\"name\":\"TriedAnalyzeNotExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"upToDepositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastAnalyzedDepositId\",\"type\":\"uint256\"}],\"name\":\"TriedCollectDepositsNotAnalyzedYet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"rejectIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastAnalyzedDepositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"upToDepositId\",\"type\":\"uint256\"}],\"name\":\"TriedToRejectOutOfRange\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"upToDepositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lastAnalyzedDepositId\",\"type\":\"uint256\"}],\"name\":\"UpToDepositIdIsTooOld\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawalAddressNotSet\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"withdrawalHash\",\"type\":\"bytes32\"}],\"name\":\"WithdrawalNotFound\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"lastProcessedClaimableWithdrawalId\",\"type\":\"uint256\"}],\"name\":\"ClaimableWithdrawalsProcessed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"}],\"name\":\"DepositCanceled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"recipientSaltHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"requestedAt\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"lastAnalyzedDepositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"rejectedIndices\",\"type\":\"uint256[]\"}],\"name\":\"DepositsAnalyzed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"lastRelayedDepositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"DepositsRelayed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"newGasLimit\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"messageNonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"DepositsReplayed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"lastProcessedDirectWithdrawalId\",\"type\":\"uint256\"}],\"name\":\"DirectWithdrawalsProcessed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"withdrawalHash\",\"type\":\"bytes32\"}],\"name\":\"WithdrawalClaimable\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ANALYZER\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"initialERC20Tokens\",\"type\":\"address[]\"}],\"name\":\"__TokenData_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"upToDepositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256[]\",\"name\":\"rejectDepositIndices\",\"type\":\"uint256[]\"}],\"name\":\"analyzeDeposits\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"depositId\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"recipientSaltHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structDepositLib.Deposit\",\"name\":\"deposit\",\"type\":\"tuple\"}],\"name\":\"cancelDeposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structWithdrawalLib.Withdrawal[]\",\"name\":\"withdrawals\",\"type\":\"tuple[]\"}],\"name\":\"claimWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipientSaltHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"depositERC1155\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipientSaltHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"depositERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipientSaltHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"depositERC721\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"recipientSaltHash\",\"type\":\"bytes32\"}],\"name\":\"depositETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"}],\"name\":\"getTokenInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"enumITokenData.TokenType\",\"name\":\"tokenType\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"internalType\":\"structITokenData.TokenInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_l1ScrollMessenger\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_rollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_withdrawal\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_analyzer\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"initialERC20Tokens\",\"type\":\"address[]\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastProcessedClaimableWithdrawalId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastProcessedDirectWithdrawalId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_lastProcessedClaimableWithdrawalId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"withdrawalHahes\",\"type\":\"bytes32[]\"}],\"name\":\"processClaimableWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_lastProcessedDirectWithdrawalId\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structWithdrawalLib.Withdrawal[]\",\"name\":\"withdrawals\",\"type\":\"tuple[]\"}],\"name\":\"processDirectWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_lastProcessedDirectWithdrawalId\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"internalType\":\"structWithdrawalLib.Withdrawal[]\",\"name\":\"withdrawals\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_lastProcessedClaimableWithdrawalId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"withdrawalHahes\",\"type\":\"bytes32[]\"}],\"name\":\"processWithdrawals\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"upToDepositId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"}],\"name\":\"relayDeposits\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"callerConfirmation\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"newGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"messageNonce\",\"type\":\"uint256\"}],\"name\":\"replayDeposits\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// LiquidityABI is the input ABI used to generate the binding from.
// Deprecated: Use LiquidityMetaData.ABI instead.
var LiquidityABI = LiquidityMetaData.ABI

// Liquidity is an auto generated Go binding around an Ethereum contract.
type Liquidity struct {
	LiquidityCaller     // Read-only binding to the contract
	LiquidityTransactor // Write-only binding to the contract
	LiquidityFilterer   // Log filterer for contract events
}

// LiquidityCaller is an auto generated read-only Go binding around an Ethereum contract.
type LiquidityCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LiquidityTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LiquidityTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LiquidityFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LiquidityFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LiquiditySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LiquiditySession struct {
	Contract     *Liquidity        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LiquidityCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LiquidityCallerSession struct {
	Contract *LiquidityCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// LiquidityTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LiquidityTransactorSession struct {
	Contract     *LiquidityTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// LiquidityRaw is an auto generated low-level Go binding around an Ethereum contract.
type LiquidityRaw struct {
	Contract *Liquidity // Generic contract binding to access the raw methods on
}

// LiquidityCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LiquidityCallerRaw struct {
	Contract *LiquidityCaller // Generic read-only contract binding to access the raw methods on
}

// LiquidityTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LiquidityTransactorRaw struct {
	Contract *LiquidityTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLiquidity creates a new instance of Liquidity, bound to a specific deployed contract.
func NewLiquidity(address common.Address, backend bind.ContractBackend) (*Liquidity, error) {
	contract, err := bindLiquidity(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Liquidity{LiquidityCaller: LiquidityCaller{contract: contract}, LiquidityTransactor: LiquidityTransactor{contract: contract}, LiquidityFilterer: LiquidityFilterer{contract: contract}}, nil
}

// NewLiquidityCaller creates a new read-only instance of Liquidity, bound to a specific deployed contract.
func NewLiquidityCaller(address common.Address, caller bind.ContractCaller) (*LiquidityCaller, error) {
	contract, err := bindLiquidity(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LiquidityCaller{contract: contract}, nil
}

// NewLiquidityTransactor creates a new write-only instance of Liquidity, bound to a specific deployed contract.
func NewLiquidityTransactor(address common.Address, transactor bind.ContractTransactor) (*LiquidityTransactor, error) {
	contract, err := bindLiquidity(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LiquidityTransactor{contract: contract}, nil
}

// NewLiquidityFilterer creates a new log filterer instance of Liquidity, bound to a specific deployed contract.
func NewLiquidityFilterer(address common.Address, filterer bind.ContractFilterer) (*LiquidityFilterer, error) {
	contract, err := bindLiquidity(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LiquidityFilterer{contract: contract}, nil
}

// bindLiquidity binds a generic wrapper to an already deployed contract.
func bindLiquidity(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := LiquidityMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Liquidity *LiquidityRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Liquidity.Contract.LiquidityCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Liquidity *LiquidityRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Liquidity.Contract.LiquidityTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Liquidity *LiquidityRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Liquidity.Contract.LiquidityTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Liquidity *LiquidityCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Liquidity.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Liquidity *LiquidityTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Liquidity.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Liquidity *LiquidityTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Liquidity.Contract.contract.Transact(opts, method, params...)
}

// ANALYZER is a free data retrieval call binding the contract method 0x0c0b59b9.
//
// Solidity: function ANALYZER() view returns(bytes32)
func (_Liquidity *LiquidityCaller) ANALYZER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "ANALYZER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ANALYZER is a free data retrieval call binding the contract method 0x0c0b59b9.
//
// Solidity: function ANALYZER() view returns(bytes32)
func (_Liquidity *LiquiditySession) ANALYZER() ([32]byte, error) {
	return _Liquidity.Contract.ANALYZER(&_Liquidity.CallOpts)
}

// ANALYZER is a free data retrieval call binding the contract method 0x0c0b59b9.
//
// Solidity: function ANALYZER() view returns(bytes32)
func (_Liquidity *LiquidityCallerSession) ANALYZER() ([32]byte, error) {
	return _Liquidity.Contract.ANALYZER(&_Liquidity.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Liquidity *LiquidityCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Liquidity *LiquiditySession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Liquidity.Contract.DEFAULTADMINROLE(&_Liquidity.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Liquidity *LiquidityCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Liquidity.Contract.DEFAULTADMINROLE(&_Liquidity.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Liquidity *LiquidityCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Liquidity *LiquiditySession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Liquidity.Contract.UPGRADEINTERFACEVERSION(&_Liquidity.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Liquidity *LiquidityCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Liquidity.Contract.UPGRADEINTERFACEVERSION(&_Liquidity.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Liquidity *LiquidityCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Liquidity *LiquiditySession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Liquidity.Contract.GetRoleAdmin(&_Liquidity.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Liquidity *LiquidityCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Liquidity.Contract.GetRoleAdmin(&_Liquidity.CallOpts, role)
}

// GetTokenInfo is a free data retrieval call binding the contract method 0xcbd90cd0.
//
// Solidity: function getTokenInfo(uint32 tokenIndex) view returns((uint8,address,uint256))
func (_Liquidity *LiquidityCaller) GetTokenInfo(opts *bind.CallOpts, tokenIndex uint32) (ITokenDataTokenInfo, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getTokenInfo", tokenIndex)

	if err != nil {
		return *new(ITokenDataTokenInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(ITokenDataTokenInfo)).(*ITokenDataTokenInfo)

	return out0, err

}

// GetTokenInfo is a free data retrieval call binding the contract method 0xcbd90cd0.
//
// Solidity: function getTokenInfo(uint32 tokenIndex) view returns((uint8,address,uint256))
func (_Liquidity *LiquiditySession) GetTokenInfo(tokenIndex uint32) (ITokenDataTokenInfo, error) {
	return _Liquidity.Contract.GetTokenInfo(&_Liquidity.CallOpts, tokenIndex)
}

// GetTokenInfo is a free data retrieval call binding the contract method 0xcbd90cd0.
//
// Solidity: function getTokenInfo(uint32 tokenIndex) view returns((uint8,address,uint256))
func (_Liquidity *LiquidityCallerSession) GetTokenInfo(tokenIndex uint32) (ITokenDataTokenInfo, error) {
	return _Liquidity.Contract.GetTokenInfo(&_Liquidity.CallOpts, tokenIndex)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Liquidity *LiquidityCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Liquidity *LiquiditySession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Liquidity.Contract.HasRole(&_Liquidity.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Liquidity *LiquidityCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Liquidity.Contract.HasRole(&_Liquidity.CallOpts, role, account)
}

// LastProcessedClaimableWithdrawalId is a free data retrieval call binding the contract method 0x32ddb179.
//
// Solidity: function lastProcessedClaimableWithdrawalId() view returns(uint256)
func (_Liquidity *LiquidityCaller) LastProcessedClaimableWithdrawalId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "lastProcessedClaimableWithdrawalId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastProcessedClaimableWithdrawalId is a free data retrieval call binding the contract method 0x32ddb179.
//
// Solidity: function lastProcessedClaimableWithdrawalId() view returns(uint256)
func (_Liquidity *LiquiditySession) LastProcessedClaimableWithdrawalId() (*big.Int, error) {
	return _Liquidity.Contract.LastProcessedClaimableWithdrawalId(&_Liquidity.CallOpts)
}

// LastProcessedClaimableWithdrawalId is a free data retrieval call binding the contract method 0x32ddb179.
//
// Solidity: function lastProcessedClaimableWithdrawalId() view returns(uint256)
func (_Liquidity *LiquidityCallerSession) LastProcessedClaimableWithdrawalId() (*big.Int, error) {
	return _Liquidity.Contract.LastProcessedClaimableWithdrawalId(&_Liquidity.CallOpts)
}

// LastProcessedDirectWithdrawalId is a free data retrieval call binding the contract method 0xdebc9946.
//
// Solidity: function lastProcessedDirectWithdrawalId() view returns(uint256)
func (_Liquidity *LiquidityCaller) LastProcessedDirectWithdrawalId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "lastProcessedDirectWithdrawalId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastProcessedDirectWithdrawalId is a free data retrieval call binding the contract method 0xdebc9946.
//
// Solidity: function lastProcessedDirectWithdrawalId() view returns(uint256)
func (_Liquidity *LiquiditySession) LastProcessedDirectWithdrawalId() (*big.Int, error) {
	return _Liquidity.Contract.LastProcessedDirectWithdrawalId(&_Liquidity.CallOpts)
}

// LastProcessedDirectWithdrawalId is a free data retrieval call binding the contract method 0xdebc9946.
//
// Solidity: function lastProcessedDirectWithdrawalId() view returns(uint256)
func (_Liquidity *LiquidityCallerSession) LastProcessedDirectWithdrawalId() (*big.Int, error) {
	return _Liquidity.Contract.LastProcessedDirectWithdrawalId(&_Liquidity.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Liquidity *LiquidityCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Liquidity *LiquiditySession) ProxiableUUID() ([32]byte, error) {
	return _Liquidity.Contract.ProxiableUUID(&_Liquidity.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Liquidity *LiquidityCallerSession) ProxiableUUID() ([32]byte, error) {
	return _Liquidity.Contract.ProxiableUUID(&_Liquidity.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Liquidity *LiquidityCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Liquidity *LiquiditySession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Liquidity.Contract.SupportsInterface(&_Liquidity.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Liquidity *LiquidityCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Liquidity.Contract.SupportsInterface(&_Liquidity.CallOpts, interfaceId)
}

// TokenDataInit is a paid mutator transaction binding the contract method 0x20c7b8d0.
//
// Solidity: function __TokenData_init(address[] initialERC20Tokens) returns()
func (_Liquidity *LiquidityTransactor) TokenDataInit(opts *bind.TransactOpts, initialERC20Tokens []common.Address) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "__TokenData_init", initialERC20Tokens)
}

// TokenDataInit is a paid mutator transaction binding the contract method 0x20c7b8d0.
//
// Solidity: function __TokenData_init(address[] initialERC20Tokens) returns()
func (_Liquidity *LiquiditySession) TokenDataInit(initialERC20Tokens []common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.TokenDataInit(&_Liquidity.TransactOpts, initialERC20Tokens)
}

// TokenDataInit is a paid mutator transaction binding the contract method 0x20c7b8d0.
//
// Solidity: function __TokenData_init(address[] initialERC20Tokens) returns()
func (_Liquidity *LiquidityTransactorSession) TokenDataInit(initialERC20Tokens []common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.TokenDataInit(&_Liquidity.TransactOpts, initialERC20Tokens)
}

// AnalyzeDeposits is a paid mutator transaction binding the contract method 0xd9013ab8.
//
// Solidity: function analyzeDeposits(uint256 upToDepositId, uint256[] rejectDepositIndices) returns()
func (_Liquidity *LiquidityTransactor) AnalyzeDeposits(opts *bind.TransactOpts, upToDepositId *big.Int, rejectDepositIndices []*big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "analyzeDeposits", upToDepositId, rejectDepositIndices)
}

// AnalyzeDeposits is a paid mutator transaction binding the contract method 0xd9013ab8.
//
// Solidity: function analyzeDeposits(uint256 upToDepositId, uint256[] rejectDepositIndices) returns()
func (_Liquidity *LiquiditySession) AnalyzeDeposits(upToDepositId *big.Int, rejectDepositIndices []*big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.AnalyzeDeposits(&_Liquidity.TransactOpts, upToDepositId, rejectDepositIndices)
}

// AnalyzeDeposits is a paid mutator transaction binding the contract method 0xd9013ab8.
//
// Solidity: function analyzeDeposits(uint256 upToDepositId, uint256[] rejectDepositIndices) returns()
func (_Liquidity *LiquidityTransactorSession) AnalyzeDeposits(upToDepositId *big.Int, rejectDepositIndices []*big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.AnalyzeDeposits(&_Liquidity.TransactOpts, upToDepositId, rejectDepositIndices)
}

// CancelDeposit is a paid mutator transaction binding the contract method 0x6ef5475b.
//
// Solidity: function cancelDeposit(uint256 depositId, (bytes32,uint32,uint256) deposit) returns()
func (_Liquidity *LiquidityTransactor) CancelDeposit(opts *bind.TransactOpts, depositId *big.Int, deposit DepositLibDeposit) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "cancelDeposit", depositId, deposit)
}

// CancelDeposit is a paid mutator transaction binding the contract method 0x6ef5475b.
//
// Solidity: function cancelDeposit(uint256 depositId, (bytes32,uint32,uint256) deposit) returns()
func (_Liquidity *LiquiditySession) CancelDeposit(depositId *big.Int, deposit DepositLibDeposit) (*types.Transaction, error) {
	return _Liquidity.Contract.CancelDeposit(&_Liquidity.TransactOpts, depositId, deposit)
}

// CancelDeposit is a paid mutator transaction binding the contract method 0x6ef5475b.
//
// Solidity: function cancelDeposit(uint256 depositId, (bytes32,uint32,uint256) deposit) returns()
func (_Liquidity *LiquidityTransactorSession) CancelDeposit(depositId *big.Int, deposit DepositLibDeposit) (*types.Transaction, error) {
	return _Liquidity.Contract.CancelDeposit(&_Liquidity.TransactOpts, depositId, deposit)
}

// ClaimWithdrawals is a paid mutator transaction binding the contract method 0x729f67be.
//
// Solidity: function claimWithdrawals((address,uint32,uint256,uint256)[] withdrawals) returns()
func (_Liquidity *LiquidityTransactor) ClaimWithdrawals(opts *bind.TransactOpts, withdrawals []WithdrawalLibWithdrawal) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "claimWithdrawals", withdrawals)
}

// ClaimWithdrawals is a paid mutator transaction binding the contract method 0x729f67be.
//
// Solidity: function claimWithdrawals((address,uint32,uint256,uint256)[] withdrawals) returns()
func (_Liquidity *LiquiditySession) ClaimWithdrawals(withdrawals []WithdrawalLibWithdrawal) (*types.Transaction, error) {
	return _Liquidity.Contract.ClaimWithdrawals(&_Liquidity.TransactOpts, withdrawals)
}

// ClaimWithdrawals is a paid mutator transaction binding the contract method 0x729f67be.
//
// Solidity: function claimWithdrawals((address,uint32,uint256,uint256)[] withdrawals) returns()
func (_Liquidity *LiquidityTransactorSession) ClaimWithdrawals(withdrawals []WithdrawalLibWithdrawal) (*types.Transaction, error) {
	return _Liquidity.Contract.ClaimWithdrawals(&_Liquidity.TransactOpts, withdrawals)
}

// DepositERC1155 is a paid mutator transaction binding the contract method 0x26fc02dd.
//
// Solidity: function depositERC1155(address tokenAddress, bytes32 recipientSaltHash, uint256 tokenId, uint256 amount) returns()
func (_Liquidity *LiquidityTransactor) DepositERC1155(opts *bind.TransactOpts, tokenAddress common.Address, recipientSaltHash [32]byte, tokenId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "depositERC1155", tokenAddress, recipientSaltHash, tokenId, amount)
}

// DepositERC1155 is a paid mutator transaction binding the contract method 0x26fc02dd.
//
// Solidity: function depositERC1155(address tokenAddress, bytes32 recipientSaltHash, uint256 tokenId, uint256 amount) returns()
func (_Liquidity *LiquiditySession) DepositERC1155(tokenAddress common.Address, recipientSaltHash [32]byte, tokenId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC1155(&_Liquidity.TransactOpts, tokenAddress, recipientSaltHash, tokenId, amount)
}

// DepositERC1155 is a paid mutator transaction binding the contract method 0x26fc02dd.
//
// Solidity: function depositERC1155(address tokenAddress, bytes32 recipientSaltHash, uint256 tokenId, uint256 amount) returns()
func (_Liquidity *LiquidityTransactorSession) DepositERC1155(tokenAddress common.Address, recipientSaltHash [32]byte, tokenId *big.Int, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC1155(&_Liquidity.TransactOpts, tokenAddress, recipientSaltHash, tokenId, amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x788c2ecf.
//
// Solidity: function depositERC20(address tokenAddress, bytes32 recipientSaltHash, uint256 amount) returns()
func (_Liquidity *LiquidityTransactor) DepositERC20(opts *bind.TransactOpts, tokenAddress common.Address, recipientSaltHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "depositERC20", tokenAddress, recipientSaltHash, amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x788c2ecf.
//
// Solidity: function depositERC20(address tokenAddress, bytes32 recipientSaltHash, uint256 amount) returns()
func (_Liquidity *LiquiditySession) DepositERC20(tokenAddress common.Address, recipientSaltHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC20(&_Liquidity.TransactOpts, tokenAddress, recipientSaltHash, amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x788c2ecf.
//
// Solidity: function depositERC20(address tokenAddress, bytes32 recipientSaltHash, uint256 amount) returns()
func (_Liquidity *LiquidityTransactorSession) DepositERC20(tokenAddress common.Address, recipientSaltHash [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC20(&_Liquidity.TransactOpts, tokenAddress, recipientSaltHash, amount)
}

// DepositERC721 is a paid mutator transaction binding the contract method 0x9f401f58.
//
// Solidity: function depositERC721(address tokenAddress, bytes32 recipientSaltHash, uint256 tokenId) returns()
func (_Liquidity *LiquidityTransactor) DepositERC721(opts *bind.TransactOpts, tokenAddress common.Address, recipientSaltHash [32]byte, tokenId *big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "depositERC721", tokenAddress, recipientSaltHash, tokenId)
}

// DepositERC721 is a paid mutator transaction binding the contract method 0x9f401f58.
//
// Solidity: function depositERC721(address tokenAddress, bytes32 recipientSaltHash, uint256 tokenId) returns()
func (_Liquidity *LiquiditySession) DepositERC721(tokenAddress common.Address, recipientSaltHash [32]byte, tokenId *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC721(&_Liquidity.TransactOpts, tokenAddress, recipientSaltHash, tokenId)
}

// DepositERC721 is a paid mutator transaction binding the contract method 0x9f401f58.
//
// Solidity: function depositERC721(address tokenAddress, bytes32 recipientSaltHash, uint256 tokenId) returns()
func (_Liquidity *LiquidityTransactorSession) DepositERC721(tokenAddress common.Address, recipientSaltHash [32]byte, tokenId *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC721(&_Liquidity.TransactOpts, tokenAddress, recipientSaltHash, tokenId)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 recipientSaltHash) payable returns()
func (_Liquidity *LiquidityTransactor) DepositETH(opts *bind.TransactOpts, recipientSaltHash [32]byte) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "depositETH", recipientSaltHash)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 recipientSaltHash) payable returns()
func (_Liquidity *LiquiditySession) DepositETH(recipientSaltHash [32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositETH(&_Liquidity.TransactOpts, recipientSaltHash)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 recipientSaltHash) payable returns()
func (_Liquidity *LiquidityTransactorSession) DepositETH(recipientSaltHash [32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositETH(&_Liquidity.TransactOpts, recipientSaltHash)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Liquidity *LiquidityTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Liquidity *LiquiditySession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.GrantRole(&_Liquidity.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Liquidity *LiquidityTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.GrantRole(&_Liquidity.TransactOpts, role, account)
}

// Initialize is a paid mutator transaction binding the contract method 0xf8453e7c.
//
// Solidity: function initialize(address _l1ScrollMessenger, address _rollup, address _withdrawal, address _analyzer, address[] initialERC20Tokens) returns()
func (_Liquidity *LiquidityTransactor) Initialize(opts *bind.TransactOpts, _l1ScrollMessenger common.Address, _rollup common.Address, _withdrawal common.Address, _analyzer common.Address, initialERC20Tokens []common.Address) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "initialize", _l1ScrollMessenger, _rollup, _withdrawal, _analyzer, initialERC20Tokens)
}

// Initialize is a paid mutator transaction binding the contract method 0xf8453e7c.
//
// Solidity: function initialize(address _l1ScrollMessenger, address _rollup, address _withdrawal, address _analyzer, address[] initialERC20Tokens) returns()
func (_Liquidity *LiquiditySession) Initialize(_l1ScrollMessenger common.Address, _rollup common.Address, _withdrawal common.Address, _analyzer common.Address, initialERC20Tokens []common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.Initialize(&_Liquidity.TransactOpts, _l1ScrollMessenger, _rollup, _withdrawal, _analyzer, initialERC20Tokens)
}

// Initialize is a paid mutator transaction binding the contract method 0xf8453e7c.
//
// Solidity: function initialize(address _l1ScrollMessenger, address _rollup, address _withdrawal, address _analyzer, address[] initialERC20Tokens) returns()
func (_Liquidity *LiquidityTransactorSession) Initialize(_l1ScrollMessenger common.Address, _rollup common.Address, _withdrawal common.Address, _analyzer common.Address, initialERC20Tokens []common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.Initialize(&_Liquidity.TransactOpts, _l1ScrollMessenger, _rollup, _withdrawal, _analyzer, initialERC20Tokens)
}

// ProcessClaimableWithdrawals is a paid mutator transaction binding the contract method 0x9ef300e1.
//
// Solidity: function processClaimableWithdrawals(uint256 _lastProcessedClaimableWithdrawalId, bytes32[] withdrawalHahes) returns()
func (_Liquidity *LiquidityTransactor) ProcessClaimableWithdrawals(opts *bind.TransactOpts, _lastProcessedClaimableWithdrawalId *big.Int, withdrawalHahes [][32]byte) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "processClaimableWithdrawals", _lastProcessedClaimableWithdrawalId, withdrawalHahes)
}

// ProcessClaimableWithdrawals is a paid mutator transaction binding the contract method 0x9ef300e1.
//
// Solidity: function processClaimableWithdrawals(uint256 _lastProcessedClaimableWithdrawalId, bytes32[] withdrawalHahes) returns()
func (_Liquidity *LiquiditySession) ProcessClaimableWithdrawals(_lastProcessedClaimableWithdrawalId *big.Int, withdrawalHahes [][32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.ProcessClaimableWithdrawals(&_Liquidity.TransactOpts, _lastProcessedClaimableWithdrawalId, withdrawalHahes)
}

// ProcessClaimableWithdrawals is a paid mutator transaction binding the contract method 0x9ef300e1.
//
// Solidity: function processClaimableWithdrawals(uint256 _lastProcessedClaimableWithdrawalId, bytes32[] withdrawalHahes) returns()
func (_Liquidity *LiquidityTransactorSession) ProcessClaimableWithdrawals(_lastProcessedClaimableWithdrawalId *big.Int, withdrawalHahes [][32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.ProcessClaimableWithdrawals(&_Liquidity.TransactOpts, _lastProcessedClaimableWithdrawalId, withdrawalHahes)
}

// ProcessDirectWithdrawals is a paid mutator transaction binding the contract method 0x652531a5.
//
// Solidity: function processDirectWithdrawals(uint256 _lastProcessedDirectWithdrawalId, (address,uint32,uint256,uint256)[] withdrawals) returns()
func (_Liquidity *LiquidityTransactor) ProcessDirectWithdrawals(opts *bind.TransactOpts, _lastProcessedDirectWithdrawalId *big.Int, withdrawals []WithdrawalLibWithdrawal) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "processDirectWithdrawals", _lastProcessedDirectWithdrawalId, withdrawals)
}

// ProcessDirectWithdrawals is a paid mutator transaction binding the contract method 0x652531a5.
//
// Solidity: function processDirectWithdrawals(uint256 _lastProcessedDirectWithdrawalId, (address,uint32,uint256,uint256)[] withdrawals) returns()
func (_Liquidity *LiquiditySession) ProcessDirectWithdrawals(_lastProcessedDirectWithdrawalId *big.Int, withdrawals []WithdrawalLibWithdrawal) (*types.Transaction, error) {
	return _Liquidity.Contract.ProcessDirectWithdrawals(&_Liquidity.TransactOpts, _lastProcessedDirectWithdrawalId, withdrawals)
}

// ProcessDirectWithdrawals is a paid mutator transaction binding the contract method 0x652531a5.
//
// Solidity: function processDirectWithdrawals(uint256 _lastProcessedDirectWithdrawalId, (address,uint32,uint256,uint256)[] withdrawals) returns()
func (_Liquidity *LiquidityTransactorSession) ProcessDirectWithdrawals(_lastProcessedDirectWithdrawalId *big.Int, withdrawals []WithdrawalLibWithdrawal) (*types.Transaction, error) {
	return _Liquidity.Contract.ProcessDirectWithdrawals(&_Liquidity.TransactOpts, _lastProcessedDirectWithdrawalId, withdrawals)
}

// ProcessWithdrawals is a paid mutator transaction binding the contract method 0x088f0bdd.
//
// Solidity: function processWithdrawals(uint256 _lastProcessedDirectWithdrawalId, (address,uint32,uint256,uint256)[] withdrawals, uint256 _lastProcessedClaimableWithdrawalId, bytes32[] withdrawalHahes) returns()
func (_Liquidity *LiquidityTransactor) ProcessWithdrawals(opts *bind.TransactOpts, _lastProcessedDirectWithdrawalId *big.Int, withdrawals []WithdrawalLibWithdrawal, _lastProcessedClaimableWithdrawalId *big.Int, withdrawalHahes [][32]byte) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "processWithdrawals", _lastProcessedDirectWithdrawalId, withdrawals, _lastProcessedClaimableWithdrawalId, withdrawalHahes)
}

// ProcessWithdrawals is a paid mutator transaction binding the contract method 0x088f0bdd.
//
// Solidity: function processWithdrawals(uint256 _lastProcessedDirectWithdrawalId, (address,uint32,uint256,uint256)[] withdrawals, uint256 _lastProcessedClaimableWithdrawalId, bytes32[] withdrawalHahes) returns()
func (_Liquidity *LiquiditySession) ProcessWithdrawals(_lastProcessedDirectWithdrawalId *big.Int, withdrawals []WithdrawalLibWithdrawal, _lastProcessedClaimableWithdrawalId *big.Int, withdrawalHahes [][32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.ProcessWithdrawals(&_Liquidity.TransactOpts, _lastProcessedDirectWithdrawalId, withdrawals, _lastProcessedClaimableWithdrawalId, withdrawalHahes)
}

// ProcessWithdrawals is a paid mutator transaction binding the contract method 0x088f0bdd.
//
// Solidity: function processWithdrawals(uint256 _lastProcessedDirectWithdrawalId, (address,uint32,uint256,uint256)[] withdrawals, uint256 _lastProcessedClaimableWithdrawalId, bytes32[] withdrawalHahes) returns()
func (_Liquidity *LiquidityTransactorSession) ProcessWithdrawals(_lastProcessedDirectWithdrawalId *big.Int, withdrawals []WithdrawalLibWithdrawal, _lastProcessedClaimableWithdrawalId *big.Int, withdrawalHahes [][32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.ProcessWithdrawals(&_Liquidity.TransactOpts, _lastProcessedDirectWithdrawalId, withdrawals, _lastProcessedClaimableWithdrawalId, withdrawalHahes)
}

// RelayDeposits is a paid mutator transaction binding the contract method 0x76918c0d.
//
// Solidity: function relayDeposits(uint256 upToDepositId, uint256 gasLimit) payable returns()
func (_Liquidity *LiquidityTransactor) RelayDeposits(opts *bind.TransactOpts, upToDepositId *big.Int, gasLimit *big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "relayDeposits", upToDepositId, gasLimit)
}

// RelayDeposits is a paid mutator transaction binding the contract method 0x76918c0d.
//
// Solidity: function relayDeposits(uint256 upToDepositId, uint256 gasLimit) payable returns()
func (_Liquidity *LiquiditySession) RelayDeposits(upToDepositId *big.Int, gasLimit *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.RelayDeposits(&_Liquidity.TransactOpts, upToDepositId, gasLimit)
}

// RelayDeposits is a paid mutator transaction binding the contract method 0x76918c0d.
//
// Solidity: function relayDeposits(uint256 upToDepositId, uint256 gasLimit) payable returns()
func (_Liquidity *LiquidityTransactorSession) RelayDeposits(upToDepositId *big.Int, gasLimit *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.RelayDeposits(&_Liquidity.TransactOpts, upToDepositId, gasLimit)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Liquidity *LiquidityTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "renounceRole", role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Liquidity *LiquiditySession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.RenounceRole(&_Liquidity.TransactOpts, role, callerConfirmation)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address callerConfirmation) returns()
func (_Liquidity *LiquidityTransactorSession) RenounceRole(role [32]byte, callerConfirmation common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.RenounceRole(&_Liquidity.TransactOpts, role, callerConfirmation)
}

// ReplayDeposits is a paid mutator transaction binding the contract method 0xf33e7c11.
//
// Solidity: function replayDeposits(bytes message, uint32 newGasLimit, uint256 messageNonce) payable returns()
func (_Liquidity *LiquidityTransactor) ReplayDeposits(opts *bind.TransactOpts, message []byte, newGasLimit uint32, messageNonce *big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "replayDeposits", message, newGasLimit, messageNonce)
}

// ReplayDeposits is a paid mutator transaction binding the contract method 0xf33e7c11.
//
// Solidity: function replayDeposits(bytes message, uint32 newGasLimit, uint256 messageNonce) payable returns()
func (_Liquidity *LiquiditySession) ReplayDeposits(message []byte, newGasLimit uint32, messageNonce *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.ReplayDeposits(&_Liquidity.TransactOpts, message, newGasLimit, messageNonce)
}

// ReplayDeposits is a paid mutator transaction binding the contract method 0xf33e7c11.
//
// Solidity: function replayDeposits(bytes message, uint32 newGasLimit, uint256 messageNonce) payable returns()
func (_Liquidity *LiquidityTransactorSession) ReplayDeposits(message []byte, newGasLimit uint32, messageNonce *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.ReplayDeposits(&_Liquidity.TransactOpts, message, newGasLimit, messageNonce)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Liquidity *LiquidityTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Liquidity *LiquiditySession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.RevokeRole(&_Liquidity.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Liquidity *LiquidityTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.RevokeRole(&_Liquidity.TransactOpts, role, account)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Liquidity *LiquidityTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Liquidity *LiquiditySession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Liquidity.Contract.UpgradeToAndCall(&_Liquidity.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Liquidity *LiquidityTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Liquidity.Contract.UpgradeToAndCall(&_Liquidity.TransactOpts, newImplementation, data)
}

// LiquidityClaimableWithdrawalsProcessedIterator is returned from FilterClaimableWithdrawalsProcessed and is used to iterate over the raw logs and unpacked data for ClaimableWithdrawalsProcessed events raised by the Liquidity contract.
type LiquidityClaimableWithdrawalsProcessedIterator struct {
	Event *LiquidityClaimableWithdrawalsProcessed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityClaimableWithdrawalsProcessedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityClaimableWithdrawalsProcessed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityClaimableWithdrawalsProcessed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityClaimableWithdrawalsProcessedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityClaimableWithdrawalsProcessedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityClaimableWithdrawalsProcessed represents a ClaimableWithdrawalsProcessed event raised by the Liquidity contract.
type LiquidityClaimableWithdrawalsProcessed struct {
	LastProcessedClaimableWithdrawalId *big.Int
	Raw                                types.Log // Blockchain specific contextual infos
}

// FilterClaimableWithdrawalsProcessed is a free log retrieval operation binding the contract event 0xc822ed466f0466d78c7d829b38f3d4f38f64b1f07363274ac66dde2ed8d833e7.
//
// Solidity: event ClaimableWithdrawalsProcessed(uint256 indexed lastProcessedClaimableWithdrawalId)
func (_Liquidity *LiquidityFilterer) FilterClaimableWithdrawalsProcessed(opts *bind.FilterOpts, lastProcessedClaimableWithdrawalId []*big.Int) (*LiquidityClaimableWithdrawalsProcessedIterator, error) {

	var lastProcessedClaimableWithdrawalIdRule []interface{}
	for _, lastProcessedClaimableWithdrawalIdItem := range lastProcessedClaimableWithdrawalId {
		lastProcessedClaimableWithdrawalIdRule = append(lastProcessedClaimableWithdrawalIdRule, lastProcessedClaimableWithdrawalIdItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "ClaimableWithdrawalsProcessed", lastProcessedClaimableWithdrawalIdRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityClaimableWithdrawalsProcessedIterator{contract: _Liquidity.contract, event: "ClaimableWithdrawalsProcessed", logs: logs, sub: sub}, nil
}

// WatchClaimableWithdrawalsProcessed is a free log subscription operation binding the contract event 0xc822ed466f0466d78c7d829b38f3d4f38f64b1f07363274ac66dde2ed8d833e7.
//
// Solidity: event ClaimableWithdrawalsProcessed(uint256 indexed lastProcessedClaimableWithdrawalId)
func (_Liquidity *LiquidityFilterer) WatchClaimableWithdrawalsProcessed(opts *bind.WatchOpts, sink chan<- *LiquidityClaimableWithdrawalsProcessed, lastProcessedClaimableWithdrawalId []*big.Int) (event.Subscription, error) {

	var lastProcessedClaimableWithdrawalIdRule []interface{}
	for _, lastProcessedClaimableWithdrawalIdItem := range lastProcessedClaimableWithdrawalId {
		lastProcessedClaimableWithdrawalIdRule = append(lastProcessedClaimableWithdrawalIdRule, lastProcessedClaimableWithdrawalIdItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "ClaimableWithdrawalsProcessed", lastProcessedClaimableWithdrawalIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityClaimableWithdrawalsProcessed)
				if err := _Liquidity.contract.UnpackLog(event, "ClaimableWithdrawalsProcessed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseClaimableWithdrawalsProcessed is a log parse operation binding the contract event 0xc822ed466f0466d78c7d829b38f3d4f38f64b1f07363274ac66dde2ed8d833e7.
//
// Solidity: event ClaimableWithdrawalsProcessed(uint256 indexed lastProcessedClaimableWithdrawalId)
func (_Liquidity *LiquidityFilterer) ParseClaimableWithdrawalsProcessed(log types.Log) (*LiquidityClaimableWithdrawalsProcessed, error) {
	event := new(LiquidityClaimableWithdrawalsProcessed)
	if err := _Liquidity.contract.UnpackLog(event, "ClaimableWithdrawalsProcessed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityDepositCanceledIterator is returned from FilterDepositCanceled and is used to iterate over the raw logs and unpacked data for DepositCanceled events raised by the Liquidity contract.
type LiquidityDepositCanceledIterator struct {
	Event *LiquidityDepositCanceled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityDepositCanceledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityDepositCanceled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityDepositCanceled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityDepositCanceledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityDepositCanceledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityDepositCanceled represents a DepositCanceled event raised by the Liquidity contract.
type LiquidityDepositCanceled struct {
	DepositId *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDepositCanceled is a free log retrieval operation binding the contract event 0x0c1c0787650562778c282502dfd6dcea2fb79d23783920aa18aee36b4aa3dc48.
//
// Solidity: event DepositCanceled(uint256 indexed depositId)
func (_Liquidity *LiquidityFilterer) FilterDepositCanceled(opts *bind.FilterOpts, depositId []*big.Int) (*LiquidityDepositCanceledIterator, error) {

	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "DepositCanceled", depositIdRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityDepositCanceledIterator{contract: _Liquidity.contract, event: "DepositCanceled", logs: logs, sub: sub}, nil
}

// WatchDepositCanceled is a free log subscription operation binding the contract event 0x0c1c0787650562778c282502dfd6dcea2fb79d23783920aa18aee36b4aa3dc48.
//
// Solidity: event DepositCanceled(uint256 indexed depositId)
func (_Liquidity *LiquidityFilterer) WatchDepositCanceled(opts *bind.WatchOpts, sink chan<- *LiquidityDepositCanceled, depositId []*big.Int) (event.Subscription, error) {

	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "DepositCanceled", depositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityDepositCanceled)
				if err := _Liquidity.contract.UnpackLog(event, "DepositCanceled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDepositCanceled is a log parse operation binding the contract event 0x0c1c0787650562778c282502dfd6dcea2fb79d23783920aa18aee36b4aa3dc48.
//
// Solidity: event DepositCanceled(uint256 indexed depositId)
func (_Liquidity *LiquidityFilterer) ParseDepositCanceled(log types.Log) (*LiquidityDepositCanceled, error) {
	event := new(LiquidityDepositCanceled)
	if err := _Liquidity.contract.UnpackLog(event, "DepositCanceled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityDepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the Liquidity contract.
type LiquidityDepositedIterator struct {
	Event *LiquidityDeposited // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityDeposited)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityDeposited)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityDeposited represents a Deposited event raised by the Liquidity contract.
type LiquidityDeposited struct {
	DepositId         *big.Int
	Sender            common.Address
	RecipientSaltHash [32]byte
	TokenIndex        uint32
	Amount            *big.Int
	RequestedAt       *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0x1e88950eef3c1bd8dd83d765aec1f21f34ca153104f0acd7a6218bf8f48e8410.
//
// Solidity: event Deposited(uint256 indexed depositId, address indexed sender, bytes32 recipientSaltHash, uint32 tokenIndex, uint256 amount, uint256 requestedAt)
func (_Liquidity *LiquidityFilterer) FilterDeposited(opts *bind.FilterOpts, depositId []*big.Int, sender []common.Address) (*LiquidityDepositedIterator, error) {

	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "Deposited", depositIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityDepositedIterator{contract: _Liquidity.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0x1e88950eef3c1bd8dd83d765aec1f21f34ca153104f0acd7a6218bf8f48e8410.
//
// Solidity: event Deposited(uint256 indexed depositId, address indexed sender, bytes32 recipientSaltHash, uint32 tokenIndex, uint256 amount, uint256 requestedAt)
func (_Liquidity *LiquidityFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *LiquidityDeposited, depositId []*big.Int, sender []common.Address) (event.Subscription, error) {

	var depositIdRule []interface{}
	for _, depositIdItem := range depositId {
		depositIdRule = append(depositIdRule, depositIdItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "Deposited", depositIdRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityDeposited)
				if err := _Liquidity.contract.UnpackLog(event, "Deposited", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeposited is a log parse operation binding the contract event 0x1e88950eef3c1bd8dd83d765aec1f21f34ca153104f0acd7a6218bf8f48e8410.
//
// Solidity: event Deposited(uint256 indexed depositId, address indexed sender, bytes32 recipientSaltHash, uint32 tokenIndex, uint256 amount, uint256 requestedAt)
func (_Liquidity *LiquidityFilterer) ParseDeposited(log types.Log) (*LiquidityDeposited, error) {
	event := new(LiquidityDeposited)
	if err := _Liquidity.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityDepositsAnalyzedIterator is returned from FilterDepositsAnalyzed and is used to iterate over the raw logs and unpacked data for DepositsAnalyzed events raised by the Liquidity contract.
type LiquidityDepositsAnalyzedIterator struct {
	Event *LiquidityDepositsAnalyzed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityDepositsAnalyzedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityDepositsAnalyzed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityDepositsAnalyzed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityDepositsAnalyzedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityDepositsAnalyzedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityDepositsAnalyzed represents a DepositsAnalyzed event raised by the Liquidity contract.
type LiquidityDepositsAnalyzed struct {
	LastAnalyzedDepositId *big.Int
	RejectedIndices       []*big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterDepositsAnalyzed is a free log retrieval operation binding the contract event 0x48f1daf7209f4eda2120ee12101bdb8e7f108f2e8d84fc2d6a62b96682f9592c.
//
// Solidity: event DepositsAnalyzed(uint256 indexed lastAnalyzedDepositId, uint256[] rejectedIndices)
func (_Liquidity *LiquidityFilterer) FilterDepositsAnalyzed(opts *bind.FilterOpts, lastAnalyzedDepositId []*big.Int) (*LiquidityDepositsAnalyzedIterator, error) {

	var lastAnalyzedDepositIdRule []interface{}
	for _, lastAnalyzedDepositIdItem := range lastAnalyzedDepositId {
		lastAnalyzedDepositIdRule = append(lastAnalyzedDepositIdRule, lastAnalyzedDepositIdItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "DepositsAnalyzed", lastAnalyzedDepositIdRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityDepositsAnalyzedIterator{contract: _Liquidity.contract, event: "DepositsAnalyzed", logs: logs, sub: sub}, nil
}

// WatchDepositsAnalyzed is a free log subscription operation binding the contract event 0x48f1daf7209f4eda2120ee12101bdb8e7f108f2e8d84fc2d6a62b96682f9592c.
//
// Solidity: event DepositsAnalyzed(uint256 indexed lastAnalyzedDepositId, uint256[] rejectedIndices)
func (_Liquidity *LiquidityFilterer) WatchDepositsAnalyzed(opts *bind.WatchOpts, sink chan<- *LiquidityDepositsAnalyzed, lastAnalyzedDepositId []*big.Int) (event.Subscription, error) {

	var lastAnalyzedDepositIdRule []interface{}
	for _, lastAnalyzedDepositIdItem := range lastAnalyzedDepositId {
		lastAnalyzedDepositIdRule = append(lastAnalyzedDepositIdRule, lastAnalyzedDepositIdItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "DepositsAnalyzed", lastAnalyzedDepositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityDepositsAnalyzed)
				if err := _Liquidity.contract.UnpackLog(event, "DepositsAnalyzed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDepositsAnalyzed is a log parse operation binding the contract event 0x48f1daf7209f4eda2120ee12101bdb8e7f108f2e8d84fc2d6a62b96682f9592c.
//
// Solidity: event DepositsAnalyzed(uint256 indexed lastAnalyzedDepositId, uint256[] rejectedIndices)
func (_Liquidity *LiquidityFilterer) ParseDepositsAnalyzed(log types.Log) (*LiquidityDepositsAnalyzed, error) {
	event := new(LiquidityDepositsAnalyzed)
	if err := _Liquidity.contract.UnpackLog(event, "DepositsAnalyzed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityDepositsRelayedIterator is returned from FilterDepositsRelayed and is used to iterate over the raw logs and unpacked data for DepositsRelayed events raised by the Liquidity contract.
type LiquidityDepositsRelayedIterator struct {
	Event *LiquidityDepositsRelayed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityDepositsRelayedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityDepositsRelayed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityDepositsRelayed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityDepositsRelayedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityDepositsRelayedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityDepositsRelayed represents a DepositsRelayed event raised by the Liquidity contract.
type LiquidityDepositsRelayed struct {
	LastRelayedDepositId *big.Int
	GasLimit             *big.Int
	Message              []byte
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterDepositsRelayed is a free log retrieval operation binding the contract event 0x71df2a50f3ab1563277a27043770eb570eeaa3fc42ef76c607cdc667a82d03a4.
//
// Solidity: event DepositsRelayed(uint256 indexed lastRelayedDepositId, uint256 gasLimit, bytes message)
func (_Liquidity *LiquidityFilterer) FilterDepositsRelayed(opts *bind.FilterOpts, lastRelayedDepositId []*big.Int) (*LiquidityDepositsRelayedIterator, error) {

	var lastRelayedDepositIdRule []interface{}
	for _, lastRelayedDepositIdItem := range lastRelayedDepositId {
		lastRelayedDepositIdRule = append(lastRelayedDepositIdRule, lastRelayedDepositIdItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "DepositsRelayed", lastRelayedDepositIdRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityDepositsRelayedIterator{contract: _Liquidity.contract, event: "DepositsRelayed", logs: logs, sub: sub}, nil
}

// WatchDepositsRelayed is a free log subscription operation binding the contract event 0x71df2a50f3ab1563277a27043770eb570eeaa3fc42ef76c607cdc667a82d03a4.
//
// Solidity: event DepositsRelayed(uint256 indexed lastRelayedDepositId, uint256 gasLimit, bytes message)
func (_Liquidity *LiquidityFilterer) WatchDepositsRelayed(opts *bind.WatchOpts, sink chan<- *LiquidityDepositsRelayed, lastRelayedDepositId []*big.Int) (event.Subscription, error) {

	var lastRelayedDepositIdRule []interface{}
	for _, lastRelayedDepositIdItem := range lastRelayedDepositId {
		lastRelayedDepositIdRule = append(lastRelayedDepositIdRule, lastRelayedDepositIdItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "DepositsRelayed", lastRelayedDepositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityDepositsRelayed)
				if err := _Liquidity.contract.UnpackLog(event, "DepositsRelayed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDepositsRelayed is a log parse operation binding the contract event 0x71df2a50f3ab1563277a27043770eb570eeaa3fc42ef76c607cdc667a82d03a4.
//
// Solidity: event DepositsRelayed(uint256 indexed lastRelayedDepositId, uint256 gasLimit, bytes message)
func (_Liquidity *LiquidityFilterer) ParseDepositsRelayed(log types.Log) (*LiquidityDepositsRelayed, error) {
	event := new(LiquidityDepositsRelayed)
	if err := _Liquidity.contract.UnpackLog(event, "DepositsRelayed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityDepositsReplayedIterator is returned from FilterDepositsReplayed and is used to iterate over the raw logs and unpacked data for DepositsReplayed events raised by the Liquidity contract.
type LiquidityDepositsReplayedIterator struct {
	Event *LiquidityDepositsReplayed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityDepositsReplayedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityDepositsReplayed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityDepositsReplayed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityDepositsReplayedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityDepositsReplayedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityDepositsReplayed represents a DepositsReplayed event raised by the Liquidity contract.
type LiquidityDepositsReplayed struct {
	NewGasLimit  uint32
	MessageNonce *big.Int
	Message      []byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDepositsReplayed is a free log retrieval operation binding the contract event 0xf5c12b2d2cfce1b0a26156801ec4afed5e94b6fd3cbe492e197b66e59fbe944f.
//
// Solidity: event DepositsReplayed(uint32 newGasLimit, uint256 messageNonce, bytes message)
func (_Liquidity *LiquidityFilterer) FilterDepositsReplayed(opts *bind.FilterOpts) (*LiquidityDepositsReplayedIterator, error) {

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "DepositsReplayed")
	if err != nil {
		return nil, err
	}
	return &LiquidityDepositsReplayedIterator{contract: _Liquidity.contract, event: "DepositsReplayed", logs: logs, sub: sub}, nil
}

// WatchDepositsReplayed is a free log subscription operation binding the contract event 0xf5c12b2d2cfce1b0a26156801ec4afed5e94b6fd3cbe492e197b66e59fbe944f.
//
// Solidity: event DepositsReplayed(uint32 newGasLimit, uint256 messageNonce, bytes message)
func (_Liquidity *LiquidityFilterer) WatchDepositsReplayed(opts *bind.WatchOpts, sink chan<- *LiquidityDepositsReplayed) (event.Subscription, error) {

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "DepositsReplayed")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityDepositsReplayed)
				if err := _Liquidity.contract.UnpackLog(event, "DepositsReplayed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDepositsReplayed is a log parse operation binding the contract event 0xf5c12b2d2cfce1b0a26156801ec4afed5e94b6fd3cbe492e197b66e59fbe944f.
//
// Solidity: event DepositsReplayed(uint32 newGasLimit, uint256 messageNonce, bytes message)
func (_Liquidity *LiquidityFilterer) ParseDepositsReplayed(log types.Log) (*LiquidityDepositsReplayed, error) {
	event := new(LiquidityDepositsReplayed)
	if err := _Liquidity.contract.UnpackLog(event, "DepositsReplayed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityDirectWithdrawalsProcessedIterator is returned from FilterDirectWithdrawalsProcessed and is used to iterate over the raw logs and unpacked data for DirectWithdrawalsProcessed events raised by the Liquidity contract.
type LiquidityDirectWithdrawalsProcessedIterator struct {
	Event *LiquidityDirectWithdrawalsProcessed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityDirectWithdrawalsProcessedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityDirectWithdrawalsProcessed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityDirectWithdrawalsProcessed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityDirectWithdrawalsProcessedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityDirectWithdrawalsProcessedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityDirectWithdrawalsProcessed represents a DirectWithdrawalsProcessed event raised by the Liquidity contract.
type LiquidityDirectWithdrawalsProcessed struct {
	LastProcessedDirectWithdrawalId *big.Int
	Raw                             types.Log // Blockchain specific contextual infos
}

// FilterDirectWithdrawalsProcessed is a free log retrieval operation binding the contract event 0x6d89692e6d34d23438c6aee95e29bb1a47404cb2868acb8a3041d16dceadff18.
//
// Solidity: event DirectWithdrawalsProcessed(uint256 indexed lastProcessedDirectWithdrawalId)
func (_Liquidity *LiquidityFilterer) FilterDirectWithdrawalsProcessed(opts *bind.FilterOpts, lastProcessedDirectWithdrawalId []*big.Int) (*LiquidityDirectWithdrawalsProcessedIterator, error) {

	var lastProcessedDirectWithdrawalIdRule []interface{}
	for _, lastProcessedDirectWithdrawalIdItem := range lastProcessedDirectWithdrawalId {
		lastProcessedDirectWithdrawalIdRule = append(lastProcessedDirectWithdrawalIdRule, lastProcessedDirectWithdrawalIdItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "DirectWithdrawalsProcessed", lastProcessedDirectWithdrawalIdRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityDirectWithdrawalsProcessedIterator{contract: _Liquidity.contract, event: "DirectWithdrawalsProcessed", logs: logs, sub: sub}, nil
}

// WatchDirectWithdrawalsProcessed is a free log subscription operation binding the contract event 0x6d89692e6d34d23438c6aee95e29bb1a47404cb2868acb8a3041d16dceadff18.
//
// Solidity: event DirectWithdrawalsProcessed(uint256 indexed lastProcessedDirectWithdrawalId)
func (_Liquidity *LiquidityFilterer) WatchDirectWithdrawalsProcessed(opts *bind.WatchOpts, sink chan<- *LiquidityDirectWithdrawalsProcessed, lastProcessedDirectWithdrawalId []*big.Int) (event.Subscription, error) {

	var lastProcessedDirectWithdrawalIdRule []interface{}
	for _, lastProcessedDirectWithdrawalIdItem := range lastProcessedDirectWithdrawalId {
		lastProcessedDirectWithdrawalIdRule = append(lastProcessedDirectWithdrawalIdRule, lastProcessedDirectWithdrawalIdItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "DirectWithdrawalsProcessed", lastProcessedDirectWithdrawalIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityDirectWithdrawalsProcessed)
				if err := _Liquidity.contract.UnpackLog(event, "DirectWithdrawalsProcessed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDirectWithdrawalsProcessed is a log parse operation binding the contract event 0x6d89692e6d34d23438c6aee95e29bb1a47404cb2868acb8a3041d16dceadff18.
//
// Solidity: event DirectWithdrawalsProcessed(uint256 indexed lastProcessedDirectWithdrawalId)
func (_Liquidity *LiquidityFilterer) ParseDirectWithdrawalsProcessed(log types.Log) (*LiquidityDirectWithdrawalsProcessed, error) {
	event := new(LiquidityDirectWithdrawalsProcessed)
	if err := _Liquidity.contract.UnpackLog(event, "DirectWithdrawalsProcessed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Liquidity contract.
type LiquidityInitializedIterator struct {
	Event *LiquidityInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityInitialized represents a Initialized event raised by the Liquidity contract.
type LiquidityInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Liquidity *LiquidityFilterer) FilterInitialized(opts *bind.FilterOpts) (*LiquidityInitializedIterator, error) {

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &LiquidityInitializedIterator{contract: _Liquidity.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Liquidity *LiquidityFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *LiquidityInitialized) (event.Subscription, error) {

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityInitialized)
				if err := _Liquidity.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Liquidity *LiquidityFilterer) ParseInitialized(log types.Log) (*LiquidityInitialized, error) {
	event := new(LiquidityInitialized)
	if err := _Liquidity.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Liquidity contract.
type LiquidityRoleAdminChangedIterator struct {
	Event *LiquidityRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityRoleAdminChanged represents a RoleAdminChanged event raised by the Liquidity contract.
type LiquidityRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Liquidity *LiquidityFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*LiquidityRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityRoleAdminChangedIterator{contract: _Liquidity.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Liquidity *LiquidityFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *LiquidityRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityRoleAdminChanged)
				if err := _Liquidity.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Liquidity *LiquidityFilterer) ParseRoleAdminChanged(log types.Log) (*LiquidityRoleAdminChanged, error) {
	event := new(LiquidityRoleAdminChanged)
	if err := _Liquidity.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Liquidity contract.
type LiquidityRoleGrantedIterator struct {
	Event *LiquidityRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityRoleGranted represents a RoleGranted event raised by the Liquidity contract.
type LiquidityRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Liquidity *LiquidityFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*LiquidityRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityRoleGrantedIterator{contract: _Liquidity.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Liquidity *LiquidityFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *LiquidityRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityRoleGranted)
				if err := _Liquidity.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Liquidity *LiquidityFilterer) ParseRoleGranted(log types.Log) (*LiquidityRoleGranted, error) {
	event := new(LiquidityRoleGranted)
	if err := _Liquidity.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Liquidity contract.
type LiquidityRoleRevokedIterator struct {
	Event *LiquidityRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityRoleRevoked represents a RoleRevoked event raised by the Liquidity contract.
type LiquidityRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Liquidity *LiquidityFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*LiquidityRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityRoleRevokedIterator{contract: _Liquidity.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Liquidity *LiquidityFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *LiquidityRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityRoleRevoked)
				if err := _Liquidity.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Liquidity *LiquidityFilterer) ParseRoleRevoked(log types.Log) (*LiquidityRoleRevoked, error) {
	event := new(LiquidityRoleRevoked)
	if err := _Liquidity.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Liquidity contract.
type LiquidityUpgradedIterator struct {
	Event *LiquidityUpgraded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityUpgraded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityUpgraded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityUpgraded represents a Upgraded event raised by the Liquidity contract.
type LiquidityUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Liquidity *LiquidityFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*LiquidityUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityUpgradedIterator{contract: _Liquidity.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Liquidity *LiquidityFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *LiquidityUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityUpgraded)
				if err := _Liquidity.contract.UnpackLog(event, "Upgraded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Liquidity *LiquidityFilterer) ParseUpgraded(log types.Log) (*LiquidityUpgraded, error) {
	event := new(LiquidityUpgraded)
	if err := _Liquidity.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LiquidityWithdrawalClaimableIterator is returned from FilterWithdrawalClaimable and is used to iterate over the raw logs and unpacked data for WithdrawalClaimable events raised by the Liquidity contract.
type LiquidityWithdrawalClaimableIterator struct {
	Event *LiquidityWithdrawalClaimable // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *LiquidityWithdrawalClaimableIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityWithdrawalClaimable)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(LiquidityWithdrawalClaimable)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *LiquidityWithdrawalClaimableIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityWithdrawalClaimableIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityWithdrawalClaimable represents a WithdrawalClaimable event raised by the Liquidity contract.
type LiquidityWithdrawalClaimable struct {
	WithdrawalHash [32]byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalClaimable is a free log retrieval operation binding the contract event 0xd63a18c6858559d79659151118b33c6ad7cf4f7735b64fe9dbd85c509e11a117.
//
// Solidity: event WithdrawalClaimable(bytes32 indexed withdrawalHash)
func (_Liquidity *LiquidityFilterer) FilterWithdrawalClaimable(opts *bind.FilterOpts, withdrawalHash [][32]byte) (*LiquidityWithdrawalClaimableIterator, error) {

	var withdrawalHashRule []interface{}
	for _, withdrawalHashItem := range withdrawalHash {
		withdrawalHashRule = append(withdrawalHashRule, withdrawalHashItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "WithdrawalClaimable", withdrawalHashRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityWithdrawalClaimableIterator{contract: _Liquidity.contract, event: "WithdrawalClaimable", logs: logs, sub: sub}, nil
}

// WatchWithdrawalClaimable is a free log subscription operation binding the contract event 0xd63a18c6858559d79659151118b33c6ad7cf4f7735b64fe9dbd85c509e11a117.
//
// Solidity: event WithdrawalClaimable(bytes32 indexed withdrawalHash)
func (_Liquidity *LiquidityFilterer) WatchWithdrawalClaimable(opts *bind.WatchOpts, sink chan<- *LiquidityWithdrawalClaimable, withdrawalHash [][32]byte) (event.Subscription, error) {

	var withdrawalHashRule []interface{}
	for _, withdrawalHashItem := range withdrawalHash {
		withdrawalHashRule = append(withdrawalHashRule, withdrawalHashItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "WithdrawalClaimable", withdrawalHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityWithdrawalClaimable)
				if err := _Liquidity.contract.UnpackLog(event, "WithdrawalClaimable", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawalClaimable is a log parse operation binding the contract event 0xd63a18c6858559d79659151118b33c6ad7cf4f7735b64fe9dbd85c509e11a117.
//
// Solidity: event WithdrawalClaimable(bytes32 indexed withdrawalHash)
func (_Liquidity *LiquidityFilterer) ParseWithdrawalClaimable(log types.Log) (*LiquidityWithdrawalClaimable, error) {
	event := new(LiquidityWithdrawalClaimable)
	if err := _Liquidity.contract.UnpackLog(event, "WithdrawalClaimable", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
