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

// FraudProofPublicInputsLibFraudProofPublicInputs is an auto generated low-level Go binding around an user-defined struct.
type FraudProofPublicInputsLibFraudProofPublicInputs struct {
	BlockHash   [32]byte
	BlockNumber uint32
	Challenger  common.Address
}

// BlockBuilderRegistryMetaData contains all meta data concerning the BlockBuilderRegistry contract.
var BlockBuilderRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BlockBuilderNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CannotUnstakeWithinChallengeDuration\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"FailedTransfer\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FraudProofAlreadySubmitted\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"given\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"expected\",\"type\":\"bytes32\"}],\"name\":\"FraudProofBlockHashMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FraudProofChallengerMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FraudProofVerificationFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InsufficientStakeAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"challenger\",\"type\":\"address\"}],\"name\":\"BlockBuilderSlashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"}],\"name\":\"BlockBuilderStopped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"url\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"}],\"name\":\"BlockBuilderUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"challenger\",\"type\":\"address\"}],\"name\":\"BlockFraudProofSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"blockBuilders\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"blockBuilderUrl\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stopTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numSlashes\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isValid\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_rollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_fraudVerifier\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"}],\"name\":\"isValidBlockBuilder\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_burnAddress\",\"type\":\"address\"}],\"name\":\"setBurnAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stopBlockBuilder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"challenger\",\"type\":\"address\"}],\"internalType\":\"structFraudProofPublicInputsLib.FraudProofPublicInputs\",\"name\":\"publicInputs\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"submitBlockFraudProof\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"url\",\"type\":\"string\"}],\"name\":\"updateBlockBuilder\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// BlockBuilderRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use BlockBuilderRegistryMetaData.ABI instead.
var BlockBuilderRegistryABI = BlockBuilderRegistryMetaData.ABI

// BlockBuilderRegistry is an auto generated Go binding around an Ethereum contract.
type BlockBuilderRegistry struct {
	BlockBuilderRegistryCaller     // Read-only binding to the contract
	BlockBuilderRegistryTransactor // Write-only binding to the contract
	BlockBuilderRegistryFilterer   // Log filterer for contract events
}

// BlockBuilderRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type BlockBuilderRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BlockBuilderRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type BlockBuilderRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BlockBuilderRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type BlockBuilderRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// BlockBuilderRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type BlockBuilderRegistrySession struct {
	Contract     *BlockBuilderRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// BlockBuilderRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type BlockBuilderRegistryCallerSession struct {
	Contract *BlockBuilderRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// BlockBuilderRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type BlockBuilderRegistryTransactorSession struct {
	Contract     *BlockBuilderRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// BlockBuilderRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type BlockBuilderRegistryRaw struct {
	Contract *BlockBuilderRegistry // Generic contract binding to access the raw methods on
}

// BlockBuilderRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type BlockBuilderRegistryCallerRaw struct {
	Contract *BlockBuilderRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// BlockBuilderRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type BlockBuilderRegistryTransactorRaw struct {
	Contract *BlockBuilderRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewBlockBuilderRegistry creates a new instance of BlockBuilderRegistry, bound to a specific deployed contract.
func NewBlockBuilderRegistry(address common.Address, backend bind.ContractBackend) (*BlockBuilderRegistry, error) {
	contract, err := bindBlockBuilderRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistry{BlockBuilderRegistryCaller: BlockBuilderRegistryCaller{contract: contract}, BlockBuilderRegistryTransactor: BlockBuilderRegistryTransactor{contract: contract}, BlockBuilderRegistryFilterer: BlockBuilderRegistryFilterer{contract: contract}}, nil
}

// NewBlockBuilderRegistryCaller creates a new read-only instance of BlockBuilderRegistry, bound to a specific deployed contract.
func NewBlockBuilderRegistryCaller(address common.Address, caller bind.ContractCaller) (*BlockBuilderRegistryCaller, error) {
	contract, err := bindBlockBuilderRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryCaller{contract: contract}, nil
}

