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

// ILiquidityDepositData is an auto generated low-level Go binding around an user-defined struct.
type ILiquidityDepositData struct {
	Recipient   [32]byte
	Amount      *big.Int
	Sender      common.Address
	TokenIndex  uint32
	IsRejected  bool
	RequestedAt *big.Int
}

// ILiquidityTokenInfo is an auto generated low-level Go binding around an user-defined struct.
type ILiquidityTokenInfo struct {
	TokenType    uint8
	TokenAddress common.Address
	TokenId      *big.Int
}

// LiquidityMetaData contains all meta data concerning the Liquidity contract.
var LiquidityMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"scrollMessenger\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"depositIndex\",\"type\":\"uint64\"}],\"name\":\"CanceledDeposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"depositIndex\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposited\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"_rollupContract\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_scrollMessenger\",\"outputs\":[{\"internalType\":\"contractIScrollMessenger\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"depositIndex\",\"type\":\"uint64\"}],\"name\":\"cancelDeposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"depositERC20\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"depositERC721\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"}],\"name\":\"depositETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"depositIndex\",\"type\":\"uint64\"}],\"name\":\"getDeposit\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"},{\"internalType\":\"bool\",\"name\":\"isRejected\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"requestedAt\",\"type\":\"uint256\"}],\"internalType\":\"structILiquidity.DepositData\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDepositCounter\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLastProcessedDepositIndex\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLastSeenDepositIndex\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"enumILiquidity.TokenType\",\"name\":\"tokenType\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getTokenIndex\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"tokenIndex\",\"type\":\"uint32\"}],\"name\":\"getTokenInfo\",\"outputs\":[{\"components\":[{\"internalType\":\"enumILiquidity.TokenType\",\"name\":\"tokenType\",\"type\":\"uint8\"},{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"internalType\":\"structILiquidity.TokenInfo\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"lastSeenDepositIndex\",\"type\":\"uint64\"},{\"internalType\":\"uint64[]\",\"name\":\"rejectedIndices\",\"type\":\"uint64[]\"}],\"name\":\"rejectDeposits\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"lastProcessedDepositIndex\",\"type\":\"uint64\"}],\"name\":\"submitDepositRoot\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newRollupContract\",\"type\":\"address\"}],\"name\":\"updateRollupContract\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// RollupContract is a free data retrieval call binding the contract method 0x6a49c47d.
//
// Solidity: function _rollupContract() view returns(address)
func (_Liquidity *LiquidityCaller) RollupContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "_rollupContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err
}

// RollupContract is a free data retrieval call binding the contract method 0x6a49c47d.
//
// Solidity: function _rollupContract() view returns(address)
func (_Liquidity *LiquiditySession) RollupContract() (common.Address, error) {
	return _Liquidity.Contract.RollupContract(&_Liquidity.CallOpts)
}

// RollupContract is a free data retrieval call binding the contract method 0x6a49c47d.
//
// Solidity: function _rollupContract() view returns(address)
func (_Liquidity *LiquidityCallerSession) RollupContract() (common.Address, error) {
	return _Liquidity.Contract.RollupContract(&_Liquidity.CallOpts)
}

// ScrollMessenger is a free data retrieval call binding the contract method 0xaa8fc35b.
//
// Solidity: function _scrollMessenger() view returns(address)
func (_Liquidity *LiquidityCaller) ScrollMessenger(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "_scrollMessenger")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err
}

// ScrollMessenger is a free data retrieval call binding the contract method 0xaa8fc35b.
//
// Solidity: function _scrollMessenger() view returns(address)
func (_Liquidity *LiquiditySession) ScrollMessenger() (common.Address, error) {
	return _Liquidity.Contract.ScrollMessenger(&_Liquidity.CallOpts)
}

// ScrollMessenger is a free data retrieval call binding the contract method 0xaa8fc35b.
//
// Solidity: function _scrollMessenger() view returns(address)
func (_Liquidity *LiquidityCallerSession) ScrollMessenger() (common.Address, error) {
	return _Liquidity.Contract.ScrollMessenger(&_Liquidity.CallOpts)
}

