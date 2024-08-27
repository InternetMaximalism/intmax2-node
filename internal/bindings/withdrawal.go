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

// ChainedWithdrawalLibChainedWithdrawal is an auto generated low-level Go binding around an user-defined struct.
type ChainedWithdrawalLibChainedWithdrawal struct {
	Recipient   common.Address
	TokenIndex  uint32
	Amount      *big.Int
	Nullifier   [32]byte
	BlockHash   [32]byte
	BlockNumber uint32
}

// WithdrawalLibWithdrawal is an auto generated low-level Go binding around an user-defined struct.
type WithdrawalLibWithdrawal struct {
	Recipient  common.Address
	TokenIndex uint32
	Amount     *big.Int
	Id         *big.Int
}

// WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs is an auto generated low-level Go binding around an user-defined struct.
type WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs struct {
	LastWithdrawalHash   [32]byte
	WithdrawalAggregator common.Address
}

// WithdrawalMetaData contains all meta data concerning the Withdrawal contract.
var WithdrawalMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"AddressEmptyCode\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"}],\"name\":\"BlockHashNotExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"claimableWithdrawalId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rear\",\"type\":\"uint256\"}],\"name\":\"ClaimableWithdrawalIsTooLarge\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"directWithdrawalId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rear\",\"type\":\"uint256\"}],\"name\":\"DirectWithdrawalIsTooLarge\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"ERC1967InvalidImplementation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ERC1967NonPayable\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"FailedInnerCall\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidInitialization\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotInitializing\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenIndice\",\"type\":\"uint256\"}],\"name\":\"TokenAlreadyExist\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenIndice\",\"type\":\"uint256\"}],\"name\":\"TokenNotExist\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"count\",\"type\":\"uint256\"}],\"name\":\"TooManyRelayClaimableWithdrawals\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"count\",\"type\":\"uint256\"}],\"name\":\"TooManyRelayDirectWithdrawals\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"UUPSUnauthorizedCallContext\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"slot\",\"type\":\"bytes32\"}],\"name\":\"UUPSUnsupportedProxiableUUID\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawalAggregatorMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requestIndex\",\"type\":\"uint256\"}],\"name\":\"WithdrawalBlockHashNotPosted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawalChainVerificationFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawalProofVerificationFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawalsHashMismatch\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"claimableWithdrawalId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structWithdrawalLib.Withdrawal\",\"name\":\"withdrawal\",\"type\":\"tuple\"}],\"name\":\"ClaimableWithdrawalQueued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"directWithdrawalId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"indexed\":false,\"internalType\":\"structWithdrawalLib.Withdrawal\",\"name\":\"withdrawal\",\"type\":\"tuple\"}],\"name\":\"DirectWithdrawalQueued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"version\",\"type\":\"uint64\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lastDirectWithdrawalId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"lastClaimableWithdrawalId\",\"type\":\"uint256\"}],\"name\":\"WithdrawalsQueued\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"UPGRADE_INTERFACE_VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"tokenIndices\",\"type\":\"uint256[]\"}],\"name\":\"addDirectWithdrawalTokenIndices\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDirectWithdrawalTokenIndices\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_scrollMessenger\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_withdrawalVerifier\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_liquidity\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_rollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_contribution\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"_directWithdrawalTokenIndices\",\"type\":\"uint256[]\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastClaimableWithdrawalId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastDirectWithdrawalId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxiableUUID\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"tokenIndices\",\"type\":\"uint256[]\"}],\"name\":\"removeDirectWithdrawalTokenIndices\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"nullifier\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"blockHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"}],\"internalType\":\"structChainedWithdrawalLib.ChainedWithdrawal[]\",\"name\":\"withdrawals\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"lastWithdrawalHash\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"withdrawalAggregator\",\"type\":\"address\"}],\"internalType\":\"structWithdrawalProofPublicInputsLib.WithdrawalProofPublicInputs\",\"name\":\"publicInputs\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"submitWithdrawalProof\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newImplementation\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"upgradeToAndCall\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
}

// WithdrawalABI is the input ABI used to generate the binding from.
// Deprecated: Use WithdrawalMetaData.ABI instead.
var WithdrawalABI = WithdrawalMetaData.ABI