// NewBlockBuilderRegistryTransactor creates a new write-only instance of BlockBuilderRegistry, bound to a specific deployed contract.
func NewBlockBuilderRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*BlockBuilderRegistryTransactor, error) {
	contract, err := bindBlockBuilderRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryTransactor{contract: contract}, nil
}

// NewBlockBuilderRegistryFilterer creates a new log filterer instance of BlockBuilderRegistry, bound to a specific deployed contract.
func NewBlockBuilderRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*BlockBuilderRegistryFilterer, error) {
	contract, err := bindBlockBuilderRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryFilterer{contract: contract}, nil
}

// bindBlockBuilderRegistry binds a generic wrapper to an already deployed contract.
func bindBlockBuilderRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := BlockBuilderRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BlockBuilderRegistry *BlockBuilderRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BlockBuilderRegistry.Contract.BlockBuilderRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BlockBuilderRegistry *BlockBuilderRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.BlockBuilderRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BlockBuilderRegistry *BlockBuilderRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.BlockBuilderRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _BlockBuilderRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.contract.Transact(opts, method, params...)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) UPGRADEINTERFACEVERSION() (string, error) {
	return _BlockBuilderRegistry.Contract.UPGRADEINTERFACEVERSION(&_BlockBuilderRegistry.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _BlockBuilderRegistry.Contract.UPGRADEINTERFACEVERSION(&_BlockBuilderRegistry.CallOpts)
}

// BlockBuilders is a free data retrieval call binding the contract method 0x63cebe62.
//
// Solidity: function blockBuilders(address ) view returns(string blockBuilderUrl, uint256 stakeAmount, uint256 stopTime, uint256 numSlashes, bool isValid)
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) BlockBuilders(opts *bind.CallOpts, arg0 common.Address) (struct {
	BlockBuilderUrl string
	StakeAmount     *big.Int
	StopTime        *big.Int
	NumSlashes      *big.Int
	IsValid         bool
}, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "blockBuilders", arg0)

	outstruct := new(struct {
		BlockBuilderUrl string
		StakeAmount     *big.Int
		StopTime        *big.Int
		NumSlashes      *big.Int
		IsValid         bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.BlockBuilderUrl = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.StakeAmount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.StopTime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.NumSlashes = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.IsValid = *abi.ConvertType(out[4], new(bool)).(*bool)

	return *outstruct, err

}

// BlockBuilders is a free data retrieval call binding the contract method 0x63cebe62.
//
// Solidity: function blockBuilders(address ) view returns(string blockBuilderUrl, uint256 stakeAmount, uint256 stopTime, uint256 numSlashes, bool isValid)
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) BlockBuilders(arg0 common.Address) (struct {
	BlockBuilderUrl string
	StakeAmount     *big.Int
	StopTime        *big.Int
	NumSlashes      *big.Int
	IsValid         bool
}, error) {
	return _BlockBuilderRegistry.Contract.BlockBuilders(&_BlockBuilderRegistry.CallOpts, arg0)
}

// BlockBuilders is a free data retrieval call binding the contract method 0x63cebe62.
//
// Solidity: function blockBuilders(address ) view returns(string blockBuilderUrl, uint256 stakeAmount, uint256 stopTime, uint256 numSlashes, bool isValid)
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) BlockBuilders(arg0 common.Address) (struct {
	BlockBuilderUrl string
	StakeAmount     *big.Int
	StopTime        *big.Int
	NumSlashes      *big.Int
	IsValid         bool
}, error) {
	return _BlockBuilderRegistry.Contract.BlockBuilders(&_BlockBuilderRegistry.CallOpts, arg0)
}