// GetDeposit is a free data retrieval call binding the contract method 0xae59a6b2.
//
// Solidity: function getDeposit(uint64 depositIndex) view returns((bytes32,uint256,address,uint32,bool,uint256))
func (_Liquidity *LiquidityCaller) GetDeposit(opts *bind.CallOpts, depositIndex uint64) (ILiquidityDepositData, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getDeposit", depositIndex)

	if err != nil {
		return *new(ILiquidityDepositData), err
	}

	out0 := *abi.ConvertType(out[0], new(ILiquidityDepositData)).(*ILiquidityDepositData)

	return out0, err
}

// GetDeposit is a free data retrieval call binding the contract method 0xae59a6b2.
//
// Solidity: function getDeposit(uint64 depositIndex) view returns((bytes32,uint256,address,uint32,bool,uint256))
func (_Liquidity *LiquiditySession) GetDeposit(depositIndex uint64) (ILiquidityDepositData, error) {
	return _Liquidity.Contract.GetDeposit(&_Liquidity.CallOpts, depositIndex)
}

// GetDeposit is a free data retrieval call binding the contract method 0xae59a6b2.
//
// Solidity: function getDeposit(uint64 depositIndex) view returns((bytes32,uint256,address,uint32,bool,uint256))
func (_Liquidity *LiquidityCallerSession) GetDeposit(depositIndex uint64) (ILiquidityDepositData, error) {
	return _Liquidity.Contract.GetDeposit(&_Liquidity.CallOpts, depositIndex)
}

// GetDepositCounter is a free data retrieval call binding the contract method 0xd28906c9.
//
// Solidity: function getDepositCounter() view returns(uint64)
func (_Liquidity *LiquidityCaller) GetDepositCounter(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getDepositCounter")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err
}

// GetDepositCounter is a free data retrieval call binding the contract method 0xd28906c9.
//
// Solidity: function getDepositCounter() view returns(uint64)
func (_Liquidity *LiquiditySession) GetDepositCounter() (uint64, error) {
	return _Liquidity.Contract.GetDepositCounter(&_Liquidity.CallOpts)
}

// GetDepositCounter is a free data retrieval call binding the contract method 0xd28906c9.
//
// Solidity: function getDepositCounter() view returns(uint64)
func (_Liquidity *LiquidityCallerSession) GetDepositCounter() (uint64, error) {
	return _Liquidity.Contract.GetDepositCounter(&_Liquidity.CallOpts)
}

// GetLastProcessedDepositIndex is a free data retrieval call binding the contract method 0x107ef267.
//
// Solidity: function getLastProcessedDepositIndex() view returns(uint64)
func (_Liquidity *LiquidityCaller) GetLastProcessedDepositIndex(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getLastProcessedDepositIndex")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err
}

// GetLastProcessedDepositIndex is a free data retrieval call binding the contract method 0x107ef267.
//
// Solidity: function getLastProcessedDepositIndex() view returns(uint64)
func (_Liquidity *LiquiditySession) GetLastProcessedDepositIndex() (uint64, error) {
	return _Liquidity.Contract.GetLastProcessedDepositIndex(&_Liquidity.CallOpts)
}

// GetLastProcessedDepositIndex is a free data retrieval call binding the contract method 0x107ef267.
//
// Solidity: function getLastProcessedDepositIndex() view returns(uint64)
func (_Liquidity *LiquidityCallerSession) GetLastProcessedDepositIndex() (uint64, error) {
	return _Liquidity.Contract.GetLastProcessedDepositIndex(&_Liquidity.CallOpts)
}

// GetLastSeenDepositIndex is a free data retrieval call binding the contract method 0xbeb1e041.
//
// Solidity: function getLastSeenDepositIndex() view returns(uint64)
func (_Liquidity *LiquidityCaller) GetLastSeenDepositIndex(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getLastSeenDepositIndex")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err
}

// GetLastSeenDepositIndex is a free data retrieval call binding the contract method 0xbeb1e041.
//
// Solidity: function getLastSeenDepositIndex() view returns(uint64)
func (_Liquidity *LiquiditySession) GetLastSeenDepositIndex() (uint64, error) {
	return _Liquidity.Contract.GetLastSeenDepositIndex(&_Liquidity.CallOpts)
}

