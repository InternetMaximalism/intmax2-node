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

// GasPriceOracleScrollL1ETHMetaData contains all meta data concerning the GasPriceOracleScrollL1ETH contract.
var GasPriceOracleScrollL1ETHMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"blobScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"commitScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"enableCurie\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getL1Fee\",\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getL1GasUsed\",\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isCurie\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"l1BaseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"l1BlobBaseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"overhead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"scalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setBlobScalar\",\"inputs\":[{\"name\":\"_scalar\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setCommitScalar\",\"inputs\":[{\"name\":\"_scalar\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setL1BaseFee\",\"inputs\":[{\"name\":\"_l1BaseFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setL1BaseFeeAndBlobBaseFee\",\"inputs\":[{\"name\":\"_l1BaseFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_l1BlobBaseFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOverhead\",\"inputs\":[{\"name\":\"_overhead\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setScalar\",\"inputs\":[{\"name\":\"_scalar\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"_newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"updateWhitelist\",\"inputs\":[{\"name\":\"_newWhitelist\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"whitelist\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIWhitelist\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"BlobScalarUpdated\",\"inputs\":[{\"name\":\"scalar\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"CommitScalarUpdated\",\"inputs\":[{\"name\":\"scalar\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"L1BaseFeeUpdated\",\"inputs\":[{\"name\":\"l1BaseFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"L1BlobBaseFeeUpdated\",\"inputs\":[{\"name\":\"l1BlobBaseFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OverheadUpdated\",\"inputs\":[{\"name\":\"overhead\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"_oldOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"_newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ScalarUpdated\",\"inputs\":[{\"name\":\"scalar\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UpdateWhitelist\",\"inputs\":[{\"name\":\"_oldWhitelist\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"_newWhitelist\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"ErrAlreadyInCurieFork\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ErrCallerNotWhitelisted\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ErrExceedMaxBlobScalar\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ErrExceedMaxCommitScalar\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ErrExceedMaxOverhead\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ErrExceedMaxScalar\",\"inputs\":[]}]",
}

// GasPriceOracleScrollL1ETHABI is the input ABI used to generate the binding from.
// Deprecated: Use GasPriceOracleScrollL1ETHMetaData.ABI instead.
var GasPriceOracleScrollL1ETHABI = GasPriceOracleScrollL1ETHMetaData.ABI

// GasPriceOracleScrollL1ETH is an auto generated Go binding around an Ethereum contract.
type GasPriceOracleScrollL1ETH struct {
	GasPriceOracleScrollL1ETHCaller     // Read-only binding to the contract
	GasPriceOracleScrollL1ETHTransactor // Write-only binding to the contract
	GasPriceOracleScrollL1ETHFilterer   // Log filterer for contract events
}

// GasPriceOracleScrollL1ETHCaller is an auto generated read-only Go binding around an Ethereum contract.
type GasPriceOracleScrollL1ETHCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GasPriceOracleScrollL1ETHTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GasPriceOracleScrollL1ETHTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GasPriceOracleScrollL1ETHFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GasPriceOracleScrollL1ETHFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GasPriceOracleScrollL1ETHSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GasPriceOracleScrollL1ETHSession struct {
	Contract     *GasPriceOracleScrollL1ETH // Generic contract binding to set the session for
	CallOpts     bind.CallOpts              // Call options to use throughout this session
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// GasPriceOracleScrollL1ETHCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GasPriceOracleScrollL1ETHCallerSession struct {
	Contract *GasPriceOracleScrollL1ETHCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                    // Call options to use throughout this session
}

// GasPriceOracleScrollL1ETHTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GasPriceOracleScrollL1ETHTransactorSession struct {
	Contract     *GasPriceOracleScrollL1ETHTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                    // Transaction auth options to use throughout this session
}

// GasPriceOracleScrollL1ETHRaw is an auto generated low-level Go binding around an Ethereum contract.
type GasPriceOracleScrollL1ETHRaw struct {
	Contract *GasPriceOracleScrollL1ETH // Generic contract binding to access the raw methods on
}

// GasPriceOracleScrollL1ETHCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GasPriceOracleScrollL1ETHCallerRaw struct {
	Contract *GasPriceOracleScrollL1ETHCaller // Generic read-only contract binding to access the raw methods on
}

// GasPriceOracleScrollL1ETHTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GasPriceOracleScrollL1ETHTransactorRaw struct {
	Contract *GasPriceOracleScrollL1ETHTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGasPriceOracleScrollL1ETH creates a new instance of GasPriceOracleScrollL1ETH, bound to a specific deployed contract.