// Withdrawal is an auto generated Go binding around an Ethereum contract.
type Withdrawal struct {
	WithdrawalCaller     // Read-only binding to the contract
	WithdrawalTransactor // Write-only binding to the contract
	WithdrawalFilterer   // Log filterer for contract events
}

// WithdrawalCaller is an auto generated read-only Go binding around an Ethereum contract.
type WithdrawalCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawalTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WithdrawalTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawalFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type WithdrawalFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawalSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WithdrawalSession struct {
	Contract     *Withdrawal       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// WithdrawalCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WithdrawalCallerSession struct {
	Contract *WithdrawalCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// WithdrawalTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WithdrawalTransactorSession struct {
	Contract     *WithdrawalTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// WithdrawalRaw is an auto generated low-level Go binding around an Ethereum contract.
type WithdrawalRaw struct {
	Contract *Withdrawal // Generic contract binding to access the raw methods on
}

// WithdrawalCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WithdrawalCallerRaw struct {
	Contract *WithdrawalCaller // Generic read-only contract binding to access the raw methods on
}

// WithdrawalTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WithdrawalTransactorRaw struct {
	Contract *WithdrawalTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWithdrawal creates a new instance of Withdrawal, bound to a specific deployed contract.
func NewWithdrawal(address common.Address, backend bind.ContractBackend) (*Withdrawal, error) {
	contract, err := bindWithdrawal(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Withdrawal{WithdrawalCaller: WithdrawalCaller{contract: contract}, WithdrawalTransactor: WithdrawalTransactor{contract: contract}, WithdrawalFilterer: WithdrawalFilterer{contract: contract}}, nil
}

// NewWithdrawalCaller creates a new read-only instance of Withdrawal, bound to a specific deployed contract.
func NewWithdrawalCaller(address common.Address, caller bind.ContractCaller) (*WithdrawalCaller, error) {
	contract, err := bindWithdrawal(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &WithdrawalCaller{contract: contract}, nil
}

// NewWithdrawalTransactor creates a new write-only instance of Withdrawal, bound to a specific deployed contract.
func NewWithdrawalTransactor(address common.Address, transactor bind.ContractTransactor) (*WithdrawalTransactor, error) {
	contract, err := bindWithdrawal(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &WithdrawalTransactor{contract: contract}, nil
}

// NewWithdrawalFilterer creates a new log filterer instance of Withdrawal, bound to a specific deployed contract.
func NewWithdrawalFilterer(address common.Address, filterer bind.ContractFilterer) (*WithdrawalFilterer, error) {
	contract, err := bindWithdrawal(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &WithdrawalFilterer{contract: contract}, nil
}

// bindWithdrawal binds a generic wrapper to an already deployed contract.
func bindWithdrawal(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := WithdrawalMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Withdrawal *WithdrawalRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Withdrawal.Contract.WithdrawalCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Withdrawal *WithdrawalRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawal.Contract.WithdrawalTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Withdrawal *WithdrawalRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Withdrawal.Contract.WithdrawalTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Withdrawal *WithdrawalCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Withdrawal.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Withdrawal *WithdrawalTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawal.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Withdrawal *WithdrawalTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Withdrawal.Contract.contract.Transact(opts, method, params...)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Withdrawal *WithdrawalCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Withdrawal.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Withdrawal *WithdrawalSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Withdrawal.Contract.UPGRADEINTERFACEVERSION(&_Withdrawal.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_Withdrawal *WithdrawalCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _Withdrawal.Contract.UPGRADEINTERFACEVERSION(&_Withdrawal.CallOpts)
}

// GetDirectWithdrawalTokenIndices is a free data retrieval call binding the contract method 0xa7178c37.
//
// Solidity: function getDirectWithdrawalTokenIndices() view returns(uint256[])
func (_Withdrawal *WithdrawalCaller) GetDirectWithdrawalTokenIndices(opts *bind.CallOpts) ([]*big.Int, error) {
	var out []interface{}
	err := _Withdrawal.contract.Call(opts, &out, "getDirectWithdrawalTokenIndices")

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// GetDirectWithdrawalTokenIndices is a free data retrieval call binding the contract method 0xa7178c37.
//
// Solidity: function getDirectWithdrawalTokenIndices() view returns(uint256[])
func (_Withdrawal *WithdrawalSession) GetDirectWithdrawalTokenIndices() ([]*big.Int, error) {
	return _Withdrawal.Contract.GetDirectWithdrawalTokenIndices(&_Withdrawal.CallOpts)
}

// GetDirectWithdrawalTokenIndices is a free data retrieval call binding the contract method 0xa7178c37.
//
// Solidity: function getDirectWithdrawalTokenIndices() view returns(uint256[])
func (_Withdrawal *WithdrawalCallerSession) GetDirectWithdrawalTokenIndices() ([]*big.Int, error) {
	return _Withdrawal.Contract.GetDirectWithdrawalTokenIndices(&_Withdrawal.CallOpts)
}

// LastClaimableWithdrawalId is a free data retrieval call binding the contract method 0x3c3377a5.
//
// Solidity: function lastClaimableWithdrawalId() view returns(uint256)
func (_Withdrawal *WithdrawalCaller) LastClaimableWithdrawalId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Withdrawal.contract.Call(opts, &out, "lastClaimableWithdrawalId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastClaimableWithdrawalId is a free data retrieval call binding the contract method 0x3c3377a5.
//
// Solidity: function lastClaimableWithdrawalId() view returns(uint256)
func (_Withdrawal *WithdrawalSession) LastClaimableWithdrawalId() (*big.Int, error) {
	return _Withdrawal.Contract.LastClaimableWithdrawalId(&_Withdrawal.CallOpts)
}

// LastClaimableWithdrawalId is a free data retrieval call binding the contract method 0x3c3377a5.
//
// Solidity: function lastClaimableWithdrawalId() view returns(uint256)
func (_Withdrawal *WithdrawalCallerSession) LastClaimableWithdrawalId() (*big.Int, error) {
	return _Withdrawal.Contract.LastClaimableWithdrawalId(&_Withdrawal.CallOpts)
}

// LastDirectWithdrawalId is a free data retrieval call binding the contract method 0x705aeb47.
//
// Solidity: function lastDirectWithdrawalId() view returns(uint256)
func (_Withdrawal *WithdrawalCaller) LastDirectWithdrawalId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Withdrawal.contract.Call(opts, &out, "lastDirectWithdrawalId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastDirectWithdrawalId is a free data retrieval call binding the contract method 0x705aeb47.
//
// Solidity: function lastDirectWithdrawalId() view returns(uint256)
func (_Withdrawal *WithdrawalSession) LastDirectWithdrawalId() (*big.Int, error) {
	return _Withdrawal.Contract.LastDirectWithdrawalId(&_Withdrawal.CallOpts)
}

// LastDirectWithdrawalId is a free data retrieval call binding the contract method 0x705aeb47.
//
// Solidity: function lastDirectWithdrawalId() view returns(uint256)
func (_Withdrawal *WithdrawalCallerSession) LastDirectWithdrawalId() (*big.Int, error) {
	return _Withdrawal.Contract.LastDirectWithdrawalId(&_Withdrawal.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Withdrawal *WithdrawalCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Withdrawal.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Withdrawal *WithdrawalSession) Owner() (common.Address, error) {
	return _Withdrawal.Contract.Owner(&_Withdrawal.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Withdrawal *WithdrawalCallerSession) Owner() (common.Address, error) {
	return _Withdrawal.Contract.Owner(&_Withdrawal.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Withdrawal *WithdrawalCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Withdrawal.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Withdrawal *WithdrawalSession) ProxiableUUID() ([32]byte, error) {
	return _Withdrawal.Contract.ProxiableUUID(&_Withdrawal.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_Withdrawal *WithdrawalCallerSession) ProxiableUUID() ([32]byte, error) {
	return _Withdrawal.Contract.ProxiableUUID(&_Withdrawal.CallOpts)
}

// AddDirectWithdrawalTokenIndices is a paid mutator transaction binding the contract method 0x871c5fe7.
//
// Solidity: function addDirectWithdrawalTokenIndices(uint256[] tokenIndices) returns()
func (_Withdrawal *WithdrawalTransactor) AddDirectWithdrawalTokenIndices(opts *bind.TransactOpts, tokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.contract.Transact(opts, "addDirectWithdrawalTokenIndices", tokenIndices)
}

// AddDirectWithdrawalTokenIndices is a paid mutator transaction binding the contract method 0x871c5fe7.
//
// Solidity: function addDirectWithdrawalTokenIndices(uint256[] tokenIndices) returns()
func (_Withdrawal *WithdrawalSession) AddDirectWithdrawalTokenIndices(tokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.Contract.AddDirectWithdrawalTokenIndices(&_Withdrawal.TransactOpts, tokenIndices)
}

// AddDirectWithdrawalTokenIndices is a paid mutator transaction binding the contract method 0x871c5fe7.
//
// Solidity: function addDirectWithdrawalTokenIndices(uint256[] tokenIndices) returns()
func (_Withdrawal *WithdrawalTransactorSession) AddDirectWithdrawalTokenIndices(tokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.Contract.AddDirectWithdrawalTokenIndices(&_Withdrawal.TransactOpts, tokenIndices)
}

// Initialize is a paid mutator transaction binding the contract method 0xf05d0982.
//
// Solidity: function initialize(address _scrollMessenger, address _withdrawalVerifier, address _liquidity, address _rollup, address _contribution, uint256[] _directWithdrawalTokenIndices) returns()
func (_Withdrawal *WithdrawalTransactor) Initialize(opts *bind.TransactOpts, _scrollMessenger common.Address, _withdrawalVerifier common.Address, _liquidity common.Address, _rollup common.Address, _contribution common.Address, _directWithdrawalTokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.contract.Transact(opts, "initialize", _scrollMessenger, _withdrawalVerifier, _liquidity, _rollup, _contribution, _directWithdrawalTokenIndices)
}

// Initialize is a paid mutator transaction binding the contract method 0xf05d0982.
//
// Solidity: function initialize(address _scrollMessenger, address _withdrawalVerifier, address _liquidity, address _rollup, address _contribution, uint256[] _directWithdrawalTokenIndices) returns()
func (_Withdrawal *WithdrawalSession) Initialize(_scrollMessenger common.Address, _withdrawalVerifier common.Address, _liquidity common.Address, _rollup common.Address, _contribution common.Address, _directWithdrawalTokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.Contract.Initialize(&_Withdrawal.TransactOpts, _scrollMessenger, _withdrawalVerifier, _liquidity, _rollup, _contribution, _directWithdrawalTokenIndices)
}

// Initialize is a paid mutator transaction binding the contract method 0xf05d0982.
//
// Solidity: function initialize(address _scrollMessenger, address _withdrawalVerifier, address _liquidity, address _rollup, address _contribution, uint256[] _directWithdrawalTokenIndices) returns()
func (_Withdrawal *WithdrawalTransactorSession) Initialize(_scrollMessenger common.Address, _withdrawalVerifier common.Address, _liquidity common.Address, _rollup common.Address, _contribution common.Address, _directWithdrawalTokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.Contract.Initialize(&_Withdrawal.TransactOpts, _scrollMessenger, _withdrawalVerifier, _liquidity, _rollup, _contribution, _directWithdrawalTokenIndices)
}

// RemoveDirectWithdrawalTokenIndices is a paid mutator transaction binding the contract method 0x9a0002a9.
//
// Solidity: function removeDirectWithdrawalTokenIndices(uint256[] tokenIndices) returns()
func (_Withdrawal *WithdrawalTransactor) RemoveDirectWithdrawalTokenIndices(opts *bind.TransactOpts, tokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.contract.Transact(opts, "removeDirectWithdrawalTokenIndices", tokenIndices)
}

// RemoveDirectWithdrawalTokenIndices is a paid mutator transaction binding the contract method 0x9a0002a9.
//
// Solidity: function removeDirectWithdrawalTokenIndices(uint256[] tokenIndices) returns()
func (_Withdrawal *WithdrawalSession) RemoveDirectWithdrawalTokenIndices(tokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.Contract.RemoveDirectWithdrawalTokenIndices(&_Withdrawal.TransactOpts, tokenIndices)
}

// RemoveDirectWithdrawalTokenIndices is a paid mutator transaction binding the contract method 0x9a0002a9.
//
// Solidity: function removeDirectWithdrawalTokenIndices(uint256[] tokenIndices) returns()
func (_Withdrawal *WithdrawalTransactorSession) RemoveDirectWithdrawalTokenIndices(tokenIndices []*big.Int) (*types.Transaction, error) {
	return _Withdrawal.Contract.RemoveDirectWithdrawalTokenIndices(&_Withdrawal.TransactOpts, tokenIndices)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Withdrawal *WithdrawalTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawal.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Withdrawal *WithdrawalSession) RenounceOwnership() (*types.Transaction, error) {
	return _Withdrawal.Contract.RenounceOwnership(&_Withdrawal.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Withdrawal *WithdrawalTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Withdrawal.Contract.RenounceOwnership(&_Withdrawal.TransactOpts)
}

// SubmitWithdrawalProof is a paid mutator transaction binding the contract method 0xa93e83a6.
//
// Solidity: function submitWithdrawalProof((address,uint32,uint256,bytes32,bytes32,uint32)[] withdrawals, (bytes32,address) publicInputs, bytes proof) returns()
func (_Withdrawal *WithdrawalTransactor) SubmitWithdrawalProof(opts *bind.TransactOpts, withdrawals []ChainedWithdrawalLibChainedWithdrawal, publicInputs WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs, proof []byte) (*types.Transaction, error) {
	return _Withdrawal.contract.Transact(opts, "submitWithdrawalProof", withdrawals, publicInputs, proof)
}

// SubmitWithdrawalProof is a paid mutator transaction binding the contract method 0xa93e83a6.
//
// Solidity: function submitWithdrawalProof((address,uint32,uint256,bytes32,bytes32,uint32)[] withdrawals, (bytes32,address) publicInputs, bytes proof) returns()
func (_Withdrawal *WithdrawalSession) SubmitWithdrawalProof(withdrawals []ChainedWithdrawalLibChainedWithdrawal, publicInputs WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs, proof []byte) (*types.Transaction, error) {
	return _Withdrawal.Contract.SubmitWithdrawalProof(&_Withdrawal.TransactOpts, withdrawals, publicInputs, proof)
}

// SubmitWithdrawalProof is a paid mutator transaction binding the contract method 0xa93e83a6.
//
// Solidity: function submitWithdrawalProof((address,uint32,uint256,bytes32,bytes32,uint32)[] withdrawals, (bytes32,address) publicInputs, bytes proof) returns()
func (_Withdrawal *WithdrawalTransactorSession) SubmitWithdrawalProof(withdrawals []ChainedWithdrawalLibChainedWithdrawal, publicInputs WithdrawalProofPublicInputsLibWithdrawalProofPublicInputs, proof []byte) (*types.Transaction, error) {
	return _Withdrawal.Contract.SubmitWithdrawalProof(&_Withdrawal.TransactOpts, withdrawals, publicInputs, proof)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Withdrawal *WithdrawalTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Withdrawal.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Withdrawal *WithdrawalSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Withdrawal.Contract.TransferOwnership(&_Withdrawal.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Withdrawal *WithdrawalTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Withdrawal.Contract.TransferOwnership(&_Withdrawal.TransactOpts, newOwner)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Withdrawal *WithdrawalTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Withdrawal.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Withdrawal *WithdrawalSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Withdrawal.Contract.UpgradeToAndCall(&_Withdrawal.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_Withdrawal *WithdrawalTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _Withdrawal.Contract.UpgradeToAndCall(&_Withdrawal.TransactOpts, newImplementation, data)
}

// WithdrawalClaimableWithdrawalQueuedIterator is returned from FilterClaimableWithdrawalQueued and is used to iterate over the raw logs and unpacked data for ClaimableWithdrawalQueued events raised by the Withdrawal contract.
type WithdrawalClaimableWithdrawalQueuedIterator struct {
	Event *WithdrawalClaimableWithdrawalQueued // Event containing the contract specifics and raw log

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
func (it *WithdrawalClaimableWithdrawalQueuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalClaimableWithdrawalQueued)
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
		it.Event = new(WithdrawalClaimableWithdrawalQueued)
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
func (it *WithdrawalClaimableWithdrawalQueuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalClaimableWithdrawalQueuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalClaimableWithdrawalQueued represents a ClaimableWithdrawalQueued event raised by the Withdrawal contract.
type WithdrawalClaimableWithdrawalQueued struct {
	ClaimableWithdrawalId *big.Int
	Recipient             common.Address
	Withdrawal            WithdrawalLibWithdrawal
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterClaimableWithdrawalQueued is a free log retrieval operation binding the contract event 0xa82167f0d0baa7a4ee5cec53c8b897a8de2debb946fc5e79d62e3c502396b1e1.
//
// Solidity: event ClaimableWithdrawalQueued(uint256 indexed claimableWithdrawalId, address indexed recipient, (address,uint32,uint256,uint256) withdrawal)
func (_Withdrawal *WithdrawalFilterer) FilterClaimableWithdrawalQueued(opts *bind.FilterOpts, claimableWithdrawalId []*big.Int, recipient []common.Address) (*WithdrawalClaimableWithdrawalQueuedIterator, error) {

	var claimableWithdrawalIdRule []interface{}
	for _, claimableWithdrawalIdItem := range claimableWithdrawalId {
		claimableWithdrawalIdRule = append(claimableWithdrawalIdRule, claimableWithdrawalIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _Withdrawal.contract.FilterLogs(opts, "ClaimableWithdrawalQueued", claimableWithdrawalIdRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawalClaimableWithdrawalQueuedIterator{contract: _Withdrawal.contract, event: "ClaimableWithdrawalQueued", logs: logs, sub: sub}, nil
}

// WatchClaimableWithdrawalQueued is a free log subscription operation binding the contract event 0xa82167f0d0baa7a4ee5cec53c8b897a8de2debb946fc5e79d62e3c502396b1e1.
//
// Solidity: event ClaimableWithdrawalQueued(uint256 indexed claimableWithdrawalId, address indexed recipient, (address,uint32,uint256,uint256) withdrawal)
func (_Withdrawal *WithdrawalFilterer) WatchClaimableWithdrawalQueued(opts *bind.WatchOpts, sink chan<- *WithdrawalClaimableWithdrawalQueued, claimableWithdrawalId []*big.Int, recipient []common.Address) (event.Subscription, error) {

	var claimableWithdrawalIdRule []interface{}
	for _, claimableWithdrawalIdItem := range claimableWithdrawalId {
		claimableWithdrawalIdRule = append(claimableWithdrawalIdRule, claimableWithdrawalIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _Withdrawal.contract.WatchLogs(opts, "ClaimableWithdrawalQueued", claimableWithdrawalIdRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalClaimableWithdrawalQueued)
				if err := _Withdrawal.contract.UnpackLog(event, "ClaimableWithdrawalQueued", log); err != nil {
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

// ParseClaimableWithdrawalQueued is a log parse operation binding the contract event 0xa82167f0d0baa7a4ee5cec53c8b897a8de2debb946fc5e79d62e3c502396b1e1.
//
// Solidity: event ClaimableWithdrawalQueued(uint256 indexed claimableWithdrawalId, address indexed recipient, (address,uint32,uint256,uint256) withdrawal)
func (_Withdrawal *WithdrawalFilterer) ParseClaimableWithdrawalQueued(log types.Log) (*WithdrawalClaimableWithdrawalQueued, error) {
	event := new(WithdrawalClaimableWithdrawalQueued)
	if err := _Withdrawal.contract.UnpackLog(event, "ClaimableWithdrawalQueued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDirectWithdrawalQueuedIterator is returned from FilterDirectWithdrawalQueued and is used to iterate over the raw logs and unpacked data for DirectWithdrawalQueued events raised by the Withdrawal contract.
type WithdrawalDirectWithdrawalQueuedIterator struct {
	Event *WithdrawalDirectWithdrawalQueued // Event containing the contract specifics and raw log

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
func (it *WithdrawalDirectWithdrawalQueuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDirectWithdrawalQueued)
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
		it.Event = new(WithdrawalDirectWithdrawalQueued)
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
func (it *WithdrawalDirectWithdrawalQueuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDirectWithdrawalQueuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDirectWithdrawalQueued represents a DirectWithdrawalQueued event raised by the Withdrawal contract.
type WithdrawalDirectWithdrawalQueued struct {
	DirectWithdrawalId *big.Int
	Recipient          common.Address
	Withdrawal         WithdrawalLibWithdrawal
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterDirectWithdrawalQueued is a free log retrieval operation binding the contract event 0x9623b10e54d4fc7d3add91d9d36dc3e9c58bcb11a65d6d1f6f6c40752ac7a252.
//
// Solidity: event DirectWithdrawalQueued(uint256 indexed directWithdrawalId, address indexed recipient, (address,uint32,uint256,uint256) withdrawal)
func (_Withdrawal *WithdrawalFilterer) FilterDirectWithdrawalQueued(opts *bind.FilterOpts, directWithdrawalId []*big.Int, recipient []common.Address) (*WithdrawalDirectWithdrawalQueuedIterator, error) {

	var directWithdrawalIdRule []interface{}
	for _, directWithdrawalIdItem := range directWithdrawalId {
		directWithdrawalIdRule = append(directWithdrawalIdRule, directWithdrawalIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _Withdrawal.contract.FilterLogs(opts, "DirectWithdrawalQueued", directWithdrawalIdRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDirectWithdrawalQueuedIterator{contract: _Withdrawal.contract, event: "DirectWithdrawalQueued", logs: logs, sub: sub}, nil
}

// WatchDirectWithdrawalQueued is a free log subscription operation binding the contract event 0x9623b10e54d4fc7d3add91d9d36dc3e9c58bcb11a65d6d1f6f6c40752ac7a252.
//
// Solidity: event DirectWithdrawalQueued(uint256 indexed directWithdrawalId, address indexed recipient, (address,uint32,uint256,uint256) withdrawal)
func (_Withdrawal *WithdrawalFilterer) WatchDirectWithdrawalQueued(opts *bind.WatchOpts, sink chan<- *WithdrawalDirectWithdrawalQueued, directWithdrawalId []*big.Int, recipient []common.Address) (event.Subscription, error) {

	var directWithdrawalIdRule []interface{}
	for _, directWithdrawalIdItem := range directWithdrawalId {
		directWithdrawalIdRule = append(directWithdrawalIdRule, directWithdrawalIdItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _Withdrawal.contract.WatchLogs(opts, "DirectWithdrawalQueued", directWithdrawalIdRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDirectWithdrawalQueued)
				if err := _Withdrawal.contract.UnpackLog(event, "DirectWithdrawalQueued", log); err != nil {
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

// ParseDirectWithdrawalQueued is a log parse operation binding the contract event 0x9623b10e54d4fc7d3add91d9d36dc3e9c58bcb11a65d6d1f6f6c40752ac7a252.
//
// Solidity: event DirectWithdrawalQueued(uint256 indexed directWithdrawalId, address indexed recipient, (address,uint32,uint256,uint256) withdrawal)
func (_Withdrawal *WithdrawalFilterer) ParseDirectWithdrawalQueued(log types.Log) (*WithdrawalDirectWithdrawalQueued, error) {
	event := new(WithdrawalDirectWithdrawalQueued)
	if err := _Withdrawal.contract.UnpackLog(event, "DirectWithdrawalQueued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Withdrawal contract.
type WithdrawalInitializedIterator struct {
	Event *WithdrawalInitialized // Event containing the contract specifics and raw log

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
func (it *WithdrawalInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalInitialized)
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
		it.Event = new(WithdrawalInitialized)
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
func (it *WithdrawalInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalInitialized represents a Initialized event raised by the Withdrawal contract.
type WithdrawalInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Withdrawal *WithdrawalFilterer) FilterInitialized(opts *bind.FilterOpts) (*WithdrawalInitializedIterator, error) {

	logs, sub, err := _Withdrawal.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &WithdrawalInitializedIterator{contract: _Withdrawal.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_Withdrawal *WithdrawalFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *WithdrawalInitialized) (event.Subscription, error) {

	logs, sub, err := _Withdrawal.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalInitialized)
				if err := _Withdrawal.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Withdrawal *WithdrawalFilterer) ParseInitialized(log types.Log) (*WithdrawalInitialized, error) {
	event := new(WithdrawalInitialized)
	if err := _Withdrawal.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Withdrawal contract.
type WithdrawalOwnershipTransferredIterator struct {
	Event *WithdrawalOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *WithdrawalOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalOwnershipTransferred)
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
		it.Event = new(WithdrawalOwnershipTransferred)
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
func (it *WithdrawalOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalOwnershipTransferred represents a OwnershipTransferred event raised by the Withdrawal contract.
type WithdrawalOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Withdrawal *WithdrawalFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*WithdrawalOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Withdrawal.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawalOwnershipTransferredIterator{contract: _Withdrawal.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Withdrawal *WithdrawalFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *WithdrawalOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Withdrawal.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalOwnershipTransferred)
				if err := _Withdrawal.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Withdrawal *WithdrawalFilterer) ParseOwnershipTransferred(log types.Log) (*WithdrawalOwnershipTransferred, error) {
	event := new(WithdrawalOwnershipTransferred)
	if err := _Withdrawal.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the Withdrawal contract.
type WithdrawalUpgradedIterator struct {
	Event *WithdrawalUpgraded // Event containing the contract specifics and raw log

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
func (it *WithdrawalUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalUpgraded)
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
		it.Event = new(WithdrawalUpgraded)
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
func (it *WithdrawalUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalUpgraded represents a Upgraded event raised by the Withdrawal contract.
type WithdrawalUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Withdrawal *WithdrawalFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*WithdrawalUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Withdrawal.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawalUpgradedIterator{contract: _Withdrawal.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_Withdrawal *WithdrawalFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *WithdrawalUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _Withdrawal.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalUpgraded)
				if err := _Withdrawal.contract.UnpackLog(event, "Upgraded", log); err != nil {
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
func (_Withdrawal *WithdrawalFilterer) ParseUpgraded(log types.Log) (*WithdrawalUpgraded, error) {
	event := new(WithdrawalUpgraded)
	if err := _Withdrawal.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalWithdrawalsQueuedIterator is returned from FilterWithdrawalsQueued and is used to iterate over the raw logs and unpacked data for WithdrawalsQueued events raised by the Withdrawal contract.
type WithdrawalWithdrawalsQueuedIterator struct {
	Event *WithdrawalWithdrawalsQueued // Event containing the contract specifics and raw log

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
func (it *WithdrawalWithdrawalsQueuedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalWithdrawalsQueued)
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
		it.Event = new(WithdrawalWithdrawalsQueued)
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
func (it *WithdrawalWithdrawalsQueuedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalWithdrawalsQueuedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalWithdrawalsQueued represents a WithdrawalsQueued event raised by the Withdrawal contract.
type WithdrawalWithdrawalsQueued struct {
	LastDirectWithdrawalId    *big.Int
	LastClaimableWithdrawalId *big.Int
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalsQueued is a free log retrieval operation binding the contract event 0x3be07060f55557adc400984f0e06a224d80f3a4bbcbb959ac1bbeedb0638e5ae.
//
// Solidity: event WithdrawalsQueued(uint256 lastDirectWithdrawalId, uint256 lastClaimableWithdrawalId)
func (_Withdrawal *WithdrawalFilterer) FilterWithdrawalsQueued(opts *bind.FilterOpts) (*WithdrawalWithdrawalsQueuedIterator, error) {

	logs, sub, err := _Withdrawal.contract.FilterLogs(opts, "WithdrawalsQueued")
	if err != nil {
		return nil, err
	}
	return &WithdrawalWithdrawalsQueuedIterator{contract: _Withdrawal.contract, event: "WithdrawalsQueued", logs: logs, sub: sub}, nil
}

// WatchWithdrawalsQueued is a free log subscription operation binding the contract event 0x3be07060f55557adc400984f0e06a224d80f3a4bbcbb959ac1bbeedb0638e5ae.
//
// Solidity: event WithdrawalsQueued(uint256 lastDirectWithdrawalId, uint256 lastClaimableWithdrawalId)
func (_Withdrawal *WithdrawalFilterer) WatchWithdrawalsQueued(opts *bind.WatchOpts, sink chan<- *WithdrawalWithdrawalsQueued) (event.Subscription, error) {

	logs, sub, err := _Withdrawal.contract.WatchLogs(opts, "WithdrawalsQueued")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalWithdrawalsQueued)
				if err := _Withdrawal.contract.UnpackLog(event, "WithdrawalsQueued", log); err != nil {
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

// ParseWithdrawalsQueued is a log parse operation binding the contract event 0x3be07060f55557adc400984f0e06a224d80f3a4bbcbb959ac1bbeedb0638e5ae.
//
// Solidity: event WithdrawalsQueued(uint256 lastDirectWithdrawalId, uint256 lastClaimableWithdrawalId)
func (_Withdrawal *WithdrawalFilterer) ParseWithdrawalsQueued(log types.Log) (*WithdrawalWithdrawalsQueued, error) {
	event := new(WithdrawalWithdrawalsQueued)
	if err := _Withdrawal.contract.UnpackLog(event, "WithdrawalsQueued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