// GetLastSeenDepositIndex is a free data retrieval call binding the contract method 0xbeb1e041.
//
// Solidity: function getLastSeenDepositIndex() view returns(uint64)
func (_Liquidity *LiquidityCallerSession) GetLastSeenDepositIndex() (uint64, error) {
	return _Liquidity.Contract.GetLastSeenDepositIndex(&_Liquidity.CallOpts)
}

// GetTokenIndex is a free data retrieval call binding the contract method 0x381175f6.
//
// Solidity: function getTokenIndex(uint8 tokenType, address tokenAddress, uint256 tokenId) view returns(uint32)
func (_Liquidity *LiquidityCaller) GetTokenIndex(opts *bind.CallOpts, tokenType uint8, tokenAddress common.Address, tokenId *big.Int) (uint32, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getTokenIndex", tokenType, tokenAddress, tokenId)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err
}

// GetTokenIndex is a free data retrieval call binding the contract method 0x381175f6.
//
// Solidity: function getTokenIndex(uint8 tokenType, address tokenAddress, uint256 tokenId) view returns(uint32)
func (_Liquidity *LiquiditySession) GetTokenIndex(tokenType uint8, tokenAddress common.Address, tokenId *big.Int) (uint32, error) {
	return _Liquidity.Contract.GetTokenIndex(&_Liquidity.CallOpts, tokenType, tokenAddress, tokenId)
}

// GetTokenIndex is a free data retrieval call binding the contract method 0x381175f6.
//
// Solidity: function getTokenIndex(uint8 tokenType, address tokenAddress, uint256 tokenId) view returns(uint32)
func (_Liquidity *LiquidityCallerSession) GetTokenIndex(tokenType uint8, tokenAddress common.Address, tokenId *big.Int) (uint32, error) {
	return _Liquidity.Contract.GetTokenIndex(&_Liquidity.CallOpts, tokenType, tokenAddress, tokenId)
}

// GetTokenInfo is a free data retrieval call binding the contract method 0xcbd90cd0.
//
// Solidity: function getTokenInfo(uint32 tokenIndex) view returns((uint8,address,uint256))
func (_Liquidity *LiquidityCaller) GetTokenInfo(opts *bind.CallOpts, tokenIndex uint32) (ILiquidityTokenInfo, error) {
	var out []interface{}
	err := _Liquidity.contract.Call(opts, &out, "getTokenInfo", tokenIndex)

	if err != nil {
		return *new(ILiquidityTokenInfo), err
	}

	out0 := *abi.ConvertType(out[0], new(ILiquidityTokenInfo)).(*ILiquidityTokenInfo)

	return out0, err
}

// GetTokenInfo is a free data retrieval call binding the contract method 0xcbd90cd0.
//
// Solidity: function getTokenInfo(uint32 tokenIndex) view returns((uint8,address,uint256))
func (_Liquidity *LiquiditySession) GetTokenInfo(tokenIndex uint32) (ILiquidityTokenInfo, error) {
	return _Liquidity.Contract.GetTokenInfo(&_Liquidity.CallOpts, tokenIndex)
}

// GetTokenInfo is a free data retrieval call binding the contract method 0xcbd90cd0.
//
// Solidity: function getTokenInfo(uint32 tokenIndex) view returns((uint8,address,uint256))
func (_Liquidity *LiquidityCallerSession) GetTokenInfo(tokenIndex uint32) (ILiquidityTokenInfo, error) {
	return _Liquidity.Contract.GetTokenInfo(&_Liquidity.CallOpts, tokenIndex)
}

// CancelDeposit is a paid mutator transaction binding the contract method 0x575d5907.
//
// Solidity: function cancelDeposit(uint64 depositIndex) returns()
func (_Liquidity *LiquidityTransactor) CancelDeposit(opts *bind.TransactOpts, depositIndex uint64) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "cancelDeposit", depositIndex)
}

// CancelDeposit is a paid mutator transaction binding the contract method 0x575d5907.
//
// Solidity: function cancelDeposit(uint64 depositIndex) returns()
func (_Liquidity *LiquiditySession) CancelDeposit(depositIndex uint64) (*types.Transaction, error) {
	return _Liquidity.Contract.CancelDeposit(&_Liquidity.TransactOpts, depositIndex)
}

