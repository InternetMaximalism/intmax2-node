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

// RollupMetaData contains all meta data concerning the Rollup contract.
var RollupMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BlockHashAlreadyPosted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BlockNumberOutOfRange\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidBlockBuilder\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"MerkleTreeFull\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlyLiquidity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"OnlyScrollMessenger\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PairingCheckFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PairingOpCodeFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderAccountIdsEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderAccountIdsInvalidLength\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SenderPublicKeysEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooManyAccountIds\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"TooManySenderPublicKeys\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"accountIds\",\"type\":\"bytes\"}],\"name\":\"AccountIdsPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"prevBlockHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"depositTreeRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"signatureHash\",\"type\":\"bytes32\"}],\"name\":\"BlockPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"lastProcessedDepositId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"depositTreeRoot\",\"type\":\"bytes32\"}],\"name\":\"DepositsProcessed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"senderPublicKeys\",\"type\":\"uint256[]\"}],\"name\":\"PubKeysPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"blockBuilders\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"blockHashes\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"depositTreeRoot\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"}],\"name\":\"getBlockBuilder\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"}],\"name\":\"getBlockHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatestBlockNumber\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_scrollMessenger\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_liquidity\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_blockBuilderRegistry\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastProcessedDepositId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"txTreeRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes16\",\"name\":\"senderFlags\",\"type\":\"bytes16\"},{\"internalType\":\"bytes32[2]\",\"name\":\"aggregatedPublicKey\",\"type\":\"bytes32[2]\"},{\"internalType\":\"bytes32[4]\",\"name\":\"aggregatedSignature\",\"type\":\"bytes32[4]\"},{\"internalType\":\"bytes32[4]\",\"name\":\"messagePoint\",\"type\":\"bytes32[4]\"},{\"internalType\":\"bytes32\",\"name\":\"publicKeysHash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"senderAccountIds\",\"type\":\"bytes\"}],\"name\":\"postNonRegistrationBlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"txTreeRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes16\",\"name\":\"senderFlags\",\"type\":\"bytes16\"},{\"internalType\":\"bytes32[2]\",\"name\":\"aggregatedPublicKey\",\"type\":\"bytes32[2]\"},{\"internalType\":\"bytes32[4]\",\"name\":\"aggregatedSignature\",\"type\":\"bytes32[4]\"},{\"internalType\":\"bytes32[4]\",\"name\":\"messagePoint\",\"type\":\"bytes32[4]\"},{\"internalType\":\"uint256[]\",\"name\":\"senderPublicKeys\",\"type\":\"uint256[]\"}],\"name\":\"postRegistrationBlock\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_lastProcessedDepositId\",\"type\":\"uint256\"},{\"internalType\":\"bytes32[]\",\"name\":\"depositHashes\",\"type\":\"bytes32[]\"}],\"name\":\"processDeposits\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// RollupABI is the input ABI used to generate the binding from.
// Deprecated: Use RollupMetaData.ABI instead.
var RollupABI = RollupMetaData.ABI

// Rollup is an auto generated Go binding around an Ethereum contract.
type Rollup struct {
	RollupCaller     // Read-only binding to the contract
	RollupTransactor // Write-only binding to the contract
	RollupFilterer   // Log filterer for contract events
}