// IsValidBlockBuilder is a free data retrieval call binding the contract method 0x6a5387a8.
//
// Solidity: function isValidBlockBuilder(address blockBuilder) view returns(bool)
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) IsValidBlockBuilder(opts *bind.CallOpts, blockBuilder common.Address) (bool, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "isValidBlockBuilder", blockBuilder)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidBlockBuilder is a free data retrieval call binding the contract method 0x6a5387a8.
//
// Solidity: function isValidBlockBuilder(address blockBuilder) view returns(bool)
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) IsValidBlockBuilder(blockBuilder common.Address) (bool, error) {
	return _BlockBuilderRegistry.Contract.IsValidBlockBuilder(&_BlockBuilderRegistry.CallOpts, blockBuilder)
}

// IsValidBlockBuilder is a free data retrieval call binding the contract method 0x6a5387a8.
//
// Solidity: function isValidBlockBuilder(address blockBuilder) view returns(bool)
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) IsValidBlockBuilder(blockBuilder common.Address) (bool, error) {
	return _BlockBuilderRegistry.Contract.IsValidBlockBuilder(&_BlockBuilderRegistry.CallOpts, blockBuilder)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) Owner() (common.Address, error) {
	return _BlockBuilderRegistry.Contract.Owner(&_BlockBuilderRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) Owner() (common.Address, error) {
	return _BlockBuilderRegistry.Contract.Owner(&_BlockBuilderRegistry.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) ProxiableUUID() ([32]byte, error) {
	return _BlockBuilderRegistry.Contract.ProxiableUUID(&_BlockBuilderRegistry.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) ProxiableUUID() ([32]byte, error) {
	return _BlockBuilderRegistry.Contract.ProxiableUUID(&_BlockBuilderRegistry.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _rollup, address _fraudVerifier) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) Initialize(opts *bind.TransactOpts, _rollup common.Address, _fraudVerifier common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "initialize", _rollup, _fraudVerifier)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _rollup, address _fraudVerifier) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) Initialize(_rollup common.Address, _fraudVerifier common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.Initialize(&_BlockBuilderRegistry.TransactOpts, _rollup, _fraudVerifier)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _rollup, address _fraudVerifier) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) Initialize(_rollup common.Address, _fraudVerifier common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.Initialize(&_BlockBuilderRegistry.TransactOpts, _rollup, _fraudVerifier)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.RenounceOwnership(&_BlockBuilderRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.RenounceOwnership(&_BlockBuilderRegistry.TransactOpts)
}

// SetBurnAddress is a paid mutator transaction binding the contract method 0x4b0e7216.
//
// Solidity: function setBurnAddress(address _burnAddress) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) SetBurnAddress(opts *bind.TransactOpts, _burnAddress common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "setBurnAddress", _burnAddress)
}

// SetBurnAddress is a paid mutator transaction binding the contract method 0x4b0e7216.
//
// Solidity: function setBurnAddress(address _burnAddress) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) SetBurnAddress(_burnAddress common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.SetBurnAddress(&_BlockBuilderRegistry.TransactOpts, _burnAddress)
}

// SetBurnAddress is a paid mutator transaction binding the contract method 0x4b0e7216.
//
// Solidity: function setBurnAddress(address _burnAddress) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) SetBurnAddress(_burnAddress common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.SetBurnAddress(&_BlockBuilderRegistry.TransactOpts, _burnAddress)
}

// StopBlockBuilder is a paid mutator transaction binding the contract method 0x51334851.
//
// Solidity: function stopBlockBuilder() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) StopBlockBuilder(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "stopBlockBuilder")
}

// StopBlockBuilder is a paid mutator transaction binding the contract method 0x51334851.
//
// Solidity: function stopBlockBuilder() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) StopBlockBuilder() (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.StopBlockBuilder(&_BlockBuilderRegistry.TransactOpts)
}

// StopBlockBuilder is a paid mutator transaction binding the contract method 0x51334851.
//
// Solidity: function stopBlockBuilder() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) StopBlockBuilder() (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.StopBlockBuilder(&_BlockBuilderRegistry.TransactOpts)
}