// CancelDeposit is a paid mutator transaction binding the contract method 0x575d5907.
//
// Solidity: function cancelDeposit(uint64 depositIndex) returns()
func (_Liquidity *LiquidityTransactorSession) CancelDeposit(depositIndex uint64) (*types.Transaction, error) {
	return _Liquidity.Contract.CancelDeposit(&_Liquidity.TransactOpts, depositIndex)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x788c2ecf.
//
// Solidity: function depositERC20(address tokenAddress, bytes32 recipient, uint256 amount) returns()
func (_Liquidity *LiquidityTransactor) DepositERC20(opts *bind.TransactOpts, tokenAddress common.Address, recipient [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "depositERC20", tokenAddress, recipient, amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x788c2ecf.
//
// Solidity: function depositERC20(address tokenAddress, bytes32 recipient, uint256 amount) returns()
func (_Liquidity *LiquiditySession) DepositERC20(tokenAddress common.Address, recipient [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC20(&_Liquidity.TransactOpts, tokenAddress, recipient, amount)
}

// DepositERC20 is a paid mutator transaction binding the contract method 0x788c2ecf.
//
// Solidity: function depositERC20(address tokenAddress, bytes32 recipient, uint256 amount) returns()
func (_Liquidity *LiquidityTransactorSession) DepositERC20(tokenAddress common.Address, recipient [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC20(&_Liquidity.TransactOpts, tokenAddress, recipient, amount)
}

// DepositERC721 is a paid mutator transaction binding the contract method 0x9f401f58.
//
// Solidity: function depositERC721(address tokenAddress, bytes32 recipient, uint256 tokenId) returns()
func (_Liquidity *LiquidityTransactor) DepositERC721(opts *bind.TransactOpts, tokenAddress common.Address, recipient [32]byte, tokenId *big.Int) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "depositERC721", tokenAddress, recipient, tokenId)
}

// DepositERC721 is a paid mutator transaction binding the contract method 0x9f401f58.
//
// Solidity: function depositERC721(address tokenAddress, bytes32 recipient, uint256 tokenId) returns()
func (_Liquidity *LiquiditySession) DepositERC721(tokenAddress common.Address, recipient [32]byte, tokenId *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC721(&_Liquidity.TransactOpts, tokenAddress, recipient, tokenId)
}

// DepositERC721 is a paid mutator transaction binding the contract method 0x9f401f58.
//
// Solidity: function depositERC721(address tokenAddress, bytes32 recipient, uint256 tokenId) returns()
func (_Liquidity *LiquidityTransactorSession) DepositERC721(tokenAddress common.Address, recipient [32]byte, tokenId *big.Int) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositERC721(&_Liquidity.TransactOpts, tokenAddress, recipient, tokenId)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 recipient) payable returns()
func (_Liquidity *LiquidityTransactor) DepositETH(opts *bind.TransactOpts, recipient [32]byte) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "depositETH", recipient)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 recipient) payable returns()
func (_Liquidity *LiquiditySession) DepositETH(recipient [32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositETH(&_Liquidity.TransactOpts, recipient)
}

// DepositETH is a paid mutator transaction binding the contract method 0xd68d9d4e.
//
// Solidity: function depositETH(bytes32 recipient) payable returns()
func (_Liquidity *LiquidityTransactorSession) DepositETH(recipient [32]byte) (*types.Transaction, error) {
	return _Liquidity.Contract.DepositETH(&_Liquidity.TransactOpts, recipient)
}

// RejectDeposits is a paid mutator transaction binding the contract method 0xe114c7c6.
//
// Solidity: function rejectDeposits(uint64 lastSeenDepositIndex, uint64[] rejectedIndices) returns()
func (_Liquidity *LiquidityTransactor) RejectDeposits(opts *bind.TransactOpts, lastSeenDepositIndex uint64, rejectedIndices []uint64) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "rejectDeposits", lastSeenDepositIndex, rejectedIndices)
}

// RejectDeposits is a paid mutator transaction binding the contract method 0xe114c7c6.
//
// Solidity: function rejectDeposits(uint64 lastSeenDepositIndex, uint64[] rejectedIndices) returns()
func (_Liquidity *LiquiditySession) RejectDeposits(lastSeenDepositIndex uint64, rejectedIndices []uint64) (*types.Transaction, error) {
	return _Liquidity.Contract.RejectDeposits(&_Liquidity.TransactOpts, lastSeenDepositIndex, rejectedIndices)
}