func NewGasPriceOracleScrollL1ETH(address common.Address, backend bind.ContractBackend) (*GasPriceOracleScrollL1ETH, error) {
	contract, err := bindGasPriceOracleScrollL1ETH(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETH{GasPriceOracleScrollL1ETHCaller: GasPriceOracleScrollL1ETHCaller{contract: contract}, GasPriceOracleScrollL1ETHTransactor: GasPriceOracleScrollL1ETHTransactor{contract: contract}, GasPriceOracleScrollL1ETHFilterer: GasPriceOracleScrollL1ETHFilterer{contract: contract}}, nil
}

// NewGasPriceOracleScrollL1ETHCaller creates a new read-only instance of GasPriceOracleScrollL1ETH, bound to a specific deployed contract.
func NewGasPriceOracleScrollL1ETHCaller(address common.Address, caller bind.ContractCaller) (*GasPriceOracleScrollL1ETHCaller, error) {
	contract, err := bindGasPriceOracleScrollL1ETH(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHCaller{contract: contract}, nil
}

// NewGasPriceOracleScrollL1ETHTransactor creates a new write-only instance of GasPriceOracleScrollL1ETH, bound to a specific deployed contract.
func NewGasPriceOracleScrollL1ETHTransactor(address common.Address, transactor bind.ContractTransactor) (*GasPriceOracleScrollL1ETHTransactor, error) {
	contract, err := bindGasPriceOracleScrollL1ETH(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHTransactor{contract: contract}, nil
}

// NewGasPriceOracleScrollL1ETHFilterer creates a new log filterer instance of GasPriceOracleScrollL1ETH, bound to a specific deployed contract.
func NewGasPriceOracleScrollL1ETHFilterer(address common.Address, filterer bind.ContractFilterer) (*GasPriceOracleScrollL1ETHFilterer, error) {
	contract, err := bindGasPriceOracleScrollL1ETH(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHFilterer{contract: contract}, nil
}

// bindGasPriceOracleScrollL1ETH binds a generic wrapper to an already deployed contract.
func bindGasPriceOracleScrollL1ETH(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GasPriceOracleScrollL1ETHMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GasPriceOracleScrollL1ETH.Contract.GasPriceOracleScrollL1ETHCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.GasPriceOracleScrollL1ETHTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.GasPriceOracleScrollL1ETHTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GasPriceOracleScrollL1ETH.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.contract.Transact(opts, method, params...)
}

// BlobScalar is a free data retrieval call binding the contract method 0x6a5e67e5.
//
// Solidity: function blobScalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) BlobScalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "blobScalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BlobScalar is a free data retrieval call binding the contract method 0x6a5e67e5.
//
// Solidity: function blobScalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) BlobScalar() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.BlobScalar(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// BlobScalar is a free data retrieval call binding the contract method 0x6a5e67e5.
//
// Solidity: function blobScalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) BlobScalar() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.BlobScalar(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// CommitScalar is a free data retrieval call binding the contract method 0x23e524ac.
//
// Solidity: function commitScalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) CommitScalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "commitScalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CommitScalar is a free data retrieval call binding the contract method 0x23e524ac.
//
// Solidity: function commitScalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) CommitScalar() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.CommitScalar(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// CommitScalar is a free data retrieval call binding the contract method 0x23e524ac.
//
// Solidity: function commitScalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) CommitScalar() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.CommitScalar(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) GetL1Fee(opts *bind.CallOpts, _data []byte) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "getL1Fee", _data)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) GetL1Fee(_data []byte) (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.GetL1Fee(&_GasPriceOracleScrollL1ETH.CallOpts, _data)
}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) GetL1Fee(_data []byte) (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.GetL1Fee(&_GasPriceOracleScrollL1ETH.CallOpts, _data)
}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) GetL1GasUsed(opts *bind.CallOpts, _data []byte) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "getL1GasUsed", _data)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) GetL1GasUsed(_data []byte) (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.GetL1GasUsed(&_GasPriceOracleScrollL1ETH.CallOpts, _data)
}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) GetL1GasUsed(_data []byte) (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.GetL1GasUsed(&_GasPriceOracleScrollL1ETH.CallOpts, _data)
}