// SubmitBlockFraudProof is a paid mutator transaction binding the contract method 0x0cffc39c.
//
// Solidity: function submitBlockFraudProof((bytes32,uint32,address) publicInputs, bytes proof) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) SubmitBlockFraudProof(opts *bind.TransactOpts, publicInputs FraudProofPublicInputsLibFraudProofPublicInputs, proof []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "submitBlockFraudProof", publicInputs, proof)
}

// SubmitBlockFraudProof is a paid mutator transaction binding the contract method 0x0cffc39c.
//
// Solidity: function submitBlockFraudProof((bytes32,uint32,address) publicInputs, bytes proof) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) SubmitBlockFraudProof(publicInputs FraudProofPublicInputsLibFraudProofPublicInputs, proof []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.SubmitBlockFraudProof(&_BlockBuilderRegistry.TransactOpts, publicInputs, proof)
}

// SubmitBlockFraudProof is a paid mutator transaction binding the contract method 0x0cffc39c.
//
// Solidity: function submitBlockFraudProof((bytes32,uint32,address) publicInputs, bytes proof) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) SubmitBlockFraudProof(publicInputs FraudProofPublicInputsLibFraudProofPublicInputs, proof []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.SubmitBlockFraudProof(&_BlockBuilderRegistry.TransactOpts, publicInputs, proof)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.TransferOwnership(&_BlockBuilderRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.TransferOwnership(&_BlockBuilderRegistry.TransactOpts, newOwner)
}

// Unstake is a paid mutator transaction binding the contract method 0x2def6620.
//
// Solidity: function unstake() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) Unstake(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "unstake")
}

// Unstake is a paid mutator transaction binding the contract method 0x2def6620.
//
// Solidity: function unstake() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) Unstake() (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.Unstake(&_BlockBuilderRegistry.TransactOpts)
}

// Unstake is a paid mutator transaction binding the contract method 0x2def6620.
//
// Solidity: function unstake() returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) Unstake() (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.Unstake(&_BlockBuilderRegistry.TransactOpts)
}

// UpdateBlockBuilder is a paid mutator transaction binding the contract method 0x0f36c638.
//
// Solidity: function updateBlockBuilder(string url) payable returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) UpdateBlockBuilder(opts *bind.TransactOpts, url string) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "updateBlockBuilder", url)
}

// UpdateBlockBuilder is a paid mutator transaction binding the contract method 0x0f36c638.
//
// Solidity: function updateBlockBuilder(string url) payable returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) UpdateBlockBuilder(url string) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.UpdateBlockBuilder(&_BlockBuilderRegistry.TransactOpts, url)
}

// UpdateBlockBuilder is a paid mutator transaction binding the contract method 0x0f36c638.
//
// Solidity: function updateBlockBuilder(string url) payable returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) UpdateBlockBuilder(url string) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.UpdateBlockBuilder(&_BlockBuilderRegistry.TransactOpts, url)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.UpgradeToAndCall(&_BlockBuilderRegistry.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.UpgradeToAndCall(&_BlockBuilderRegistry.TransactOpts, newImplementation, data)
}

