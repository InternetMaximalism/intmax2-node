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

// IBlockBuilderRegistryBlockBuilderInfo is an auto generated low-level Go binding around an user-defined struct.
type IBlockBuilderRegistryBlockBuilderInfo struct {
	BlockBuilderUrl string
	StakeAmount     *big.Int
	StopTime        *big.Int
	NumSlashes      *big.Int
	IsValid         bool
}

// BlockBuilderRegistryMetaData contains all meta data concerning the BlockBuilderRegistry contract.
var BlockBuilderRegistryMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"}],\"name\":\"BlockBuilderStoped\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"url\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"}],\"name\":\"BlockBuilderUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"CHALLENGE_DURATION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_STAKE_AMOUNT\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"}],\"name\":\"getBlockBuilder\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"blockBuilderUrl\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stakeAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stopTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"numSlashes\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isValid\",\"type\":\"bool\"}],\"internalType\":\"structIBlockBuilderRegistry.BlockBuilderInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"}],\"name\":\"isValidBlockBuilder\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"blockBuilder\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"publicInputs\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"slashBlockBuilder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stopBlockBuilder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"url\",\"type\":\"string\"}],\"name\":\"updateBlockBuilder\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"}]",
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

// CHALLENGEDURATION is a free data retrieval call binding the contract method 0x67ab845b.
//
// Solidity: function CHALLENGE_DURATION() view returns(uint256)
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) CHALLENGEDURATION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "CHALLENGE_DURATION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CHALLENGEDURATION is a free data retrieval call binding the contract method 0x67ab845b.
//
// Solidity: function CHALLENGE_DURATION() view returns(uint256)
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) CHALLENGEDURATION() (*big.Int, error) {
	return _BlockBuilderRegistry.Contract.CHALLENGEDURATION(&_BlockBuilderRegistry.CallOpts)
}

// CHALLENGEDURATION is a free data retrieval call binding the contract method 0x67ab845b.
//
// Solidity: function CHALLENGE_DURATION() view returns(uint256)
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) CHALLENGEDURATION() (*big.Int, error) {
	return _BlockBuilderRegistry.Contract.CHALLENGEDURATION(&_BlockBuilderRegistry.CallOpts)
}

// MINSTAKEAMOUNT is a free data retrieval call binding the contract method 0x27ed7188.
//
// Solidity: function MIN_STAKE_AMOUNT() view returns(uint256)
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) MINSTAKEAMOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "MIN_STAKE_AMOUNT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINSTAKEAMOUNT is a free data retrieval call binding the contract method 0x27ed7188.
//
// Solidity: function MIN_STAKE_AMOUNT() view returns(uint256)
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) MINSTAKEAMOUNT() (*big.Int, error) {
	return _BlockBuilderRegistry.Contract.MINSTAKEAMOUNT(&_BlockBuilderRegistry.CallOpts)
}

// MINSTAKEAMOUNT is a free data retrieval call binding the contract method 0x27ed7188.
//
// Solidity: function MIN_STAKE_AMOUNT() view returns(uint256)
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) MINSTAKEAMOUNT() (*big.Int, error) {
	return _BlockBuilderRegistry.Contract.MINSTAKEAMOUNT(&_BlockBuilderRegistry.CallOpts)
}

// GetBlockBuilder is a free data retrieval call binding the contract method 0x33587abf.
//
// Solidity: function getBlockBuilder(address blockBuilder) view returns((string,uint256,uint256,uint256,bool))
func (_BlockBuilderRegistry *BlockBuilderRegistryCaller) GetBlockBuilder(opts *bind.CallOpts, blockBuilder common.Address) (IBlockBuilderRegistryBlockBuilderInfo, error) {
	var out []interface{}
	err := _BlockBuilderRegistry.contract.Call(opts, &out, "getBlockBuilder", blockBuilder)

	if err != nil {
		return *new(IBlockBuilderRegistryBlockBuilderInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(IBlockBuilderRegistryBlockBuilderInfo)).(*IBlockBuilderRegistryBlockBuilderInfo)

	return out0, err

}

// GetBlockBuilder is a free data retrieval call binding the contract method 0x33587abf.
//
// Solidity: function getBlockBuilder(address blockBuilder) view returns((string,uint256,uint256,uint256,bool))
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) GetBlockBuilder(blockBuilder common.Address) (IBlockBuilderRegistryBlockBuilderInfo, error) {
	return _BlockBuilderRegistry.Contract.GetBlockBuilder(&_BlockBuilderRegistry.CallOpts, blockBuilder)
}