// IsCurie is a free data retrieval call binding the contract method 0x13dad5be.
//
// Solidity: function isCurie() view returns(bool)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) IsCurie(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "isCurie")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsCurie is a free data retrieval call binding the contract method 0x13dad5be.
//
// Solidity: function isCurie() view returns(bool)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) IsCurie() (bool, error) {
	return _GasPriceOracleScrollL1ETH.Contract.IsCurie(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// IsCurie is a free data retrieval call binding the contract method 0x13dad5be.
//
// Solidity: function isCurie() view returns(bool)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) IsCurie() (bool, error) {
	return _GasPriceOracleScrollL1ETH.Contract.IsCurie(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) L1BaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "l1BaseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) L1BaseFee() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.L1BaseFee(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) L1BaseFee() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.L1BaseFee(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// L1BlobBaseFee is a free data retrieval call binding the contract method 0x84189161.
//
// Solidity: function l1BlobBaseFee() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) L1BlobBaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "l1BlobBaseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// L1BlobBaseFee is a free data retrieval call binding the contract method 0x84189161.
//
// Solidity: function l1BlobBaseFee() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) L1BlobBaseFee() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.L1BlobBaseFee(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// L1BlobBaseFee is a free data retrieval call binding the contract method 0x84189161.
//
// Solidity: function l1BlobBaseFee() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) L1BlobBaseFee() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.L1BlobBaseFee(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) Overhead(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "overhead")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) Overhead() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Overhead(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) Overhead() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Overhead(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) Owner() (common.Address, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Owner(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) Owner() (common.Address, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Owner(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) Scalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "scalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) Scalar() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Scalar(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) Scalar() (*big.Int, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Scalar(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Whitelist is a free data retrieval call binding the contract method 0x93e59dc1.
//
// Solidity: function whitelist() view returns(address)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCaller) Whitelist(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GasPriceOracleScrollL1ETH.contract.Call(opts, &out, "whitelist")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Whitelist is a free data retrieval call binding the contract method 0x93e59dc1.
//
// Solidity: function whitelist() view returns(address)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) Whitelist() (common.Address, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Whitelist(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// Whitelist is a free data retrieval call binding the contract method 0x93e59dc1.
//
// Solidity: function whitelist() view returns(address)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHCallerSession) Whitelist() (common.Address, error) {
	return _GasPriceOracleScrollL1ETH.Contract.Whitelist(&_GasPriceOracleScrollL1ETH.CallOpts)
}

// EnableCurie is a paid mutator transaction binding the contract method 0xa911d77f.
//
// Solidity: function enableCurie() returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) EnableCurie(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "enableCurie")
}

// EnableCurie is a paid mutator transaction binding the contract method 0xa911d77f.
//
// Solidity: function enableCurie() returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) EnableCurie() (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.EnableCurie(&_GasPriceOracleScrollL1ETH.TransactOpts)
}

// EnableCurie is a paid mutator transaction binding the contract method 0xa911d77f.
//
// Solidity: function enableCurie() returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) EnableCurie() (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.EnableCurie(&_GasPriceOracleScrollL1ETH.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) RenounceOwnership() (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.RenounceOwnership(&_GasPriceOracleScrollL1ETH.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.RenounceOwnership(&_GasPriceOracleScrollL1ETH.TransactOpts)
}

// SetBlobScalar is a paid mutator transaction binding the contract method 0xe88a60ad.
//
// Solidity: function setBlobScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) SetBlobScalar(opts *bind.TransactOpts, _scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "setBlobScalar", _scalar)
}

// SetBlobScalar is a paid mutator transaction binding the contract method 0xe88a60ad.
//
// Solidity: function setBlobScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) SetBlobScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetBlobScalar(&_GasPriceOracleScrollL1ETH.TransactOpts, _scalar)
}

// SetBlobScalar is a paid mutator transaction binding the contract method 0xe88a60ad.
//
// Solidity: function setBlobScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) SetBlobScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetBlobScalar(&_GasPriceOracleScrollL1ETH.TransactOpts, _scalar)
}

// SetCommitScalar is a paid mutator transaction binding the contract method 0x944b247f.
//
// Solidity: function setCommitScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) SetCommitScalar(opts *bind.TransactOpts, _scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "setCommitScalar", _scalar)
}

// SetCommitScalar is a paid mutator transaction binding the contract method 0x944b247f.
//
// Solidity: function setCommitScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) SetCommitScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetCommitScalar(&_GasPriceOracleScrollL1ETH.TransactOpts, _scalar)
}

// SetCommitScalar is a paid mutator transaction binding the contract method 0x944b247f.
//
// Solidity: function setCommitScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) SetCommitScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetCommitScalar(&_GasPriceOracleScrollL1ETH.TransactOpts, _scalar)
}

// SetL1BaseFee is a paid mutator transaction binding the contract method 0xbede39b5.
//
// Solidity: function setL1BaseFee(uint256 _l1BaseFee) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) SetL1BaseFee(opts *bind.TransactOpts, _l1BaseFee *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "setL1BaseFee", _l1BaseFee)
}

