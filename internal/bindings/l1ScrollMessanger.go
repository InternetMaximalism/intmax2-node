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
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"}],\"name\":\"FailedRelayedMessage\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"messageHash\",\"type\":\"bytes32\"}],\"name\":\"RelayedMessage\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"messageNonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"SentMessage\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"maxReplayTimes\",\"type\":\"uint256\"}],\"name\":\"UpdateMaxReplayTimes\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"messageNonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"}],\"name\":\"dropMessage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"batchIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"merkleProof\",\"type\":\"bytes\"}],\"internalType\":\"structIL1ScrollMessenger.L2MessageProof\",\"name\":\"proof\",\"type\":\"tuple\"}],\"name\":\"relayMessageWithProof\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"messageNonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"newGasLimit\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"refundAddress\",\"type\":\"address\"}],\"name\":\"replayMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"refundAddress\",\"type\":\"address\"}],\"name\":\"sendMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"}],\"name\":\"sendMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"xDomainMessageSender\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
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

// DropMessage is a paid mutator transaction binding the contract method 0x29907acd.
//
// Solidity: function dropMessage(address from, address to, uint256 value, uint256 messageNonce, bytes message) returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactor) DropMessage(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int, messageNonce *big.Int, message []byte) (*types.Transaction, error) {
	return _L1ScrollMessanger.contract.Transact(opts, "dropMessage", from, to, value, messageNonce, message)
}

// DropMessage is a paid mutator transaction binding the contract method 0x29907acd.
//
// Solidity: function dropMessage(address from, address to, uint256 value, uint256 messageNonce, bytes message) returns()
func (_L1ScrollMessanger *L1ScrollMessangerSession) DropMessage(from common.Address, to common.Address, value *big.Int, messageNonce *big.Int, message []byte) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.DropMessage(&_L1ScrollMessanger.TransactOpts, from, to, value, messageNonce, message)
}

// DropMessage is a paid mutator transaction binding the contract method 0x29907acd.
//
// Solidity: function dropMessage(address from, address to, uint256 value, uint256 messageNonce, bytes message) returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactorSession) DropMessage(from common.Address, to common.Address, value *big.Int, messageNonce *big.Int, message []byte) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.DropMessage(&_L1ScrollMessanger.TransactOpts, from, to, value, messageNonce, message)
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

// ReplayMessage is a paid mutator transaction binding the contract method 0x55004105.
//
// Solidity: function replayMessage(address from, address to, uint256 value, uint256 messageNonce, bytes message, uint32 newGasLimit, address refundAddress) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactor) ReplayMessage(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int, messageNonce *big.Int, message []byte, newGasLimit uint32, refundAddress common.Address) (*types.Transaction, error) {
	return _L1ScrollMessanger.contract.Transact(opts, "replayMessage", from, to, value, messageNonce, message, newGasLimit, refundAddress)
}

// ReplayMessage is a paid mutator transaction binding the contract method 0x55004105.
//
// Solidity: function replayMessage(address from, address to, uint256 value, uint256 messageNonce, bytes message, uint32 newGasLimit, address refundAddress) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerSession) ReplayMessage(from common.Address, to common.Address, value *big.Int, messageNonce *big.Int, message []byte, newGasLimit uint32, refundAddress common.Address) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.ReplayMessage(&_L1ScrollMessanger.TransactOpts, from, to, value, messageNonce, message, newGasLimit, refundAddress)
}

// ReplayMessage is a paid mutator transaction binding the contract method 0x55004105.
//
// Solidity: function replayMessage(address from, address to, uint256 value, uint256 messageNonce, bytes message, uint32 newGasLimit, address refundAddress) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactorSession) ReplayMessage(from common.Address, to common.Address, value *big.Int, messageNonce *big.Int, message []byte, newGasLimit uint32, refundAddress common.Address) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.ReplayMessage(&_L1ScrollMessanger.TransactOpts, from, to, value, messageNonce, message, newGasLimit, refundAddress)
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

// SendMessage0 is a paid mutator transaction binding the contract method 0xb2267a7b.
//
// Solidity: function sendMessage(address target, uint256 value, bytes message, uint256 gasLimit) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactor) SendMessage0(opts *bind.TransactOpts, target common.Address, value *big.Int, message []byte, gasLimit *big.Int) (*types.Transaction, error) {
	return _L1ScrollMessanger.contract.Transact(opts, "sendMessage0", target, value, message, gasLimit)
}