// RejectDeposits is a paid mutator transaction binding the contract method 0xe114c7c6.
//
// Solidity: function rejectDeposits(uint64 lastSeenDepositIndex, uint64[] rejectedIndices) returns()
func (_Liquidity *LiquidityTransactorSession) RejectDeposits(lastSeenDepositIndex uint64, rejectedIndices []uint64) (*types.Transaction, error) {
	return _Liquidity.Contract.RejectDeposits(&_Liquidity.TransactOpts, lastSeenDepositIndex, rejectedIndices)
}

// SubmitDepositRoot is a paid mutator transaction binding the contract method 0xfb564bf1.
//
// Solidity: function submitDepositRoot(uint64 lastProcessedDepositIndex) returns()
func (_Liquidity *LiquidityTransactor) SubmitDepositRoot(opts *bind.TransactOpts, lastProcessedDepositIndex uint64) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "submitDepositRoot", lastProcessedDepositIndex)
}

// SubmitDepositRoot is a paid mutator transaction binding the contract method 0xfb564bf1.
//
// Solidity: function submitDepositRoot(uint64 lastProcessedDepositIndex) returns()
func (_Liquidity *LiquiditySession) SubmitDepositRoot(lastProcessedDepositIndex uint64) (*types.Transaction, error) {
	return _Liquidity.Contract.SubmitDepositRoot(&_Liquidity.TransactOpts, lastProcessedDepositIndex)
}

// SubmitDepositRoot is a paid mutator transaction binding the contract method 0xfb564bf1.
//
// Solidity: function submitDepositRoot(uint64 lastProcessedDepositIndex) returns()
func (_Liquidity *LiquidityTransactorSession) SubmitDepositRoot(lastProcessedDepositIndex uint64) (*types.Transaction, error) {
	return _Liquidity.Contract.SubmitDepositRoot(&_Liquidity.TransactOpts, lastProcessedDepositIndex)
}

// UpdateRollupContract is a paid mutator transaction binding the contract method 0x08c5ca5c.
//
// Solidity: function updateRollupContract(address newRollupContract) returns()
func (_Liquidity *LiquidityTransactor) UpdateRollupContract(opts *bind.TransactOpts, newRollupContract common.Address) (*types.Transaction, error) {
	return _Liquidity.contract.Transact(opts, "updateRollupContract", newRollupContract)
}

// UpdateRollupContract is a paid mutator transaction binding the contract method 0x08c5ca5c.
//
// Solidity: function updateRollupContract(address newRollupContract) returns()
func (_Liquidity *LiquiditySession) UpdateRollupContract(newRollupContract common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.UpdateRollupContract(&_Liquidity.TransactOpts, newRollupContract)
}

// UpdateRollupContract is a paid mutator transaction binding the contract method 0x08c5ca5c.
//
// Solidity: function updateRollupContract(address newRollupContract) returns()
func (_Liquidity *LiquidityTransactorSession) UpdateRollupContract(newRollupContract common.Address) (*types.Transaction, error) {
	return _Liquidity.Contract.UpdateRollupContract(&_Liquidity.TransactOpts, newRollupContract)
}