// SetL1BaseFee is a paid mutator transaction binding the contract method 0xbede39b5.
//
// Solidity: function setL1BaseFee(uint256 _l1BaseFee) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) SetL1BaseFee(_l1BaseFee *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetL1BaseFee(&_GasPriceOracleScrollL1ETH.TransactOpts, _l1BaseFee)
}

// SetL1BaseFee is a paid mutator transaction binding the contract method 0xbede39b5.
//
// Solidity: function setL1BaseFee(uint256 _l1BaseFee) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) SetL1BaseFee(_l1BaseFee *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetL1BaseFee(&_GasPriceOracleScrollL1ETH.TransactOpts, _l1BaseFee)
}

// SetL1BaseFeeAndBlobBaseFee is a paid mutator transaction binding the contract method 0x39455d3a.
//
// Solidity: function setL1BaseFeeAndBlobBaseFee(uint256 _l1BaseFee, uint256 _l1BlobBaseFee) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) SetL1BaseFeeAndBlobBaseFee(opts *bind.TransactOpts, _l1BaseFee *big.Int, _l1BlobBaseFee *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "setL1BaseFeeAndBlobBaseFee", _l1BaseFee, _l1BlobBaseFee)
}

// SetL1BaseFeeAndBlobBaseFee is a paid mutator transaction binding the contract method 0x39455d3a.
//
// Solidity: function setL1BaseFeeAndBlobBaseFee(uint256 _l1BaseFee, uint256 _l1BlobBaseFee) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) SetL1BaseFeeAndBlobBaseFee(_l1BaseFee *big.Int, _l1BlobBaseFee *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetL1BaseFeeAndBlobBaseFee(&_GasPriceOracleScrollL1ETH.TransactOpts, _l1BaseFee, _l1BlobBaseFee)
}

// SetL1BaseFeeAndBlobBaseFee is a paid mutator transaction binding the contract method 0x39455d3a.
//
// Solidity: function setL1BaseFeeAndBlobBaseFee(uint256 _l1BaseFee, uint256 _l1BlobBaseFee) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) SetL1BaseFeeAndBlobBaseFee(_l1BaseFee *big.Int, _l1BlobBaseFee *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetL1BaseFeeAndBlobBaseFee(&_GasPriceOracleScrollL1ETH.TransactOpts, _l1BaseFee, _l1BlobBaseFee)
}

// SetOverhead is a paid mutator transaction binding the contract method 0x3577afc5.
//
// Solidity: function setOverhead(uint256 _overhead) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) SetOverhead(opts *bind.TransactOpts, _overhead *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "setOverhead", _overhead)
}

// SetOverhead is a paid mutator transaction binding the contract method 0x3577afc5.
//
// Solidity: function setOverhead(uint256 _overhead) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) SetOverhead(_overhead *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetOverhead(&_GasPriceOracleScrollL1ETH.TransactOpts, _overhead)
}

// SetOverhead is a paid mutator transaction binding the contract method 0x3577afc5.
//
// Solidity: function setOverhead(uint256 _overhead) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) SetOverhead(_overhead *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetOverhead(&_GasPriceOracleScrollL1ETH.TransactOpts, _overhead)
}

// SetScalar is a paid mutator transaction binding the contract method 0x70465597.
//
// Solidity: function setScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) SetScalar(opts *bind.TransactOpts, _scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "setScalar", _scalar)
}

// SetScalar is a paid mutator transaction binding the contract method 0x70465597.
//
// Solidity: function setScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) SetScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetScalar(&_GasPriceOracleScrollL1ETH.TransactOpts, _scalar)
}

// SetScalar is a paid mutator transaction binding the contract method 0x70465597.
//
// Solidity: function setScalar(uint256 _scalar) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) SetScalar(_scalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.SetScalar(&_GasPriceOracleScrollL1ETH.TransactOpts, _scalar)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) TransferOwnership(opts *bind.TransactOpts, _newOwner common.Address) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "transferOwnership", _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.TransferOwnership(&_GasPriceOracleScrollL1ETH.TransactOpts, _newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _newOwner) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) TransferOwnership(_newOwner common.Address) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.TransferOwnership(&_GasPriceOracleScrollL1ETH.TransactOpts, _newOwner)
}

// UpdateWhitelist is a paid mutator transaction binding the contract method 0x3d0f963e.
//
// Solidity: function updateWhitelist(address _newWhitelist) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactor) UpdateWhitelist(opts *bind.TransactOpts, _newWhitelist common.Address) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.contract.Transact(opts, "updateWhitelist", _newWhitelist)
}