// BlockBuilderRegistryBlockBuilderSlashedIterator is returned from FilterBlockBuilderSlashed and is used to iterate over the raw logs and unpacked data for BlockBuilderSlashed events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderSlashedIterator struct {
	Event *BlockBuilderRegistryBlockBuilderSlashed // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryBlockBuilderSlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryBlockBuilderSlashed)
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
		it.Event = new(BlockBuilderRegistryBlockBuilderSlashed)
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
func (it *BlockBuilderRegistryBlockBuilderSlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryBlockBuilderSlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryBlockBuilderSlashed represents a BlockBuilderSlashed event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderSlashed struct {
	BlockBuilder common.Address
	Challenger   common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBlockBuilderSlashed is a free log retrieval operation binding the contract event 0xbebe1ecc1c646675c6bd0423039fd8cdde200d8e4dac2b345abdc23bf8d97a69.
//
// Solidity: event BlockBuilderSlashed(address indexed blockBuilder, address indexed challenger)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterBlockBuilderSlashed(opts *bind.FilterOpts, blockBuilder []common.Address, challenger []common.Address) (*BlockBuilderRegistryBlockBuilderSlashedIterator, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}
	var challengerRule []interface{}
	for _, challengerItem := range challenger {
		challengerRule = append(challengerRule, challengerItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "BlockBuilderSlashed", blockBuilderRule, challengerRule)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryBlockBuilderSlashedIterator{contract: _BlockBuilderRegistry.contract, event: "BlockBuilderSlashed", logs: logs, sub: sub}, nil
}

// WatchBlockBuilderSlashed is a free log subscription operation binding the contract event 0xbebe1ecc1c646675c6bd0423039fd8cdde200d8e4dac2b345abdc23bf8d97a69.
//
// Solidity: event BlockBuilderSlashed(address indexed blockBuilder, address indexed challenger)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchBlockBuilderSlashed(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryBlockBuilderSlashed, blockBuilder []common.Address, challenger []common.Address) (event.Subscription, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}
	var challengerRule []interface{}
	for _, challengerItem := range challenger {
		challengerRule = append(challengerRule, challengerItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "BlockBuilderSlashed", blockBuilderRule, challengerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryBlockBuilderSlashed)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderSlashed", log); err != nil {
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

// ParseBlockBuilderSlashed is a log parse operation binding the contract event 0xbebe1ecc1c646675c6bd0423039fd8cdde200d8e4dac2b345abdc23bf8d97a69.
//
// Solidity: event BlockBuilderSlashed(address indexed blockBuilder, address indexed challenger)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseBlockBuilderSlashed(log types.Log) (*BlockBuilderRegistryBlockBuilderSlashed, error) {
	event := new(BlockBuilderRegistryBlockBuilderSlashed)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderSlashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BlockBuilderRegistryBlockBuilderStoppedIterator is returned from FilterBlockBuilderStopped and is used to iterate over the raw logs and unpacked data for BlockBuilderStopped events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderStoppedIterator struct {
	Event *BlockBuilderRegistryBlockBuilderStopped // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryBlockBuilderStoppedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryBlockBuilderStopped)
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
		it.Event = new(BlockBuilderRegistryBlockBuilderStopped)
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
func (it *BlockBuilderRegistryBlockBuilderStoppedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryBlockBuilderStoppedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryBlockBuilderStopped represents a BlockBuilderStopped event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderStopped struct {
	BlockBuilder common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBlockBuilderStopped is a free log retrieval operation binding the contract event 0x61003f58033bb3da9d31643101595ff5f23aead3167e68857530f74c4430582b.
//
// Solidity: event BlockBuilderStopped(address indexed blockBuilder)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterBlockBuilderStopped(opts *bind.FilterOpts, blockBuilder []common.Address) (*BlockBuilderRegistryBlockBuilderStoppedIterator, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "BlockBuilderStopped", blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryBlockBuilderStoppedIterator{contract: _BlockBuilderRegistry.contract, event: "BlockBuilderStopped", logs: logs, sub: sub}, nil
}

// WatchBlockBuilderStopped is a free log subscription operation binding the contract event 0x61003f58033bb3da9d31643101595ff5f23aead3167e68857530f74c4430582b.
//
// Solidity: event BlockBuilderStopped(address indexed blockBuilder)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchBlockBuilderStopped(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryBlockBuilderStopped, blockBuilder []common.Address) (event.Subscription, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "BlockBuilderStopped", blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryBlockBuilderStopped)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderStopped", log); err != nil {
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

// ParseBlockBuilderStopped is a log parse operation binding the contract event 0x61003f58033bb3da9d31643101595ff5f23aead3167e68857530f74c4430582b.
//
// Solidity: event BlockBuilderStopped(address indexed blockBuilder)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseBlockBuilderStopped(log types.Log) (*BlockBuilderRegistryBlockBuilderStopped, error) {
	event := new(BlockBuilderRegistryBlockBuilderStopped)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderStopped", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BlockBuilderRegistryBlockBuilderUpdatedIterator is returned from FilterBlockBuilderUpdated and is used to iterate over the raw logs and unpacked data for BlockBuilderUpdated events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderUpdatedIterator struct {
	Event *BlockBuilderRegistryBlockBuilderUpdated // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryBlockBuilderUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryBlockBuilderUpdated)
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
		it.Event = new(BlockBuilderRegistryBlockBuilderUpdated)
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
func (it *BlockBuilderRegistryBlockBuilderUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryBlockBuilderUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryBlockBuilderUpdated represents a BlockBuilderUpdated event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderUpdated struct {
	BlockBuilder common.Address
	Url          string
	StakeAmount  *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBlockBuilderUpdated is a free log retrieval operation binding the contract event 0x2e6a62283ed761c959aaa67da7e98c90c08118cadc35aae3c7b9d0b258a54405.
//
// Solidity: event BlockBuilderUpdated(address indexed blockBuilder, string url, uint256 stakeAmount)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterBlockBuilderUpdated(opts *bind.FilterOpts, blockBuilder []common.Address) (*BlockBuilderRegistryBlockBuilderUpdatedIterator, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "BlockBuilderUpdated", blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryBlockBuilderUpdatedIterator{contract: _BlockBuilderRegistry.contract, event: "BlockBuilderUpdated", logs: logs, sub: sub}, nil
}

// WatchBlockBuilderUpdated is a free log subscription operation binding the contract event 0x2e6a62283ed761c959aaa67da7e98c90c08118cadc35aae3c7b9d0b258a54405.
//
// Solidity: event BlockBuilderUpdated(address indexed blockBuilder, string url, uint256 stakeAmount)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchBlockBuilderUpdated(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryBlockBuilderUpdated, blockBuilder []common.Address) (event.Subscription, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "BlockBuilderUpdated", blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryBlockBuilderUpdated)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderUpdated", log); err != nil {
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

// ParseBlockBuilderUpdated is a log parse operation binding the contract event 0x2e6a62283ed761c959aaa67da7e98c90c08118cadc35aae3c7b9d0b258a54405.
//
// Solidity: event BlockBuilderUpdated(address indexed blockBuilder, string url, uint256 stakeAmount)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseBlockBuilderUpdated(log types.Log) (*BlockBuilderRegistryBlockBuilderUpdated, error) {
	event := new(BlockBuilderRegistryBlockBuilderUpdated)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BlockBuilderRegistryBlockFraudProofSubmittedIterator is returned from FilterBlockFraudProofSubmitted and is used to iterate over the raw logs and unpacked data for BlockFraudProofSubmitted events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockFraudProofSubmittedIterator struct {
	Event *BlockBuilderRegistryBlockFraudProofSubmitted // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryBlockFraudProofSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryBlockFraudProofSubmitted)
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
		it.Event = new(BlockBuilderRegistryBlockFraudProofSubmitted)
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
func (it *BlockBuilderRegistryBlockFraudProofSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryBlockFraudProofSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryBlockFraudProofSubmitted represents a BlockFraudProofSubmitted event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockFraudProofSubmitted struct {
	BlockNumber  uint32
	BlockBuilder common.Address
	Challenger   common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBlockFraudProofSubmitted is a free log retrieval operation binding the contract event 0xa2d160ed24716f0240546105d30a2e1efb2d5f23abb041888083bcc20083edd2.
//
// Solidity: event BlockFraudProofSubmitted(uint32 indexed blockNumber, address indexed blockBuilder, address indexed challenger)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterBlockFraudProofSubmitted(opts *bind.FilterOpts, blockNumber []uint32, blockBuilder []common.Address, challenger []common.Address) (*BlockBuilderRegistryBlockFraudProofSubmittedIterator, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}
	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}
	var challengerRule []interface{}
	for _, challengerItem := range challenger {
		challengerRule = append(challengerRule, challengerItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "BlockFraudProofSubmitted", blockNumberRule, blockBuilderRule, challengerRule)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryBlockFraudProofSubmittedIterator{contract: _BlockBuilderRegistry.contract, event: "BlockFraudProofSubmitted", logs: logs, sub: sub}, nil
}

// WatchBlockFraudProofSubmitted is a free log subscription operation binding the contract event 0xa2d160ed24716f0240546105d30a2e1efb2d5f23abb041888083bcc20083edd2.
//
// Solidity: event BlockFraudProofSubmitted(uint32 indexed blockNumber, address indexed blockBuilder, address indexed challenger)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchBlockFraudProofSubmitted(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryBlockFraudProofSubmitted, blockNumber []uint32, blockBuilder []common.Address, challenger []common.Address) (event.Subscription, error) {

	var blockNumberRule []interface{}
	for _, blockNumberItem := range blockNumber {
		blockNumberRule = append(blockNumberRule, blockNumberItem)
	}
	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}
	var challengerRule []interface{}
	for _, challengerItem := range challenger {
		challengerRule = append(challengerRule, challengerItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "BlockFraudProofSubmitted", blockNumberRule, blockBuilderRule, challengerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryBlockFraudProofSubmitted)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockFraudProofSubmitted", log); err != nil {
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

// ParseBlockFraudProofSubmitted is a log parse operation binding the contract event 0xa2d160ed24716f0240546105d30a2e1efb2d5f23abb041888083bcc20083edd2.
//
// Solidity: event BlockFraudProofSubmitted(uint32 indexed blockNumber, address indexed blockBuilder, address indexed challenger)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseBlockFraudProofSubmitted(log types.Log) (*BlockBuilderRegistryBlockFraudProofSubmitted, error) {
	event := new(BlockBuilderRegistryBlockFraudProofSubmitted)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockFraudProofSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BlockBuilderRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryInitializedIterator struct {
	Event *BlockBuilderRegistryInitialized // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryInitialized)
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
		it.Event = new(BlockBuilderRegistryInitialized)
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
func (it *BlockBuilderRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryInitialized represents a Initialized event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*BlockBuilderRegistryInitializedIterator, error) {

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryInitializedIterator{contract: _BlockBuilderRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryInitialized)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseInitialized(log types.Log) (*BlockBuilderRegistryInitialized, error) {
	event := new(BlockBuilderRegistryInitialized)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BlockBuilderRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryOwnershipTransferredIterator struct {
	Event *BlockBuilderRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryOwnershipTransferred)
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
		it.Event = new(BlockBuilderRegistryOwnershipTransferred)
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
func (it *BlockBuilderRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*BlockBuilderRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryOwnershipTransferredIterator{contract: _BlockBuilderRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryOwnershipTransferred)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*BlockBuilderRegistryOwnershipTransferred, error) {
	event := new(BlockBuilderRegistryOwnershipTransferred)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// BlockBuilderRegistryUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryUpgradedIterator struct {
	Event *BlockBuilderRegistryUpgraded // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryUpgraded)
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
		it.Event = new(BlockBuilderRegistryUpgraded)
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
func (it *BlockBuilderRegistryUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryUpgraded represents a Upgraded event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*BlockBuilderRegistryUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryUpgradedIterator{contract: _BlockBuilderRegistry.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryUpgraded)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseUpgraded(log types.Log) (*BlockBuilderRegistryUpgraded, error) {
	event := new(BlockBuilderRegistryUpgraded)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