// LiquidityCanceledDepositIterator is returned from FilterCanceledDeposit and is used to iterate over the raw logs and unpacked data for CanceledDeposit events raised by the Liquidity contract.
type LiquidityCanceledDepositIterator struct {
	Event *LiquidityCanceledDeposit // Event containing the contract specifics and raw log

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
func (it *LiquidityCanceledDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LiquidityCanceledDeposit)
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
		it.Event = new(LiquidityCanceledDeposit)
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
func (it *LiquidityCanceledDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LiquidityCanceledDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LiquidityCanceledDeposit represents a CanceledDeposit event raised by the Liquidity contract.
type LiquidityCanceledDeposit struct {
	DepositIndex uint64
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCanceledDeposit is a free log retrieval operation binding the contract event 0x1e04cb4168296ea5987ab05430a88e6b235ceb9f8b9b2c6367e6ca5b96e85dbb.
//
// Solidity: event CanceledDeposit(uint64 indexed depositIndex)
func (_Liquidity *LiquidityFilterer) FilterCanceledDeposit(opts *bind.FilterOpts, depositIndex []uint64) (*LiquidityCanceledDepositIterator, error) {
	var depositIndexRule []interface{}
	for _, depositIndexItem := range depositIndex {
		depositIndexRule = append(depositIndexRule, depositIndexItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "CanceledDeposit", depositIndexRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityCanceledDepositIterator{contract: _Liquidity.contract, event: "CanceledDeposit", logs: logs, sub: sub}, nil
}

// WatchCanceledDeposit is a free log subscription operation binding the contract event 0x1e04cb4168296ea5987ab05430a88e6b235ceb9f8b9b2c6367e6ca5b96e85dbb.
//
// Solidity: event CanceledDeposit(uint64 indexed depositIndex)
func (_Liquidity *LiquidityFilterer) WatchCanceledDeposit(opts *bind.WatchOpts, sink chan<- *LiquidityCanceledDeposit, depositIndex []uint64) (event.Subscription, error) {
	var depositIndexRule []interface{}
	for _, depositIndexItem := range depositIndex {
		depositIndexRule = append(depositIndexRule, depositIndexItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "CanceledDeposit", depositIndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LiquidityCanceledDeposit)
				if err := _Liquidity.contract.UnpackLog(event, "CanceledDeposit", log); err != nil {
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

// ParseCanceledDeposit is a log parse operation binding the contract event 0x1e04cb4168296ea5987ab05430a88e6b235ceb9f8b9b2c6367e6ca5b96e85dbb.
//
// Solidity: event CanceledDeposit(uint64 indexed depositIndex)
func (_Liquidity *LiquidityFilterer) ParseCanceledDeposit(log types.Log) (*LiquidityCanceledDeposit, error) {
	event := new(LiquidityCanceledDeposit)
	if err := _Liquidity.contract.UnpackLog(event, "CanceledDeposit", log); err != nil {
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
	Recipient    [32]byte
	DepositIndex uint64
	Sender       common.Address
	TokenIndex   uint32
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0xc1617bc4d3903289f4b5ba712d4f52cf520653227e0db8a358e462296f2da587.
//
// Solidity: event Deposited(bytes32 indexed recipient, uint64 indexed depositIndex, address indexed sender, uint32 tokenIndex, uint256 amount)
func (_Liquidity *LiquidityFilterer) FilterDeposited(opts *bind.FilterOpts, recipient [][32]byte, depositIndex []uint64, sender []common.Address) (*LiquidityDepositedIterator, error) {
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var depositIndexRule []interface{}
	for _, depositIndexItem := range depositIndex {
		depositIndexRule = append(depositIndexRule, depositIndexItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.FilterLogs(opts, "Deposited", recipientRule, depositIndexRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &LiquidityDepositedIterator{contract: _Liquidity.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0xc1617bc4d3903289f4b5ba712d4f52cf520653227e0db8a358e462296f2da587.
//
// Solidity: event Deposited(bytes32 indexed recipient, uint64 indexed depositIndex, address indexed sender, uint32 tokenIndex, uint256 amount)
func (_Liquidity *LiquidityFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *LiquidityDeposited, recipient [][32]byte, depositIndex []uint64, sender []common.Address) (event.Subscription, error) {
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var depositIndexRule []interface{}
	for _, depositIndexItem := range depositIndex {
		depositIndexRule = append(depositIndexRule, depositIndexItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Liquidity.contract.WatchLogs(opts, "Deposited", recipientRule, depositIndexRule, senderRule)
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

// ParseDeposited is a log parse operation binding the contract event 0xc1617bc4d3903289f4b5ba712d4f52cf520653227e0db8a358e462296f2da587.
//
// Solidity: event Deposited(bytes32 indexed recipient, uint64 indexed depositIndex, address indexed sender, uint32 tokenIndex, uint256 amount)
func (_Liquidity *LiquidityFilterer) ParseDeposited(log types.Log) (*LiquidityDeposited, error) {
	event := new(LiquidityDeposited)
	if err := _Liquidity.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