// UpdateWhitelist is a paid mutator transaction binding the contract method 0x3d0f963e.
//
// Solidity: function updateWhitelist(address _newWhitelist) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHSession) UpdateWhitelist(_newWhitelist common.Address) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.UpdateWhitelist(&_GasPriceOracleScrollL1ETH.TransactOpts, _newWhitelist)
}

// UpdateWhitelist is a paid mutator transaction binding the contract method 0x3d0f963e.
//
// Solidity: function updateWhitelist(address _newWhitelist) returns()
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHTransactorSession) UpdateWhitelist(_newWhitelist common.Address) (*types.Transaction, error) {
	return _GasPriceOracleScrollL1ETH.Contract.UpdateWhitelist(&_GasPriceOracleScrollL1ETH.TransactOpts, _newWhitelist)
}

// GasPriceOracleScrollL1ETHBlobScalarUpdatedIterator is returned from FilterBlobScalarUpdated and is used to iterate over the raw logs and unpacked data for BlobScalarUpdated events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHBlobScalarUpdatedIterator struct {
	Event *GasPriceOracleScrollL1ETHBlobScalarUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHBlobScalarUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHBlobScalarUpdated)
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
		it.Event = new(GasPriceOracleScrollL1ETHBlobScalarUpdated)
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
func (it *GasPriceOracleScrollL1ETHBlobScalarUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHBlobScalarUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHBlobScalarUpdated represents a BlobScalarUpdated event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHBlobScalarUpdated struct {
	Scalar *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBlobScalarUpdated is a free log retrieval operation binding the contract event 0x6b332a036d8c3ead57dcb06c87243bd7a2aed015ddf2d0528c2501dae56331aa.
//
// Solidity: event BlobScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterBlobScalarUpdated(opts *bind.FilterOpts) (*GasPriceOracleScrollL1ETHBlobScalarUpdatedIterator, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "BlobScalarUpdated")
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHBlobScalarUpdatedIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "BlobScalarUpdated", logs: logs, sub: sub}, nil
}

// WatchBlobScalarUpdated is a free log subscription operation binding the contract event 0x6b332a036d8c3ead57dcb06c87243bd7a2aed015ddf2d0528c2501dae56331aa.
//
// Solidity: event BlobScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchBlobScalarUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHBlobScalarUpdated) (event.Subscription, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "BlobScalarUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHBlobScalarUpdated)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "BlobScalarUpdated", log); err != nil {
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

// ParseBlobScalarUpdated is a log parse operation binding the contract event 0x6b332a036d8c3ead57dcb06c87243bd7a2aed015ddf2d0528c2501dae56331aa.
//
// Solidity: event BlobScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseBlobScalarUpdated(log types.Log) (*GasPriceOracleScrollL1ETHBlobScalarUpdated, error) {
	event := new(GasPriceOracleScrollL1ETHBlobScalarUpdated)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "BlobScalarUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleScrollL1ETHCommitScalarUpdatedIterator is returned from FilterCommitScalarUpdated and is used to iterate over the raw logs and unpacked data for CommitScalarUpdated events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHCommitScalarUpdatedIterator struct {
	Event *GasPriceOracleScrollL1ETHCommitScalarUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHCommitScalarUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHCommitScalarUpdated)
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
		it.Event = new(GasPriceOracleScrollL1ETHCommitScalarUpdated)
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
func (it *GasPriceOracleScrollL1ETHCommitScalarUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHCommitScalarUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHCommitScalarUpdated represents a CommitScalarUpdated event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHCommitScalarUpdated struct {
	Scalar *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCommitScalarUpdated is a free log retrieval operation binding the contract event 0x2ab3f5a4ebbcbf3c24f62f5454f52f10e1a8c9dcc5acac8f19199ce881a6a108.
//
// Solidity: event CommitScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterCommitScalarUpdated(opts *bind.FilterOpts) (*GasPriceOracleScrollL1ETHCommitScalarUpdatedIterator, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "CommitScalarUpdated")
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHCommitScalarUpdatedIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "CommitScalarUpdated", logs: logs, sub: sub}, nil
}

// WatchCommitScalarUpdated is a free log subscription operation binding the contract event 0x2ab3f5a4ebbcbf3c24f62f5454f52f10e1a8c9dcc5acac8f19199ce881a6a108.
//
// Solidity: event CommitScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchCommitScalarUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHCommitScalarUpdated) (event.Subscription, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "CommitScalarUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHCommitScalarUpdated)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "CommitScalarUpdated", log); err != nil {
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