// RollupCaller is an auto generated read-only Go binding around an Ethereum contract.
type RollupCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RollupTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RollupTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RollupFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RollupFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RollupSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RollupSession struct {
	Contract     *Rollup           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RollupCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RollupCallerSession struct {
	Contract *RollupCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// RollupTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RollupTransactorSession struct {
	Contract     *RollupTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RollupRaw is an auto generated low-level Go binding around an Ethereum contract.
type RollupRaw struct {
	Contract *Rollup // Generic contract binding to access the raw methods on
}

// RollupCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RollupCallerRaw struct {
	Contract *RollupCaller // Generic read-only contract binding to access the raw methods on
}

// RollupTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RollupTransactorRaw struct {
	Contract *RollupTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRollup creates a new instance of Rollup, bound to a specific deployed contract.
func NewRollup(address common.Address, backend bind.ContractBackend) (*Rollup, error) {
	contract, err := bindRollup(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Rollup{RollupCaller: RollupCaller{contract: contract}, RollupTransactor: RollupTransactor{contract: contract}, RollupFilterer: RollupFilterer{contract: contract}}, nil
}

// NewRollupCaller creates a new read-only instance of Rollup, bound to a specific deployed contract.
func NewRollupCaller(address common.Address, caller bind.ContractCaller) (*RollupCaller, error) {
	contract, err := bindRollup(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RollupCaller{contract: contract}, nil
}

// NewRollupTransactor creates a new write-only instance of Rollup, bound to a specific deployed contract.
func NewRollupTransactor(address common.Address, transactor bind.ContractTransactor) (*RollupTransactor, error) {
	contract, err := bindRollup(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RollupTransactor{contract: contract}, nil
}

// NewRollupFilterer creates a new log filterer instance of Rollup, bound to a specific deployed contract.
func NewRollupFilterer(address common.Address, filterer bind.ContractFilterer) (*RollupFilterer, error) {
	contract, err := bindRollup(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RollupFilterer{contract: contract}, nil
}

// bindRollup binds a generic wrapper to an already deployed contract.
func bindRollup(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RollupMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Rollup *RollupRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Rollup.Contract.RollupCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Rollup *RollupRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Rollup.Contract.RollupTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Rollup *RollupRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Rollup.Contract.RollupTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Rollup *RollupCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Rollup.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Rollup *RollupTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Rollup.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Rollup *RollupTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Rollup.Contract.contract.Transact(opts, method, params...)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Rollup *RollupCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Rollup *RollupSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Rollup.Contract.UPGRADEINTERFACEVERSION(&_Rollup.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Rollup *RollupCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Rollup.Contract.UPGRADEINTERFACEVERSION(&_Rollup.CallOpts)
}

// BlockBuilders is a free data retrieval call binding the contract method 0xb568a055.
//
// Solidity: function blockBuilders(uint256 ) view returns(address)
func (_Rollup *RollupCaller) BlockBuilders(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "blockBuilders", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BlockBuilders is a free data retrieval call binding the contract method 0xb568a055.
//
// Solidity: function blockBuilders(uint256 ) view returns(address)
func (_Rollup *RollupSession) BlockBuilders(arg0 *big.Int) (common.Address, error) {
	return _Rollup.Contract.BlockBuilders(&_Rollup.CallOpts, arg0)
}

// BlockBuilders is a free data retrieval call binding the contract method 0xb568a055.
//
// Solidity: function blockBuilders(uint256 ) view returns(address)
func (_Rollup *RollupCallerSession) BlockBuilders(arg0 *big.Int) (common.Address, error) {
	return _Rollup.Contract.BlockBuilders(&_Rollup.CallOpts, arg0)
}

// BlockHashes is a free data retrieval call binding the contract method 0x34cdf78d.
//
// Solidity: function blockHashes(uint256 ) view returns(bytes32)
func (_Rollup *RollupCaller) BlockHashes(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "blockHashes", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BlockHashes is a free data retrieval call binding the contract method 0x34cdf78d.
//
// Solidity: function blockHashes(uint256 ) view returns(bytes32)
func (_Rollup *RollupSession) BlockHashes(arg0 *big.Int) ([32]byte, error) {
	return _Rollup.Contract.BlockHashes(&_Rollup.CallOpts, arg0)
}

// BlockHashes is a free data retrieval call binding the contract method 0x34cdf78d.
//
// Solidity: function blockHashes(uint256 ) view returns(bytes32)
func (_Rollup *RollupCallerSession) BlockHashes(arg0 *big.Int) ([32]byte, error) {
	return _Rollup.Contract.BlockHashes(&_Rollup.CallOpts, arg0)
}

// DepositTreeRoot is a free data retrieval call binding the contract method 0x1d9a07b9.
//
// Solidity: function depositTreeRoot() view returns(bytes32)
func (_Rollup *RollupCaller) DepositTreeRoot(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "depositTreeRoot")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DepositTreeRoot is a free data retrieval call binding the contract method 0x1d9a07b9.
//
// Solidity: function depositTreeRoot() view returns(bytes32)
func (_Rollup *RollupSession) DepositTreeRoot() ([32]byte, error) {
	return _Rollup.Contract.DepositTreeRoot(&_Rollup.CallOpts)
}

// DepositTreeRoot is a free data retrieval call binding the contract method 0x1d9a07b9.
//
// Solidity: function depositTreeRoot() view returns(bytes32)
func (_Rollup *RollupCallerSession) DepositTreeRoot() ([32]byte, error) {
	return _Rollup.Contract.DepositTreeRoot(&_Rollup.CallOpts)
}

// GetBlockBuilder is a free data retrieval call binding the contract method 0x06f743c0.
//
// Solidity: function getBlockBuilder(uint32 blockNumber) view returns(address)
func (_Rollup *RollupCaller) GetBlockBuilder(opts *bind.CallOpts, blockNumber uint32) (common.Address, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "getBlockBuilder", blockNumber)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetBlockBuilder is a free data retrieval call binding the contract method 0x06f743c0.
//
// Solidity: function getBlockBuilder(uint32 blockNumber) view returns(address)
func (_Rollup *RollupSession) GetBlockBuilder(blockNumber uint32) (common.Address, error) {
	return _Rollup.Contract.GetBlockBuilder(&_Rollup.CallOpts, blockNumber)
}

// GetBlockBuilder is a free data retrieval call binding the contract method 0x06f743c0.
//
// Solidity: function getBlockBuilder(uint32 blockNumber) view returns(address)
func (_Rollup *RollupCallerSession) GetBlockBuilder(blockNumber uint32) (common.Address, error) {
	return _Rollup.Contract.GetBlockBuilder(&_Rollup.CallOpts, blockNumber)
}

// GetBlockHash is a free data retrieval call binding the contract method 0xd2b210a1.
//
// Solidity: function getBlockHash(uint32 blockNumber) view returns(bytes32)
func (_Rollup *RollupCaller) GetBlockHash(opts *bind.CallOpts, blockNumber uint32) ([32]byte, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "getBlockHash", blockNumber)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetBlockHash is a free data retrieval call binding the contract method 0xd2b210a1.
//
// Solidity: function getBlockHash(uint32 blockNumber) view returns(bytes32)
func (_Rollup *RollupSession) GetBlockHash(blockNumber uint32) ([32]byte, error) {
	return _Rollup.Contract.GetBlockHash(&_Rollup.CallOpts, blockNumber)
}

// GetBlockHash is a free data retrieval call binding the contract method 0xd2b210a1.
//
// Solidity: function getBlockHash(uint32 blockNumber) view returns(bytes32)
func (_Rollup *RollupCallerSession) GetBlockHash(blockNumber uint32) ([32]byte, error) {
	return _Rollup.Contract.GetBlockHash(&_Rollup.CallOpts, blockNumber)
}

// GetLatestBlockNumber is a free data retrieval call binding the contract method 0xc76c3884.
//
// Solidity: function getLatestBlockNumber() view returns(uint32)
func (_Rollup *RollupCaller) GetLatestBlockNumber(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "getLatestBlockNumber")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetLatestBlockNumber is a free data retrieval call binding the contract method 0xc76c3884.
//
// Solidity: function getLatestBlockNumber() view returns(uint32)
func (_Rollup *RollupSession) GetLatestBlockNumber() (uint32, error) {
	return _Rollup.Contract.GetLatestBlockNumber(&_Rollup.CallOpts)
}

// GetLatestBlockNumber is a free data retrieval call binding the contract method 0xc76c3884.
//
// Solidity: function getLatestBlockNumber() view returns(uint32)
func (_Rollup *RollupCallerSession) GetLatestBlockNumber() (uint32, error) {
	return _Rollup.Contract.GetLatestBlockNumber(&_Rollup.CallOpts)
}

// LastProcessedDepositId is a free data retrieval call binding the contract method 0x567419c1.
//
// Solidity: function lastProcessedDepositId() view returns(uint256)
func (_Rollup *RollupCaller) LastProcessedDepositId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "lastProcessedDepositId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastProcessedDepositId is a free data retrieval call binding the contract method 0x567419c1.
//
// Solidity: function lastProcessedDepositId() view returns(uint256)
func (_Rollup *RollupSession) LastProcessedDepositId() (*big.Int, error) {
	return _Rollup.Contract.LastProcessedDepositId(&_Rollup.CallOpts)
}

// LastProcessedDepositId is a free data retrieval call binding the contract method 0x567419c1.
//
// Solidity: function lastProcessedDepositId() view returns(uint256)
func (_Rollup *RollupCallerSession) LastProcessedDepositId() (*big.Int, error) {
	return _Rollup.Contract.LastProcessedDepositId(&_Rollup.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Rollup *RollupCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Rollup *RollupSession) Owner() (common.Address, error) {
	return _Rollup.Contract.Owner(&_Rollup.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Rollup *RollupCallerSession) Owner() (common.Address, error) {
	return _Rollup.Contract.Owner(&_Rollup.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Rollup *RollupCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Rollup.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Rollup *RollupSession) ProxiableUUID() ([32]byte, error) {
	return _Rollup.Contract.ProxiableUUID(&_Rollup.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Rollup *RollupCallerSession) ProxiableUUID() ([32]byte, error) {
	return _Rollup.Contract.ProxiableUUID(&_Rollup.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _scrollMessenger, address _liquidity, address _blockBuilderRegistry) returns()
func (_Rollup *RollupTransactor) Initialize(opts *bind.TransactOpts, _scrollMessenger common.Address, _liquidity common.Address, _blockBuilderRegistry common.Address) (*types.Transaction, error) {
	return _Rollup.contract.Transact(opts, "initialize", _scrollMessenger, _liquidity, _blockBuilderRegistry)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _scrollMessenger, address _liquidity, address _blockBuilderRegistry) returns()
func (_Rollup *RollupSession) Initialize(_scrollMessenger common.Address, _liquidity common.Address, _blockBuilderRegistry common.Address) (*types.Transaction, error) {
	return _Rollup.Contract.Initialize(&_Rollup.TransactOpts, _scrollMessenger, _liquidity, _blockBuilderRegistry)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _scrollMessenger, address _liquidity, address _blockBuilderRegistry) returns()
func (_Rollup *RollupTransactorSession) Initialize(_scrollMessenger common.Address, _liquidity common.Address, _blockBuilderRegistry common.Address) (*types.Transaction, error) {
	return _Rollup.Contract.Initialize(&_Rollup.TransactOpts, _scrollMessenger, _liquidity, _blockBuilderRegistry)
}

// PostNonRegistrationBlock is a paid mutator transaction binding the contract method 0x842f1bfc.
//
// Solidity: function postNonRegistrationBlock(bytes32 txTreeRoot, bytes16 senderFlags, bytes32[2] aggregatedPublicKey, bytes32[4] aggregatedSignature, bytes32[4] messagePoint, bytes32 publicKeysHash, bytes senderAccountIds) returns()
func (_Rollup *RollupTransactor) PostNonRegistrationBlock(opts *bind.TransactOpts, txTreeRoot [32]byte, senderFlags [16]byte, aggregatedPublicKey [2][32]byte, aggregatedSignature [4][32]byte, messagePoint [4][32]byte, publicKeysHash [32]byte, senderAccountIds []byte) (*types.Transaction, error) {
	return _Rollup.contract.Transact(opts, "postNonRegistrationBlock", txTreeRoot, senderFlags, aggregatedPublicKey, aggregatedSignature, messagePoint, publicKeysHash, senderAccountIds)
}

// PostNonRegistrationBlock is a paid mutator transaction binding the contract method 0x842f1bfc.
//
// Solidity: function postNonRegistrationBlock(bytes32 txTreeRoot, bytes16 senderFlags, bytes32[2] aggregatedPublicKey, bytes32[4] aggregatedSignature, bytes32[4] messagePoint, bytes32 publicKeysHash, bytes senderAccountIds) returns()
func (_Rollup *RollupSession) PostNonRegistrationBlock(txTreeRoot [32]byte, senderFlags [16]byte, aggregatedPublicKey [2][32]byte, aggregatedSignature [4][32]byte, messagePoint [4][32]byte, publicKeysHash [32]byte, senderAccountIds []byte) (*types.Transaction, error) {
	return _Rollup.Contract.PostNonRegistrationBlock(&_Rollup.TransactOpts, txTreeRoot, senderFlags, aggregatedPublicKey, aggregatedSignature, messagePoint, publicKeysHash, senderAccountIds)
}

// PostNonRegistrationBlock is a paid mutator transaction binding the contract method 0x842f1bfc.
//
// Solidity: function postNonRegistrationBlock(bytes32 txTreeRoot, bytes16 senderFlags, bytes32[2] aggregatedPublicKey, bytes32[4] aggregatedSignature, bytes32[4] messagePoint, bytes32 publicKeysHash, bytes senderAccountIds) returns()
func (_Rollup *RollupTransactorSession) PostNonRegistrationBlock(txTreeRoot [32]byte, senderFlags [16]byte, aggregatedPublicKey [2][32]byte, aggregatedSignature [4][32]byte, messagePoint [4][32]byte, publicKeysHash [32]byte, senderAccountIds []byte) (*types.Transaction, error) {
	return _Rollup.Contract.PostNonRegistrationBlock(&_Rollup.TransactOpts, txTreeRoot, senderFlags, aggregatedPublicKey, aggregatedSignature, messagePoint, publicKeysHash, senderAccountIds)
}

// PostRegistrationBlock is a paid mutator transaction binding the contract method 0xc7bffeeb.
//
// Solidity: function postRegistrationBlock(bytes32 txTreeRoot, bytes16 senderFlags, bytes32[2] aggregatedPublicKey, bytes32[4] aggregatedSignature, bytes32[4] messagePoint, uint256[] senderPublicKeys) returns()
func (_Rollup *RollupTransactor) PostRegistrationBlock(opts *bind.TransactOpts, txTreeRoot [32]byte, senderFlags [16]byte, aggregatedPublicKey [2][32]byte, aggregatedSignature [4][32]byte, messagePoint [4][32]byte, senderPublicKeys []*big.Int) (*types.Transaction, error) {
	return _Rollup.contract.Transact(opts, "postRegistrationBlock", txTreeRoot, senderFlags, aggregatedPublicKey, aggregatedSignature, messagePoint, senderPublicKeys)
}

// PostRegistrationBlock is a paid mutator transaction binding the contract method 0xc7bffeeb.
//
// Solidity: function postRegistrationBlock(bytes32 txTreeRoot, bytes16 senderFlags, bytes32[2] aggregatedPublicKey, bytes32[4] aggregatedSignature, bytes32[4] messagePoint, uint256[] senderPublicKeys) returns()
func (_Rollup *RollupSession) PostRegistrationBlock(txTreeRoot [32]byte, senderFlags [16]byte, aggregatedPublicKey [2][32]byte, aggregatedSignature [4][32]byte, messagePoint [4][32]byte, senderPublicKeys []*big.Int) (*types.Transaction, error) {
	return _Rollup.Contract.PostRegistrationBlock(&_Rollup.TransactOpts, txTreeRoot, senderFlags, aggregatedPublicKey, aggregatedSignature, messagePoint, senderPublicKeys)
}

// PostRegistrationBlock is a paid mutator transaction binding the contract method 0xc7bffeeb.
//
// Solidity: function postRegistrationBlock(bytes32 txTreeRoot, bytes16 senderFlags, bytes32[2] aggregatedPublicKey, bytes32[4] aggregatedSignature, bytes32[4] messagePoint, uint256[] senderPublicKeys) returns()
func (_Rollup *RollupTransactorSession) PostRegistrationBlock(txTreeRoot [32]byte, senderFlags [16]byte, aggregatedPublicKey [2][32]byte, aggregatedSignature [4][32]byte, messagePoint [4][32]byte, senderPublicKeys []*big.Int) (*types.Transaction, error) {
	return _Rollup.Contract.PostRegistrationBlock(&_Rollup.TransactOpts, txTreeRoot, senderFlags, aggregatedPublicKey, aggregatedSignature, messagePoint, senderPublicKeys)
}

// ProcessDeposits is a paid mutator transaction binding the contract method 0xf03efa37.
//
// Solidity: function processDeposits(uint256 _lastProcessedDepositId, bytes32[] depositHashes) returns()
func (_Rollup *RollupTransactor) ProcessDeposits(opts *bind.TransactOpts, _lastProcessedDepositId *big.Int, depositHashes [][32]byte) (*types.Transaction, error) {
	return _Rollup.contract.Transact(opts, "processDeposits", _lastProcessedDepositId, depositHashes)
}

// ProcessDeposits is a paid mutator transaction binding the contract method 0xf03efa37.
//
// Solidity: function processDeposits(uint256 _lastProcessedDepositId, bytes32[] depositHashes) returns()
func (_Rollup *RollupSession) ProcessDeposits(_lastProcessedDepositId *big.Int, depositHashes [][32]byte) (*types.Transaction, error) {
	return _Rollup.Contract.ProcessDeposits(&_Rollup.TransactOpts, _lastProcessedDepositId, depositHashes)
}

// ProcessDeposits is a paid mutator transaction binding the contract method 0xf03efa37.
//
// Solidity: function processDeposits(uint256 _lastProcessedDepositId, bytes32[] depositHashes) returns()
func (_Rollup *RollupTransactorSession) ProcessDeposits(_lastProcessedDepositId *big.Int, depositHashes [][32]byte) (*types.Transaction, error) {
	return _Rollup.Contract.ProcessDeposits(&_Rollup.TransactOpts, _lastProcessedDepositId, depositHashes)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Rollup *RollupTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Rollup.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Rollup *RollupSession) RenounceOwnership() (*types.Transaction, error) {
	return _Rollup.Contract.RenounceOwnership(&_Rollup.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Rollup *RollupTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Rollup.Contract.RenounceOwnership(&_Rollup.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Rollup *RollupTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Rollup.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Rollup *RollupSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Rollup.Contract.TransferOwnership(&_Rollup.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Rollup *RollupTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Rollup.Contract.TransferOwnership(&_Rollup.TransactOpts, newOwner)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Rollup *RollupTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Rollup.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Rollup *RollupSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Rollup.Contract.UpgradeToAndCall(&_Rollup.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Rollup *RollupTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Rollup.Contract.UpgradeToAndCall(&_Rollup.TransactOpts, newImplementation, data)
}

// RollupAccountIdsPostedIterator is returned from FilterAccountIdsPosted and is used to iterate over the raw logs and unpacked data for AccountIdsPosted events raised by the Rollup contract.
type RollupAccountIdsPostedIterator struct {
	Event *RollupAccountIdsPosted // Event containing the contract specifics and raw log

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
func (it *RollupAccountIdsPostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RollupAccountIdsPosted)
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
		it.Event = new(RollupAccountIdsPosted)
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
func (it *RollupAccountIdsPostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RollupAccountIdsPostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RollupAccountIdsPosted represents a AccountIdsPosted event raised by the Rollup contract.
type RollupAccountIdsPosted struct {
	BlockNumber *big.Int
	AccountIds  []byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterAccountIdsPosted is a free log retrieval operation binding the contract event 0x7e4f72212a0dbcb829e5fdaad6c3cb28556083f04fc687bfa2b2e40432bbb935.
//
// Solidity: event AccountIdsPosted(uint256 indexed blockNumber, bytes accountIds)
func (_Rollup *RollupFilterer) FilterAccountIdsPosted(opts *bind.FilterOpts, blockNumber []*big.Int) (*RollupAccountIdsPostedIterator, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}

	logs, sub, err := _Rollup.contract.FilterLogs(opts, "AccountIdsPosted", blockNumberRule)
	if err != nil {
		return nil, err
	}
	return &RollupAccountIdsPostedIterator{contract: _Rollup.contract, event: "AccountIdsPosted", logs: logs, sub: sub}, nil
}

// WatchAccountIdsPosted is a free log subscription operation binding the contract event 0x7e4f72212a0dbcb829e5fdaad6c3cb28556083f04fc687bfa2b2e40432bbb935.
//
// Solidity: event AccountIdsPosted(uint256 indexed blockNumber, bytes accountIds)
func (_Rollup *RollupFilterer) WatchAccountIdsPosted(opts *bind.WatchOpts, sink chan<- *RollupAccountIdsPosted, blockNumber []*big.Int) (event.Subscription, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}

	logs, sub, err := _Rollup.contract.WatchLogs(opts, "AccountIdsPosted", blockNumberRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RollupAccountIdsPosted)
				if err := _Rollup.contract.UnpackLog(event, "AccountIdsPosted", log); err != nil {
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

// ParseAccountIdsPosted is a log parse operation binding the contract event 0x7e4f72212a0dbcb829e5fdaad6c3cb28556083f04fc687bfa2b2e40432bbb935.
//
// Solidity: event AccountIdsPosted(uint256 indexed blockNumber, bytes accountIds)
func (_Rollup *RollupFilterer) ParseAccountIdsPosted(log types.Log) (*RollupAccountIdsPosted, error) {
	event := new(RollupAccountIdsPosted)
	if err := _Rollup.contract.UnpackLog(event, "AccountIdsPosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RollupBlockPostedIterator is returned from FilterBlockPosted and is used to iterate over the raw logs and unpacked data for BlockPosted events raised by the Rollup contract.
type RollupBlockPostedIterator struct {
	Event *RollupBlockPosted // Event containing the contract specifics and raw log

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
func (it *RollupBlockPostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RollupBlockPosted)
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
		it.Event = new(RollupBlockPosted)
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
func (it *RollupBlockPostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RollupBlockPostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RollupBlockPosted represents a BlockPosted event raised by the Rollup contract.
type RollupBlockPosted struct {
	PrevBlockHash   [32]byte
	BlockBuilder    common.Address
	BlockNumber     *big.Int
	DepositTreeRoot [32]byte
	SignatureHash   [32]byte
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterBlockPosted is a free log retrieval operation binding the contract event 0xe27163b76905dc373b4ad854ddc9403bbac659c5f1c5191c39e5a7c44574040a.
//
// Solidity: event BlockPosted(bytes32 indexed prevBlockHash, address indexed blockBuilder, uint256 blockNumber, bytes32 depositTreeRoot, bytes32 signatureHash)
func (_Rollup *RollupFilterer) FilterBlockPosted(opts *bind.FilterOpts, prevBlockHash [][32]byte, blockBuilder []common.Address) (*RollupBlockPostedIterator, error) {

	var prevBlockHashRule []interface{}
	for _, prevBlockHashItem := range prevBlockHash {
		prevBlockHashRule = append(prevBlockHashRule, prevBlockHashItem)
	}
	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _Rollup.contract.FilterLogs(opts, "BlockPosted", prevBlockHashRule, blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return &RollupBlockPostedIterator{contract: _Rollup.contract, event: "BlockPosted", logs: logs, sub: sub}, nil
}

// WatchBlockPosted is a free log subscription operation binding the contract event 0xe27163b76905dc373b4ad854ddc9403bbac659c5f1c5191c39e5a7c44574040a.
//
// Solidity: event BlockPosted(bytes32 indexed prevBlockHash, address indexed blockBuilder, uint256 blockNumber, bytes32 depositTreeRoot, bytes32 signatureHash)
func (_Rollup *RollupFilterer) WatchBlockPosted(opts *bind.WatchOpts, sink chan<- *RollupBlockPosted, prevBlockHash [][32]byte, blockBuilder []common.Address) (event.Subscription, error) {

	var prevBlockHashRule []interface{}
	for _, prevBlockHashItem := range prevBlockHash {
		prevBlockHashRule = append(prevBlockHashRule, prevBlockHashItem)
	}
	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _Rollup.contract.WatchLogs(opts, "BlockPosted", prevBlockHashRule, blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RollupBlockPosted)
				if err := _Rollup.contract.UnpackLog(event, "BlockPosted", log); err != nil {
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

// ParseBlockPosted is a log parse operation binding the contract event 0xe27163b76905dc373b4ad854ddc9403bbac659c5f1c5191c39e5a7c44574040a.
//
// Solidity: event BlockPosted(bytes32 indexed prevBlockHash, address indexed blockBuilder, uint256 blockNumber, bytes32 depositTreeRoot, bytes32 signatureHash)
func (_Rollup *RollupFilterer) ParseBlockPosted(log types.Log) (*RollupBlockPosted, error) {
	event := new(RollupBlockPosted)
	if err := _Rollup.contract.UnpackLog(event, "BlockPosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RollupDepositsProcessedIterator is returned from FilterDepositsProcessed and is used to iterate over the raw logs and unpacked data for DepositsProcessed events raised by the Rollup contract.
type RollupDepositsProcessedIterator struct {
	Event *RollupDepositsProcessed // Event containing the contract specifics and raw log

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
func (it *RollupDepositsProcessedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RollupDepositsProcessed)
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
		it.Event = new(RollupDepositsProcessed)
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
func (it *RollupDepositsProcessedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RollupDepositsProcessedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RollupDepositsProcessed represents a DepositsProcessed event raised by the Rollup contract.
type RollupDepositsProcessed struct {
	LastProcessedDepositId *big.Int
	DepositTreeRoot        [32]byte
	Raw                    types.Log // Blockchain specific contextual infos
}

// FilterDepositsProcessed is a free log retrieval operation binding the contract event 0x598fdeda3e3b7298359a08c9a438151e3481bdcf3c3ff9906fa71dcc723617b0.
//
// Solidity: event DepositsProcessed(uint256 indexed lastProcessedDepositId, bytes32 depositTreeRoot)
func (_Rollup *RollupFilterer) FilterDepositsProcessed(opts *bind.FilterOpts, lastProcessedDepositId []*big.Int) (*RollupDepositsProcessedIterator, error) {

	var lastProcessedDepositIdRule []interface{}
	for _, lastProcessedDepositIdItem := range lastProcessedDepositId {
		lastProcessedDepositIdRule = append(lastProcessedDepositIdRule, lastProcessedDepositIdItem)
	}

	logs, sub, err := _Rollup.contract.FilterLogs(opts, "DepositsProcessed", lastProcessedDepositIdRule)
	if err != nil {
		return nil, err
	}
	return &RollupDepositsProcessedIterator{contract: _Rollup.contract, event: "DepositsProcessed", logs: logs, sub: sub}, nil
}

// WatchDepositsProcessed is a free log subscription operation binding the contract event 0x598fdeda3e3b7298359a08c9a438151e3481bdcf3c3ff9906fa71dcc723617b0.
//
// Solidity: event DepositsProcessed(uint256 indexed lastProcessedDepositId, bytes32 depositTreeRoot)
func (_Rollup *RollupFilterer) WatchDepositsProcessed(opts *bind.WatchOpts, sink chan<- *RollupDepositsProcessed, lastProcessedDepositId []*big.Int) (event.Subscription, error) {

	var lastProcessedDepositIdRule []interface{}
	for _, lastProcessedDepositIdItem := range lastProcessedDepositId {
		lastProcessedDepositIdRule = append(lastProcessedDepositIdRule, lastProcessedDepositIdItem)
	}

	logs, sub, err := _Rollup.contract.WatchLogs(opts, "DepositsProcessed", lastProcessedDepositIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RollupDepositsProcessed)
				if err := _Rollup.contract.UnpackLog(event, "DepositsProcessed", log); err != nil {
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

// ParseDepositsProcessed is a log parse operation binding the contract event 0x598fdeda3e3b7298359a08c9a438151e3481bdcf3c3ff9906fa71dcc723617b0.
//
// Solidity: event DepositsProcessed(uint256 indexed lastProcessedDepositId, bytes32 depositTreeRoot)
func (_Rollup *RollupFilterer) ParseDepositsProcessed(log types.Log) (*RollupDepositsProcessed, error) {
	event := new(RollupDepositsProcessed)
	if err := _Rollup.contract.UnpackLog(event, "DepositsProcessed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RollupInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Rollup contract.
type RollupInitializedIterator struct {
	Event *RollupInitialized // Event containing the contract specifics and raw log

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
func (it *RollupInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RollupInitialized)
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
		it.Event = new(RollupInitialized)
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
func (it *RollupInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RollupInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RollupInitialized represents a Initialized event raised by the Rollup contract.
type RollupInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Rollup *RollupFilterer) FilterInitialized(opts *bind.FilterOpts) (*RollupInitializedIterator, error) {

	logs, sub, err := _Rollup.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &RollupInitializedIterator{contract: _Rollup.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Rollup *RollupFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *RollupInitialized) (event.Subscription, error) {

	logs, sub, err := _Rollup.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RollupInitialized)
				if err := _Rollup.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Rollup *RollupFilterer) ParseInitialized(log types.Log) (*RollupInitialized, error) {
	event := new(RollupInitialized)
	if err := _Rollup.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RollupOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Rollup contract.
type RollupOwnershipTransferredIterator struct {
	Event *RollupOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *RollupOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RollupOwnershipTransferred)
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
		it.Event = new(RollupOwnershipTransferred)
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
func (it *RollupOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RollupOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RollupOwnershipTransferred represents a OwnershipTransferred event raised by the Rollup contract.
type RollupOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Rollup *RollupFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*RollupOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Rollup.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &RollupOwnershipTransferredIterator{contract: _Rollup.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Rollup *RollupFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *RollupOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Rollup.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RollupOwnershipTransferred)
				if err := _Rollup.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Rollup *RollupFilterer) ParseOwnershipTransferred(log types.Log) (*RollupOwnershipTransferred, error) {
	event := new(RollupOwnershipTransferred)
	if err := _Rollup.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RollupPubKeysPostedIterator is returned from FilterPubKeysPosted and is used to iterate over the raw logs and unpacked data for PubKeysPosted events raised by the Rollup contract.
type RollupPubKeysPostedIterator struct {
	Event *RollupPubKeysPosted // Event containing the contract specifics and raw log

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
func (it *RollupPubKeysPostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RollupPubKeysPosted)
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
		it.Event = new(RollupPubKeysPosted)
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
func (it *RollupPubKeysPostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RollupPubKeysPostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RollupPubKeysPosted represents a PubKeysPosted event raised by the Rollup contract.
type RollupPubKeysPosted struct {
	BlockNumber      *big.Int
	SenderPublicKeys []*big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterPubKeysPosted is a free log retrieval operation binding the contract event 0x7a873ae11bb72d3848d5c34af1354cdd12aa5b30ec0d5187ec00ba5cefcd5c7f.
//
// Solidity: event PubKeysPosted(uint256 indexed blockNumber, uint256[] senderPublicKeys)
func (_Rollup *RollupFilterer) FilterPubKeysPosted(opts *bind.FilterOpts, blockNumber []*big.Int) (*RollupPubKeysPostedIterator, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}

	logs, sub, err := _Rollup.contract.FilterLogs(opts, "PubKeysPosted", blockNumberRule)
	if err != nil {
		return nil, err
	}
	return &RollupPubKeysPostedIterator{contract: _Rollup.contract, event: "PubKeysPosted", logs: logs, sub: sub}, nil
}

// WatchPubKeysPosted is a free log subscription operation binding the contract event 0x7a873ae11bb72d3848d5c34af1354cdd12aa5b30ec0d5187ec00ba5cefcd5c7f.
//
// Solidity: event PubKeysPosted(uint256 indexed blockNumber, uint256[] senderPublicKeys)
func (_Rollup *RollupFilterer) WatchPubKeysPosted(opts *bind.WatchOpts, sink chan<- *RollupPubKeysPosted, blockNumber []*big.Int) (event.Subscription, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}

	logs, sub, err := _Rollup.contract.WatchLogs(opts, "PubKeysPosted", blockNumberRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RollupPubKeysPosted)
				if err := _Rollup.contract.UnpackLog(event, "PubKeysPosted", log); err != nil {
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

// ParsePubKeysPosted is a log parse operation binding the contract event 0x7a873ae11bb72d3848d5c34af1354cdd12aa5b30ec0d5187ec00ba5cefcd5c7f.
//
// Solidity: event PubKeysPosted(uint256 indexed blockNumber, uint256[] senderPublicKeys)
func (_Rollup *RollupFilterer) ParsePubKeysPosted(log types.Log) (*RollupPubKeysPosted, error) {
	event := new(RollupPubKeysPosted)
	if err := _Rollup.contract.UnpackLog(event, "PubKeysPosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RollupUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Rollup contract.
type RollupUpgradedIterator struct {
	Event *RollupUpgraded // Event containing the contract specifics and raw log

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
func (it *RollupUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RollupUpgraded)
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
		it.Event = new(RollupUpgraded)
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
func (it *RollupUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RollupUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RollupUpgraded represents a Upgraded event raised by the Rollup contract.
type RollupUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Rollup *RollupFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*RollupUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Rollup.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &RollupUpgradedIterator{contract: _Rollup.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Rollup *RollupFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *RollupUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Rollup.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RollupUpgraded)
				if err := _Rollup.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_Rollup *RollupFilterer) ParseUpgraded(log types.Log) (*RollupUpgraded, error) {
	event := new(RollupUpgraded)
	if err := _Rollup.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
