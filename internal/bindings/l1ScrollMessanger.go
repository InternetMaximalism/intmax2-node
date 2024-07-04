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

// IL1ScrollMessengerL2MessageProof is an auto generated low-level Go binding around an user-defined struct.
type IL1ScrollMessengerL2MessageProof struct {
	BatchIndex  *big.Int
	MerkleProof []byte
}

// L1ScrollMessangerMetaData contains all meta data concerning the L1ScrollMessanger contract.
var L1ScrollMessangerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"batchIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"merkleProof\",\"type\":\"bytes\"}],\"internalType\":\"structIL1ScrollMessenger.L2MessageProof\",\"name\":\"proof\",\"type\":\"tuple\"}],\"name\":\"relayMessageWithProof\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"refundAddress\",\"type\":\"address\"}],\"name\":\"sendMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"xDomainMessageSender\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// L1ScrollMessangerABI is the input ABI used to generate the binding from.
// Deprecated: Use L1ScrollMessangerMetaData.ABI instead.
var L1ScrollMessangerABI = L1ScrollMessangerMetaData.ABI

// L1ScrollMessanger is an auto generated Go binding around an Ethereum contract.
type L1ScrollMessanger struct {
	L1ScrollMessangerCaller     // Read-only binding to the contract
	L1ScrollMessangerTransactor // Write-only binding to the contract
	L1ScrollMessangerFilterer   // Log filterer for contract events
}

// L1ScrollMessangerCaller is an auto generated read-only Go binding around an Ethereum contract.
type L1ScrollMessangerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L1ScrollMessangerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type L1ScrollMessangerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L1ScrollMessangerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type L1ScrollMessangerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L1ScrollMessangerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type L1ScrollMessangerSession struct {
	Contract     *L1ScrollMessanger // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// L1ScrollMessangerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type L1ScrollMessangerCallerSession struct {
	Contract *L1ScrollMessangerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// L1ScrollMessangerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type L1ScrollMessangerTransactorSession struct {
	Contract     *L1ScrollMessangerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// L1ScrollMessangerRaw is an auto generated low-level Go binding around an Ethereum contract.
type L1ScrollMessangerRaw struct {
	Contract *L1ScrollMessanger // Generic contract binding to access the raw methods on
}

// L1ScrollMessangerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type L1ScrollMessangerCallerRaw struct {
	Contract *L1ScrollMessangerCaller // Generic read-only contract binding to access the raw methods on
}

// L1ScrollMessangerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type L1ScrollMessangerTransactorRaw struct {
	Contract *L1ScrollMessangerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewL1ScrollMessanger creates a new instance of L1ScrollMessanger, bound to a specific deployed contract.