// ParseCommitScalarUpdated is a log parse operation binding the contract event 0x2ab3f5a4ebbcbf3c24f62f5454f52f10e1a8c9dcc5acac8f19199ce881a6a108.
//
// Solidity: event CommitScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseCommitScalarUpdated(log types.Log) (*GasPriceOracleScrollL1ETHCommitScalarUpdated, error) {
	event := new(GasPriceOracleScrollL1ETHCommitScalarUpdated)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "CommitScalarUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleScrollL1ETHL1BaseFeeUpdatedIterator is returned from FilterL1BaseFeeUpdated and is used to iterate over the raw logs and unpacked data for L1BaseFeeUpdated events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHL1BaseFeeUpdatedIterator struct {
	Event *GasPriceOracleScrollL1ETHL1BaseFeeUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHL1BaseFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHL1BaseFeeUpdated)
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
		it.Event = new(GasPriceOracleScrollL1ETHL1BaseFeeUpdated)
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
func (it *GasPriceOracleScrollL1ETHL1BaseFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHL1BaseFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHL1BaseFeeUpdated represents a L1BaseFeeUpdated event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHL1BaseFeeUpdated struct {
	L1BaseFee *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterL1BaseFeeUpdated is a free log retrieval operation binding the contract event 0x351fb23757bb5ea0546c85b7996ddd7155f96b939ebaa5ff7bc49c75f27f2c44.
//
// Solidity: event L1BaseFeeUpdated(uint256 l1BaseFee)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterL1BaseFeeUpdated(opts *bind.FilterOpts) (*GasPriceOracleScrollL1ETHL1BaseFeeUpdatedIterator, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "L1BaseFeeUpdated")
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHL1BaseFeeUpdatedIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "L1BaseFeeUpdated", logs: logs, sub: sub}, nil
}

// WatchL1BaseFeeUpdated is a free log subscription operation binding the contract event 0x351fb23757bb5ea0546c85b7996ddd7155f96b939ebaa5ff7bc49c75f27f2c44.
//
// Solidity: event L1BaseFeeUpdated(uint256 l1BaseFee)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchL1BaseFeeUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHL1BaseFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "L1BaseFeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHL1BaseFeeUpdated)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "L1BaseFeeUpdated", log); err != nil {
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

// ParseL1BaseFeeUpdated is a log parse operation binding the contract event 0x351fb23757bb5ea0546c85b7996ddd7155f96b939ebaa5ff7bc49c75f27f2c44.
//
// Solidity: event L1BaseFeeUpdated(uint256 l1BaseFee)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseL1BaseFeeUpdated(log types.Log) (*GasPriceOracleScrollL1ETHL1BaseFeeUpdated, error) {
	event := new(GasPriceOracleScrollL1ETHL1BaseFeeUpdated)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "L1BaseFeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdatedIterator is returned from FilterL1BlobBaseFeeUpdated and is used to iterate over the raw logs and unpacked data for L1BlobBaseFeeUpdated events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdatedIterator struct {
	Event *GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated)
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
		it.Event = new(GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated)
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
func (it *GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated represents a L1BlobBaseFeeUpdated event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated struct {
	L1BlobBaseFee *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterL1BlobBaseFeeUpdated is a free log retrieval operation binding the contract event 0x9a14bfb5d18c4c3cf14cae19c23d7cf1bcede357ea40ca1f75cd49542c71c214.
//
// Solidity: event L1BlobBaseFeeUpdated(uint256 l1BlobBaseFee)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterL1BlobBaseFeeUpdated(opts *bind.FilterOpts) (*GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdatedIterator, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "L1BlobBaseFeeUpdated")
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdatedIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "L1BlobBaseFeeUpdated", logs: logs, sub: sub}, nil
}

// WatchL1BlobBaseFeeUpdated is a free log subscription operation binding the contract event 0x9a14bfb5d18c4c3cf14cae19c23d7cf1bcede357ea40ca1f75cd49542c71c214.
//
// Solidity: event L1BlobBaseFeeUpdated(uint256 l1BlobBaseFee)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchL1BlobBaseFeeUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "L1BlobBaseFeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "L1BlobBaseFeeUpdated", log); err != nil {
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