// GetBlockBuilder is a free data retrieval call binding the contract method 0x33587abf.
//
// Solidity: function getBlockBuilder(address blockBuilder) view returns((string,uint256,uint256,uint256,bool))
func (_BlockBuilderRegistry *BlockBuilderRegistryCallerSession) GetBlockBuilder(blockBuilder common.Address) (IBlockBuilderRegistryBlockBuilderInfo, error) {
	return _BlockBuilderRegistry.Contract.GetBlockBuilder(&_BlockBuilderRegistry.CallOpts, blockBuilder)
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

// SlashBlockBuilder is a paid mutator transaction binding the contract method 0xd4d80a07.
//
// Solidity: function slashBlockBuilder(uint32 blockNumber, address blockBuilder, uint256[] publicInputs, bytes proof) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactor) SlashBlockBuilder(opts *bind.TransactOpts, blockNumber uint32, blockBuilder common.Address, publicInputs []*big.Int, proof []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.contract.Transact(opts, "slashBlockBuilder", blockNumber, blockBuilder, publicInputs, proof)
}

// SlashBlockBuilder is a paid mutator transaction binding the contract method 0xd4d80a07.
//
// Solidity: function slashBlockBuilder(uint32 blockNumber, address blockBuilder, uint256[] publicInputs, bytes proof) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistrySession) SlashBlockBuilder(blockNumber uint32, blockBuilder common.Address, publicInputs []*big.Int, proof []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.SlashBlockBuilder(&_BlockBuilderRegistry.TransactOpts, blockNumber, blockBuilder, publicInputs, proof)
}

// SlashBlockBuilder is a paid mutator transaction binding the contract method 0xd4d80a07.
//
// Solidity: function slashBlockBuilder(uint32 blockNumber, address blockBuilder, uint256[] publicInputs, bytes proof) returns()
func (_BlockBuilderRegistry *BlockBuilderRegistryTransactorSession) SlashBlockBuilder(blockNumber uint32, blockBuilder common.Address, publicInputs []*big.Int, proof []byte) (*types.Transaction, error) {
	return _BlockBuilderRegistry.Contract.SlashBlockBuilder(&_BlockBuilderRegistry.TransactOpts, blockNumber, blockBuilder, publicInputs, proof)
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

// BlockBuilderRegistryBlockBuilderStopedIterator is returned from FilterBlockBuilderStoped and is used to iterate over the raw logs and unpacked data for BlockBuilderStoped events raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderStopedIterator struct {
	Event *BlockBuilderRegistryBlockBuilderStoped // Event containing the contract specifics and raw log

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
func (it *BlockBuilderRegistryBlockBuilderStopedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(BlockBuilderRegistryBlockBuilderStoped)
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
		it.Event = new(BlockBuilderRegistryBlockBuilderStoped)
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
func (it *BlockBuilderRegistryBlockBuilderStopedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *BlockBuilderRegistryBlockBuilderStopedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// BlockBuilderRegistryBlockBuilderStoped represents a BlockBuilderStoped event raised by the BlockBuilderRegistry contract.
type BlockBuilderRegistryBlockBuilderStoped struct {
	BlockBuilder common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBlockBuilderStoped is a free log retrieval operation binding the contract event 0x93e5347bfe24301285595924ebe1940d4d0bce1c95b212d44fcf6ec4c0f2c145.
//
// Solidity: event BlockBuilderStoped(address indexed blockBuilder)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) FilterBlockBuilderStoped(opts *bind.FilterOpts, blockBuilder []common.Address) (*BlockBuilderRegistryBlockBuilderStopedIterator, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.FilterLogs(opts, "BlockBuilderStoped", blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return &BlockBuilderRegistryBlockBuilderStopedIterator{contract: _BlockBuilderRegistry.contract, event: "BlockBuilderStoped", logs: logs, sub: sub}, nil
}

// WatchBlockBuilderStoped is a free log subscription operation binding the contract event 0x93e5347bfe24301285595924ebe1940d4d0bce1c95b212d44fcf6ec4c0f2c145.
//
// Solidity: event BlockBuilderStoped(address indexed blockBuilder)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) WatchBlockBuilderStoped(opts *bind.WatchOpts, sink chan<- *BlockBuilderRegistryBlockBuilderStoped, blockBuilder []common.Address) (event.Subscription, error) {

	var blockBuilderRule []interface{}
	for _, blockBuilderItem := range blockBuilder {
		blockBuilderRule = append(blockBuilderRule, blockBuilderItem)
	}

	logs, sub, err := _BlockBuilderRegistry.contract.WatchLogs(opts, "BlockBuilderStoped", blockBuilderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(BlockBuilderRegistryBlockBuilderStoped)
				if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderStoped", log); err != nil {
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

// ParseBlockBuilderStoped is a log parse operation binding the contract event 0x93e5347bfe24301285595924ebe1940d4d0bce1c95b212d44fcf6ec4c0f2c145.
//
// Solidity: event BlockBuilderStoped(address indexed blockBuilder)
func (_BlockBuilderRegistry *BlockBuilderRegistryFilterer) ParseBlockBuilderStoped(log types.Log) (*BlockBuilderRegistryBlockBuilderStoped, error) {
	event := new(BlockBuilderRegistryBlockBuilderStoped)
	if err := _BlockBuilderRegistry.contract.UnpackLog(event, "BlockBuilderStoped", log); err != nil {
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