// SendMessage0 is a paid mutator transaction binding the contract method 0xb2267a7b.
//
// Solidity: function sendMessage(address target, uint256 value, bytes message, uint256 gasLimit) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerSession) SendMessage0(target common.Address, value *big.Int, message []byte, gasLimit *big.Int) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.SendMessage0(&_L1ScrollMessanger.TransactOpts, target, value, message, gasLimit)
}

// SendMessage0 is a paid mutator transaction binding the contract method 0xb2267a7b.
//
// Solidity: function sendMessage(address target, uint256 value, bytes message, uint256 gasLimit) payable returns()
func (_L1ScrollMessanger *L1ScrollMessangerTransactorSession) SendMessage0(target common.Address, value *big.Int, message []byte, gasLimit *big.Int) (*types.Transaction, error) {
	return _L1ScrollMessanger.Contract.SendMessage0(&_L1ScrollMessanger.TransactOpts, target, value, message, gasLimit)
}

// L1ScrollMessangerFailedRelayedMessageIterator is returned from FilterFailedRelayedMessage and is used to iterate over the raw logs and unpacked data for FailedRelayedMessage events raised by the L1ScrollMessanger contract.
type L1ScrollMessangerFailedRelayedMessageIterator struct {
	Event *L1ScrollMessangerFailedRelayedMessage // Event containing the contract specifics and raw log

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
func (it *L1ScrollMessangerFailedRelayedMessageIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1ScrollMessangerFailedRelayedMessage)
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
		it.Event = new(L1ScrollMessangerFailedRelayedMessage)
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
func (it *L1ScrollMessangerFailedRelayedMessageIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1ScrollMessangerFailedRelayedMessageIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1ScrollMessangerFailedRelayedMessage represents a FailedRelayedMessage event raised by the L1ScrollMessanger contract.
type L1ScrollMessangerFailedRelayedMessage struct {
	MessageHash [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterFailedRelayedMessage is a free log retrieval operation binding the contract event 0x99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f.
//
// Solidity: event FailedRelayedMessage(bytes32 indexed messageHash)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) FilterFailedRelayedMessage(opts *bind.FilterOpts, messageHash [][32]byte) (*L1ScrollMessangerFailedRelayedMessageIterator, error) {

	var messageHashRule []interface{}
	for _, messageHashItem := range messageHash {
		messageHashRule = append(messageHashRule, messageHashItem)
	}

	logs, sub, err := _L1ScrollMessanger.contract.FilterLogs(opts, "FailedRelayedMessage", messageHashRule)
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessangerFailedRelayedMessageIterator{contract: _L1ScrollMessanger.contract, event: "FailedRelayedMessage", logs: logs, sub: sub}, nil
}

// WatchFailedRelayedMessage is a free log subscription operation binding the contract event 0x99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f.
//
// Solidity: event FailedRelayedMessage(bytes32 indexed messageHash)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) WatchFailedRelayedMessage(opts *bind.WatchOpts, sink chan<- *L1ScrollMessangerFailedRelayedMessage, messageHash [][32]byte) (event.Subscription, error) {

	var messageHashRule []interface{}
	for _, messageHashItem := range messageHash {
		messageHashRule = append(messageHashRule, messageHashItem)
	}

	logs, sub, err := _L1ScrollMessanger.contract.WatchLogs(opts, "FailedRelayedMessage", messageHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1ScrollMessangerFailedRelayedMessage)
				if err := _L1ScrollMessanger.contract.UnpackLog(event, "FailedRelayedMessage", log); err != nil {
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

// ParseFailedRelayedMessage is a log parse operation binding the contract event 0x99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f.
//
// Solidity: event FailedRelayedMessage(bytes32 indexed messageHash)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) ParseFailedRelayedMessage(log types.Log) (*L1ScrollMessangerFailedRelayedMessage, error) {
	event := new(L1ScrollMessangerFailedRelayedMessage)
	if err := _L1ScrollMessanger.contract.UnpackLog(event, "FailedRelayedMessage", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1ScrollMessangerRelayedMessageIterator is returned from FilterRelayedMessage and is used to iterate over the raw logs and unpacked data for RelayedMessage events raised by the L1ScrollMessanger contract.
type L1ScrollMessangerRelayedMessageIterator struct {
	Event *L1ScrollMessangerRelayedMessage // Event containing the contract specifics and raw log

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
func (it *L1ScrollMessangerRelayedMessageIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1ScrollMessangerRelayedMessage)
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
		it.Event = new(L1ScrollMessangerRelayedMessage)
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
func (it *L1ScrollMessangerRelayedMessageIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1ScrollMessangerRelayedMessageIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1ScrollMessangerRelayedMessage represents a RelayedMessage event raised by the L1ScrollMessanger contract.
type L1ScrollMessangerRelayedMessage struct {
	MessageHash [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRelayedMessage is a free log retrieval operation binding the contract event 0x4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c.
//
// Solidity: event RelayedMessage(bytes32 indexed messageHash)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) FilterRelayedMessage(opts *bind.FilterOpts, messageHash [][32]byte) (*L1ScrollMessangerRelayedMessageIterator, error) {

	var messageHashRule []interface{}
	for _, messageHashItem := range messageHash {
		messageHashRule = append(messageHashRule, messageHashItem)
	}

	logs, sub, err := _L1ScrollMessanger.contract.FilterLogs(opts, "RelayedMessage", messageHashRule)
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessangerRelayedMessageIterator{contract: _L1ScrollMessanger.contract, event: "RelayedMessage", logs: logs, sub: sub}, nil
}

// WatchRelayedMessage is a free log subscription operation binding the contract event 0x4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c.
//
// Solidity: event RelayedMessage(bytes32 indexed messageHash)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) WatchRelayedMessage(opts *bind.WatchOpts, sink chan<- *L1ScrollMessangerRelayedMessage, messageHash [][32]byte) (event.Subscription, error) {

	var messageHashRule []interface{}
	for _, messageHashItem := range messageHash {
		messageHashRule = append(messageHashRule, messageHashItem)
	}

	logs, sub, err := _L1ScrollMessanger.contract.WatchLogs(opts, "RelayedMessage", messageHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1ScrollMessangerRelayedMessage)
				if err := _L1ScrollMessanger.contract.UnpackLog(event, "RelayedMessage", log); err != nil {
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

// ParseRelayedMessage is a log parse operation binding the contract event 0x4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c.
//
// Solidity: event RelayedMessage(bytes32 indexed messageHash)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) ParseRelayedMessage(log types.Log) (*L1ScrollMessangerRelayedMessage, error) {
	event := new(L1ScrollMessangerRelayedMessage)
	if err := _L1ScrollMessanger.contract.UnpackLog(event, "RelayedMessage", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1ScrollMessangerSentMessageIterator is returned from FilterSentMessage and is used to iterate over the raw logs and unpacked data for SentMessage events raised by the L1ScrollMessanger contract.
type L1ScrollMessangerSentMessageIterator struct {
	Event *L1ScrollMessangerSentMessage // Event containing the contract specifics and raw log

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
func (it *L1ScrollMessangerSentMessageIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1ScrollMessangerSentMessage)
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
		it.Event = new(L1ScrollMessangerSentMessage)
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
func (it *L1ScrollMessangerSentMessageIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1ScrollMessangerSentMessageIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1ScrollMessangerSentMessage represents a SentMessage event raised by the L1ScrollMessanger contract.
type L1ScrollMessangerSentMessage struct {
	Sender       common.Address
	Target       common.Address
	Value        *big.Int
	MessageNonce *big.Int
	GasLimit     *big.Int
	Message      []byte
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSentMessage is a free log retrieval operation binding the contract event 0x104371f3b442861a2a7b82a070afbbaab748bb13757bf47769e170e37809ec1e.
//
// Solidity: event SentMessage(address indexed sender, address indexed target, uint256 value, uint256 messageNonce, uint256 gasLimit, bytes message)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) FilterSentMessage(opts *bind.FilterOpts, sender []common.Address, target []common.Address) (*L1ScrollMessangerSentMessageIterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _L1ScrollMessanger.contract.FilterLogs(opts, "SentMessage", senderRule, targetRule)
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessangerSentMessageIterator{contract: _L1ScrollMessanger.contract, event: "SentMessage", logs: logs, sub: sub}, nil
}

// WatchSentMessage is a free log subscription operation binding the contract event 0x104371f3b442861a2a7b82a070afbbaab748bb13757bf47769e170e37809ec1e.
//
// Solidity: event SentMessage(address indexed sender, address indexed target, uint256 value, uint256 messageNonce, uint256 gasLimit, bytes message)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) WatchSentMessage(opts *bind.WatchOpts, sink chan<- *L1ScrollMessangerSentMessage, sender []common.Address, target []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _L1ScrollMessanger.contract.WatchLogs(opts, "SentMessage", senderRule, targetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1ScrollMessangerSentMessage)
				if err := _L1ScrollMessanger.contract.UnpackLog(event, "SentMessage", log); err != nil {
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

// ParseSentMessage is a log parse operation binding the contract event 0x104371f3b442861a2a7b82a070afbbaab748bb13757bf47769e170e37809ec1e.
//
// Solidity: event SentMessage(address indexed sender, address indexed target, uint256 value, uint256 messageNonce, uint256 gasLimit, bytes message)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) ParseSentMessage(log types.Log) (*L1ScrollMessangerSentMessage, error) {
	event := new(L1ScrollMessangerSentMessage)
	if err := _L1ScrollMessanger.contract.UnpackLog(event, "SentMessage", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L1ScrollMessangerUpdateMaxReplayTimesIterator is returned from FilterUpdateMaxReplayTimes and is used to iterate over the raw logs and unpacked data for UpdateMaxReplayTimes events raised by the L1ScrollMessanger contract.
type L1ScrollMessangerUpdateMaxReplayTimesIterator struct {
	Event *L1ScrollMessangerUpdateMaxReplayTimes // Event containing the contract specifics and raw log

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
func (it *L1ScrollMessangerUpdateMaxReplayTimesIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L1ScrollMessangerUpdateMaxReplayTimes)
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
		it.Event = new(L1ScrollMessangerUpdateMaxReplayTimes)
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
func (it *L1ScrollMessangerUpdateMaxReplayTimesIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L1ScrollMessangerUpdateMaxReplayTimesIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L1ScrollMessangerUpdateMaxReplayTimes represents a UpdateMaxReplayTimes event raised by the L1ScrollMessanger contract.
type L1ScrollMessangerUpdateMaxReplayTimes struct {
	MaxReplayTimes *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpdateMaxReplayTimes is a free log retrieval operation binding the contract event 0x87447df30e533f82a48d0b72640868bb678dbec0d35314564ad5590a0191d90a.
//
// Solidity: event UpdateMaxReplayTimes(uint256 maxReplayTimes)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) FilterUpdateMaxReplayTimes(opts *bind.FilterOpts) (*L1ScrollMessangerUpdateMaxReplayTimesIterator, error) {

	logs, sub, err := _L1ScrollMessanger.contract.FilterLogs(opts, "UpdateMaxReplayTimes")
	if err != nil {
		return nil, err
	}
	return &L1ScrollMessangerUpdateMaxReplayTimesIterator{contract: _L1ScrollMessanger.contract, event: "UpdateMaxReplayTimes", logs: logs, sub: sub}, nil
}

// WatchUpdateMaxReplayTimes is a free log subscription operation binding the contract event 0x87447df30e533f82a48d0b72640868bb678dbec0d35314564ad5590a0191d90a.
//
// Solidity: event UpdateMaxReplayTimes(uint256 maxReplayTimes)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) WatchUpdateMaxReplayTimes(opts *bind.WatchOpts, sink chan<- *L1ScrollMessangerUpdateMaxReplayTimes) (event.Subscription, error) {

	logs, sub, err := _L1ScrollMessanger.contract.WatchLogs(opts, "UpdateMaxReplayTimes")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L1ScrollMessangerUpdateMaxReplayTimes)
				if err := _L1ScrollMessanger.contract.UnpackLog(event, "UpdateMaxReplayTimes", log); err != nil {
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

// ParseUpdateMaxReplayTimes is a log parse operation binding the contract event 0x87447df30e533f82a48d0b72640868bb678dbec0d35314564ad5590a0191d90a.
//
// Solidity: event UpdateMaxReplayTimes(uint256 maxReplayTimes)
func (_L1ScrollMessanger *L1ScrollMessangerFilterer) ParseUpdateMaxReplayTimes(log types.Log) (*L1ScrollMessangerUpdateMaxReplayTimes, error) {
	event := new(L1ScrollMessangerUpdateMaxReplayTimes)
	if err := _L1ScrollMessanger.contract.UnpackLog(event, "UpdateMaxReplayTimes", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