// ParseL1BlobBaseFeeUpdated is a log parse operation binding the contract event 0x9a14bfb5d18c4c3cf14cae19c23d7cf1bcede357ea40ca1f75cd49542c71c214.
//
// Solidity: event L1BlobBaseFeeUpdated(uint256 l1BlobBaseFee)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseL1BlobBaseFeeUpdated(log types.Log) (*GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated, error) {
	event := new(GasPriceOracleScrollL1ETHL1BlobBaseFeeUpdated)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "L1BlobBaseFeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleScrollL1ETHOverheadUpdatedIterator is returned from FilterOverheadUpdated and is used to iterate over the raw logs and unpacked data for OverheadUpdated events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHOverheadUpdatedIterator struct {
	Event *GasPriceOracleScrollL1ETHOverheadUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHOverheadUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHOverheadUpdated)
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
		it.Event = new(GasPriceOracleScrollL1ETHOverheadUpdated)
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
func (it *GasPriceOracleScrollL1ETHOverheadUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHOverheadUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHOverheadUpdated represents a OverheadUpdated event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHOverheadUpdated struct {
	Overhead *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOverheadUpdated is a free log retrieval operation binding the contract event 0x32740b35c0ea213650f60d44366b4fb211c9033b50714e4a1d34e65d5beb9bb4.
//
// Solidity: event OverheadUpdated(uint256 overhead)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterOverheadUpdated(opts *bind.FilterOpts) (*GasPriceOracleScrollL1ETHOverheadUpdatedIterator, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "OverheadUpdated")
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHOverheadUpdatedIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "OverheadUpdated", logs: logs, sub: sub}, nil
}

// WatchOverheadUpdated is a free log subscription operation binding the contract event 0x32740b35c0ea213650f60d44366b4fb211c9033b50714e4a1d34e65d5beb9bb4.
//
// Solidity: event OverheadUpdated(uint256 overhead)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchOverheadUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHOverheadUpdated) (event.Subscription, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "OverheadUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHOverheadUpdated)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "OverheadUpdated", log); err != nil {
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

// ParseOverheadUpdated is a log parse operation binding the contract event 0x32740b35c0ea213650f60d44366b4fb211c9033b50714e4a1d34e65d5beb9bb4.
//
// Solidity: event OverheadUpdated(uint256 overhead)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseOverheadUpdated(log types.Log) (*GasPriceOracleScrollL1ETHOverheadUpdated, error) {
	event := new(GasPriceOracleScrollL1ETHOverheadUpdated)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "OverheadUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleScrollL1ETHOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHOwnershipTransferredIterator struct {
	Event *GasPriceOracleScrollL1ETHOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHOwnershipTransferred)
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
		it.Event = new(GasPriceOracleScrollL1ETHOwnershipTransferred)
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
func (it *GasPriceOracleScrollL1ETHOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHOwnershipTransferred represents a OwnershipTransferred event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHOwnershipTransferred struct {
	OldOwner common.Address
	NewOwner common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed _oldOwner, address indexed _newOwner)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, _oldOwner []common.Address, _newOwner []common.Address) (*GasPriceOracleScrollL1ETHOwnershipTransferredIterator, error) {

	var _oldOwnerRule []interface{}
	for _, _oldOwnerItem := range _oldOwner {
		_oldOwnerRule = append(_oldOwnerRule, _oldOwnerItem)
	}
	var _newOwnerRule []interface{}
	for _, _newOwnerItem := range _newOwner {
		_newOwnerRule = append(_newOwnerRule, _newOwnerItem)
	}

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "OwnershipTransferred", _oldOwnerRule, _newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHOwnershipTransferredIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed _oldOwner, address indexed _newOwner)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHOwnershipTransferred, _oldOwner []common.Address, _newOwner []common.Address) (event.Subscription, error) {

	var _oldOwnerRule []interface{}
	for _, _oldOwnerItem := range _oldOwner {
		_oldOwnerRule = append(_oldOwnerRule, _oldOwnerItem)
	}
	var _newOwnerRule []interface{}
	for _, _newOwnerItem := range _newOwner {
		_newOwnerRule = append(_newOwnerRule, _newOwnerItem)
	}

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "OwnershipTransferred", _oldOwnerRule, _newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHOwnershipTransferred)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
// Solidity: event OwnershipTransferred(address indexed _oldOwner, address indexed _newOwner)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseOwnershipTransferred(log types.Log) (*GasPriceOracleScrollL1ETHOwnershipTransferred, error) {
	event := new(GasPriceOracleScrollL1ETHOwnershipTransferred)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleScrollL1ETHScalarUpdatedIterator is returned from FilterScalarUpdated and is used to iterate over the raw logs and unpacked data for ScalarUpdated events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHScalarUpdatedIterator struct {
	Event *GasPriceOracleScrollL1ETHScalarUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHScalarUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHScalarUpdated)
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
		it.Event = new(GasPriceOracleScrollL1ETHScalarUpdated)
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
func (it *GasPriceOracleScrollL1ETHScalarUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHScalarUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHScalarUpdated represents a ScalarUpdated event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHScalarUpdated struct {
	Scalar *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterScalarUpdated is a free log retrieval operation binding the contract event 0x3336cd9708eaf2769a0f0dc0679f30e80f15dcd88d1921b5a16858e8b85c591a.
//
// Solidity: event ScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterScalarUpdated(opts *bind.FilterOpts) (*GasPriceOracleScrollL1ETHScalarUpdatedIterator, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "ScalarUpdated")
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHScalarUpdatedIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "ScalarUpdated", logs: logs, sub: sub}, nil
}