func NewL1ScrollMessanger(address common.Address, backend bind.ContractBackend) (*L1ScrollMessanger, error) {
	contract, err := bindL1ScrollMessanger(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessanger{L1ScrollMessangerCaller: L1ScrollMessangerCaller{contract: contract}, L1ScrollMessangerTransactor: L1ScrollMessangerTransactor{contract: contract}, L1ScrollMessangerFilterer: L1ScrollMessangerFilterer{contract: contract}}, nil
}

// NewL1ScrollMessangerCaller creates a new read-only instance of L1ScrollMessanger, bound to a specific deployed contract.
func NewL1ScrollMessangerCaller(address common.Address, caller bind.ContractCaller) (*L1ScrollMessangerCaller, error) {
	contract, err := bindL1ScrollMessanger(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessangerCaller{contract: contract}, nil
}

// NewL1ScrollMessangerTransactor creates a new write-only instance of L1ScrollMessanger, bound to a specific deployed contract.
func NewL1ScrollMessangerTransactor(address common.Address, transactor bind.ContractTransactor) (*L1ScrollMessangerTransactor, error) {
	contract, err := bindL1ScrollMessanger(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessangerTransactor{contract: contract}, nil
}

// NewL1ScrollMessangerFilterer creates a new log filterer instance of L1ScrollMessanger, bound to a specific deployed contract.
func NewL1ScrollMessangerFilterer(address common.Address, filterer bind.ContractFilterer) (*L1ScrollMessangerFilterer, error) {
	contract, err := bindL1ScrollMessanger(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessangerFilterer{contract: contract}, nil
}

// bindL1ScrollMessanger binds a generic wrapper to an already deployed contract.
func bindL1ScrollMessanger(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := L1ScrollMessangerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L1ScrollMessanger *L1ScrollMessangerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L1ScrollMessanger.Contract.L1ScrollMessangerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L1ScrollMessanger *L1ScrollMessangerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.L1ScrollMessangerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L1ScrollMessanger *L1ScrollMessangerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.L1ScrollMessangerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L1ScrollMessanger *L1ScrollMessangerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L1ScrollMessanger.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L1ScrollMessanger *L1ScrollMessangerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L1ScrollMessanger *L1ScrollMessangerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.contract.Transact(opts, method, params...)
}

// XDomainMessageSender is a free data retrieval call binding the contract method 0x6e296e45.
//
// Solidity: function xDomainMessageSender() view returns(address)
func (_L1ScrollMessanger *L1ScrollMessangerCaller) XDomainMessageSender(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L1ScrollMessanger.contract.Call(opts, &out, "xDomainMessageSender")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// XDomainMessageSender is a free data retrieval call binding the contract method 0x6e296e45.
//
// Solidity: function xDomainMessageSender() view returns(address)
func (_L1ScrollMessanger *L1ScrollMessangerSession) XDomainMessageSender() (common.Address, error) {
	return _L1ScrollMessanger.Contract.XDomainMessageSender(&_L1ScrollMessanger.CallOpts)
}

// XDomainMessageSender is a free data retrieval call binding the contract method 0x6e296e45.
//
// Solidity: function xDomainMessageSender() view returns(address)
func (_L1ScrollMessanger *L1ScrollMessangerCallerSession) XDomainMessageSender() (common.Address, error) {
	return _L1ScrollMessanger.Contract.XDomainMessageSender(&_L1ScrollMessanger.CallOpts)
}

// RelayMessageWithProof is a paid mutator transaction binding the contract method 0xc311b6fc.
//
// Solidity: function relayMessageWithProof(address from, address to, uint256 value, uint256 nonce, bytes message, (uint256,bytes) proof) returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactor) RelayMessageWithProof(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int, nonce *big.Int, message []byte, proof IL1ScrollMessengerL2MessageProof) (*types.Transaction, error) {
	return _L1ScrollMessanger.contract.Transact(opts, "relayMessageWithProof", from, to, value, nonce, message, proof)
}

// RelayMessageWithProof is a paid mutator transaction binding the contract method 0xc311b6fc.
//
// Solidity: function relayMessageWithProof(address from, address to, uint256 value, uint256 nonce, bytes message, (uint256,bytes) proof) returns()
func (_L1ScrollMessanger *L1ScrollMessangerSession) RelayMessageWithProof(from common.Address, to common.Address, value *big.Int, nonce *big.Int, message []byte, proof IL1ScrollMessengerL2MessageProof) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.RelayMessageWithProof(&_L1ScrollMessanger.TransactOpts, from, to, value, nonce, message, proof)
}

// RelayMessageWithProof is a paid mutator transaction binding the contract method 0xc311b6fc.
//
// Solidity: function relayMessageWithProof(address from, address to, uint256 value, uint256 nonce, bytes message, (uint256,bytes) proof) returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactorSession) RelayMessageWithProof(from common.Address, to common.Address, value *big.Int, nonce *big.Int, message []byte, proof IL1ScrollMessengerL2MessageProof) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.RelayMessageWithProof(&_L1ScrollMessanger.TransactOpts, from, to, value, nonce, message, proof)
}

// SendMessage is a paid mutator transaction binding the contract method 0x5f7b1577.
//
// Solidity: function sendMessage(address target, uint256 value, bytes message, uint256 gasLimit, address refundAddress) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactor) SendMessage(opts *bind.TransactOpts, target common.Address, value *big.Int, message []byte, gasLimit *big.Int, refundAddress common.Address) (*types.Transaction, error) {
	return _L1ScrollMessanger.contract.Transact(opts, "sendMessage", target, value, message, gasLimit, refundAddress)
}

// SendMessage is a paid mutator transaction binding the contract method 0x5f7b1577.
//
// Solidity: function sendMessage(address target, uint256 value, bytes message, uint256 gasLimit, address refundAddress) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerSession) SendMessage(target common.Address, value *big.Int, message []byte, gasLimit *big.Int, refundAddress common.Address) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.SendMessage(&_L1ScrollMessanger.TransactOpts, target, value, message, gasLimit, refundAddress)
}

// SendMessage is a paid mutator transaction binding the contract method 0x5f7b1577.
//
// Solidity: function sendMessage(address target, uint256 value, bytes message, uint256 gasLimit, address refundAddress) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactorSession) SendMessage(target common.Address, value *big.Int, message []byte, gasLimit *big.Int, refundAddress common.Address) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.SendMessage(&_L1ScrollMessanger.TransactOpts, target, value, message, gasLimit, refundAddress)
}