// WatchScalarUpdated is a free log subscription operation binding the contract event 0x3336cd9708eaf2769a0f0dc0679f30e80f15dcd88d1921b5a16858e8b85c591a.
//
// Solidity: event ScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchScalarUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHScalarUpdated) (event.Subscription, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "ScalarUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHScalarUpdated)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "ScalarUpdated", log); err != nil {
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

// ParseScalarUpdated is a log parse operation binding the contract event 0x3336cd9708eaf2769a0f0dc0679f30e80f15dcd88d1921b5a16858e8b85c591a.
//
// Solidity: event ScalarUpdated(uint256 scalar)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseScalarUpdated(log types.Log) (*GasPriceOracleScrollL1ETHScalarUpdated, error) {
	event := new(GasPriceOracleScrollL1ETHScalarUpdated)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "ScalarUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleScrollL1ETHUpdateWhitelistIterator is returned from FilterUpdateWhitelist and is used to iterate over the raw logs and unpacked data for UpdateWhitelist events raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHUpdateWhitelistIterator struct {
	Event *GasPriceOracleScrollL1ETHUpdateWhitelist // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleScrollL1ETHUpdateWhitelistIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleScrollL1ETHUpdateWhitelist)
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
		it.Event = new(GasPriceOracleScrollL1ETHUpdateWhitelist)
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
func (it *GasPriceOracleScrollL1ETHUpdateWhitelistIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleScrollL1ETHUpdateWhitelistIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleScrollL1ETHUpdateWhitelist represents a UpdateWhitelist event raised by the GasPriceOracleScrollL1ETH contract.
type GasPriceOracleScrollL1ETHUpdateWhitelist struct {
	OldWhitelist common.Address
	NewWhitelist common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterUpdateWhitelist is a free log retrieval operation binding the contract event 0x22d1c35fe072d2e42c3c8f9bd4a0d34aa84a0101d020a62517b33fdb3174e5f7.
//
// Solidity: event UpdateWhitelist(address _oldWhitelist, address _newWhitelist)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) FilterUpdateWhitelist(opts *bind.FilterOpts) (*GasPriceOracleScrollL1ETHUpdateWhitelistIterator, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.FilterLogs(opts, "UpdateWhitelist")
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleScrollL1ETHUpdateWhitelistIterator{contract: _GasPriceOracleScrollL1ETH.contract, event: "UpdateWhitelist", logs: logs, sub: sub}, nil
}

// WatchUpdateWhitelist is a free log subscription operation binding the contract event 0x22d1c35fe072d2e42c3c8f9bd4a0d34aa84a0101d020a62517b33fdb3174e5f7.
//
// Solidity: event UpdateWhitelist(address _oldWhitelist, address _newWhitelist)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) WatchUpdateWhitelist(opts *bind.WatchOpts, sink chan<- *GasPriceOracleScrollL1ETHUpdateWhitelist) (event.Subscription, error) {

	logs, sub, err := _GasPriceOracleScrollL1ETH.contract.WatchLogs(opts, "UpdateWhitelist")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleScrollL1ETHUpdateWhitelist)
				if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "UpdateWhitelist", log); err != nil {
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

// ParseUpdateWhitelist is a log parse operation binding the contract event 0x22d1c35fe072d2e42c3c8f9bd4a0d34aa84a0101d020a62517b33fdb3174e5f7.
//
// Solidity: event UpdateWhitelist(address _oldWhitelist, address _newWhitelist)
func (_GasPriceOracleScrollL1ETH *GasPriceOracleScrollL1ETHFilterer) ParseUpdateWhitelist(log types.Log) (*GasPriceOracleScrollL1ETHUpdateWhitelist, error) {
	event := new(GasPriceOracleScrollL1ETHUpdateWhitelist)
	if err := _GasPriceOracleScrollL1ETH.contract.UnpackLog(event, "UpdateWhitelist", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
