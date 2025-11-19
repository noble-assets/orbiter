// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, NASD Inc. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package types

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

// OrbiterGatewayCCTPMetaData contains all meta data concerning the OrbiterGatewayCCTP contract.
var OrbiterGatewayCCTPMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"token_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"tokenMessenger_\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"destinationCaller_\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DESTINATION_CALLER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DESTINATION_DOMAIN\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MESSAGE_TRANSMITTER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIMessageTransmitter\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MINT_RECIPIENT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"TOKEN\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIFiatToken\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"TOKEN_MESSENGER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractITokenMessenger\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"depositForBurnWithOrbiter\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"blocktimeDeadline\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"v\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"r\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"orbiterPayload\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"destinationCaller\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"DepositForBurnWithOrbiter\",\"inputs\":[{\"name\":\"transferNonce\",\"type\":\"uint64\",\"indexed\":true,\"internalType\":\"uint64\"},{\"name\":\"payloadNonce\",\"type\":\"uint64\",\"indexed\":true,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"ApproveFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"TransferFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroDestinationCaller\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroTokenAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroTokenMessengerAddress\",\"inputs\":[]}]",
}

// OrbiterGatewayCCTPABI is the input ABI used to generate the binding from.
// Deprecated: Use OrbiterGatewayCCTPMetaData.ABI instead.
var OrbiterGatewayCCTPABI = OrbiterGatewayCCTPMetaData.ABI

// OrbiterGatewayCCTP is an auto generated Go binding around an Ethereum contract.
type OrbiterGatewayCCTP struct {
	OrbiterGatewayCCTPCaller     // Read-only binding to the contract
	OrbiterGatewayCCTPTransactor // Write-only binding to the contract
	OrbiterGatewayCCTPFilterer   // Log filterer for contract events
}

// OrbiterGatewayCCTPCaller is an auto generated read-only Go binding around an Ethereum contract.
type OrbiterGatewayCCTPCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrbiterGatewayCCTPTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OrbiterGatewayCCTPTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrbiterGatewayCCTPFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OrbiterGatewayCCTPFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OrbiterGatewayCCTPSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OrbiterGatewayCCTPSession struct {
	Contract     *OrbiterGatewayCCTP // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// OrbiterGatewayCCTPCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OrbiterGatewayCCTPCallerSession struct {
	Contract *OrbiterGatewayCCTPCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// OrbiterGatewayCCTPTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OrbiterGatewayCCTPTransactorSession struct {
	Contract     *OrbiterGatewayCCTPTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// OrbiterGatewayCCTPRaw is an auto generated low-level Go binding around an Ethereum contract.
type OrbiterGatewayCCTPRaw struct {
	Contract *OrbiterGatewayCCTP // Generic contract binding to access the raw methods on
}

// OrbiterGatewayCCTPCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OrbiterGatewayCCTPCallerRaw struct {
	Contract *OrbiterGatewayCCTPCaller // Generic read-only contract binding to access the raw methods on
}

// OrbiterGatewayCCTPTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OrbiterGatewayCCTPTransactorRaw struct {
	Contract *OrbiterGatewayCCTPTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOrbiterGatewayCCTP creates a new instance of OrbiterGatewayCCTP, bound to a specific deployed contract.
func NewOrbiterGatewayCCTP(address common.Address, backend bind.ContractBackend) (*OrbiterGatewayCCTP, error) {
	contract, err := bindOrbiterGatewayCCTP(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OrbiterGatewayCCTP{OrbiterGatewayCCTPCaller: OrbiterGatewayCCTPCaller{contract: contract}, OrbiterGatewayCCTPTransactor: OrbiterGatewayCCTPTransactor{contract: contract}, OrbiterGatewayCCTPFilterer: OrbiterGatewayCCTPFilterer{contract: contract}}, nil
}

// NewOrbiterGatewayCCTPCaller creates a new read-only instance of OrbiterGatewayCCTP, bound to a specific deployed contract.
func NewOrbiterGatewayCCTPCaller(address common.Address, caller bind.ContractCaller) (*OrbiterGatewayCCTPCaller, error) {
	contract, err := bindOrbiterGatewayCCTP(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OrbiterGatewayCCTPCaller{contract: contract}, nil
}

// NewOrbiterGatewayCCTPTransactor creates a new write-only instance of OrbiterGatewayCCTP, bound to a specific deployed contract.
func NewOrbiterGatewayCCTPTransactor(address common.Address, transactor bind.ContractTransactor) (*OrbiterGatewayCCTPTransactor, error) {
	contract, err := bindOrbiterGatewayCCTP(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OrbiterGatewayCCTPTransactor{contract: contract}, nil
}

// NewOrbiterGatewayCCTPFilterer creates a new log filterer instance of OrbiterGatewayCCTP, bound to a specific deployed contract.
func NewOrbiterGatewayCCTPFilterer(address common.Address, filterer bind.ContractFilterer) (*OrbiterGatewayCCTPFilterer, error) {
	contract, err := bindOrbiterGatewayCCTP(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OrbiterGatewayCCTPFilterer{contract: contract}, nil
}

// bindOrbiterGatewayCCTP binds a generic wrapper to an already deployed contract.
func bindOrbiterGatewayCCTP(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OrbiterGatewayCCTPMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OrbiterGatewayCCTP.Contract.OrbiterGatewayCCTPCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OrbiterGatewayCCTP.Contract.OrbiterGatewayCCTPTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OrbiterGatewayCCTP.Contract.OrbiterGatewayCCTPTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OrbiterGatewayCCTP.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OrbiterGatewayCCTP.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OrbiterGatewayCCTP.Contract.contract.Transact(opts, method, params...)
}

// DESTINATIONCALLER is a free data retrieval call binding the contract method 0x62cef4b9.
//
// Solidity: function DESTINATION_CALLER() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCaller) DESTINATIONCALLER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OrbiterGatewayCCTP.contract.Call(opts, &out, "DESTINATION_CALLER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DESTINATIONCALLER is a free data retrieval call binding the contract method 0x62cef4b9.
//
// Solidity: function DESTINATION_CALLER() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) DESTINATIONCALLER() ([32]byte, error) {
	return _OrbiterGatewayCCTP.Contract.DESTINATIONCALLER(&_OrbiterGatewayCCTP.CallOpts)
}

// DESTINATIONCALLER is a free data retrieval call binding the contract method 0x62cef4b9.
//
// Solidity: function DESTINATION_CALLER() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerSession) DESTINATIONCALLER() ([32]byte, error) {
	return _OrbiterGatewayCCTP.Contract.DESTINATIONCALLER(&_OrbiterGatewayCCTP.CallOpts)
}

// DESTINATIONDOMAIN is a free data retrieval call binding the contract method 0x42c62279.
//
// Solidity: function DESTINATION_DOMAIN() view returns(uint32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCaller) DESTINATIONDOMAIN(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _OrbiterGatewayCCTP.contract.Call(opts, &out, "DESTINATION_DOMAIN")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// DESTINATIONDOMAIN is a free data retrieval call binding the contract method 0x42c62279.
//
// Solidity: function DESTINATION_DOMAIN() view returns(uint32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) DESTINATIONDOMAIN() (uint32, error) {
	return _OrbiterGatewayCCTP.Contract.DESTINATIONDOMAIN(&_OrbiterGatewayCCTP.CallOpts)
}

// DESTINATIONDOMAIN is a free data retrieval call binding the contract method 0x42c62279.
//
// Solidity: function DESTINATION_DOMAIN() view returns(uint32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerSession) DESTINATIONDOMAIN() (uint32, error) {
	return _OrbiterGatewayCCTP.Contract.DESTINATIONDOMAIN(&_OrbiterGatewayCCTP.CallOpts)
}

// MESSAGETRANSMITTER is a free data retrieval call binding the contract method 0xb6a84a5f.
//
// Solidity: function MESSAGE_TRANSMITTER() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCaller) MESSAGETRANSMITTER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OrbiterGatewayCCTP.contract.Call(opts, &out, "MESSAGE_TRANSMITTER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// MESSAGETRANSMITTER is a free data retrieval call binding the contract method 0xb6a84a5f.
//
// Solidity: function MESSAGE_TRANSMITTER() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) MESSAGETRANSMITTER() (common.Address, error) {
	return _OrbiterGatewayCCTP.Contract.MESSAGETRANSMITTER(&_OrbiterGatewayCCTP.CallOpts)
}

// MESSAGETRANSMITTER is a free data retrieval call binding the contract method 0xb6a84a5f.
//
// Solidity: function MESSAGE_TRANSMITTER() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerSession) MESSAGETRANSMITTER() (common.Address, error) {
	return _OrbiterGatewayCCTP.Contract.MESSAGETRANSMITTER(&_OrbiterGatewayCCTP.CallOpts)
}

// MINTRECIPIENT is a free data retrieval call binding the contract method 0x949b7b92.
//
// Solidity: function MINT_RECIPIENT() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCaller) MINTRECIPIENT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OrbiterGatewayCCTP.contract.Call(opts, &out, "MINT_RECIPIENT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MINTRECIPIENT is a free data retrieval call binding the contract method 0x949b7b92.
//
// Solidity: function MINT_RECIPIENT() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) MINTRECIPIENT() ([32]byte, error) {
	return _OrbiterGatewayCCTP.Contract.MINTRECIPIENT(&_OrbiterGatewayCCTP.CallOpts)
}

// MINTRECIPIENT is a free data retrieval call binding the contract method 0x949b7b92.
//
// Solidity: function MINT_RECIPIENT() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerSession) MINTRECIPIENT() ([32]byte, error) {
	return _OrbiterGatewayCCTP.Contract.MINTRECIPIENT(&_OrbiterGatewayCCTP.CallOpts)
}

// TOKEN is a free data retrieval call binding the contract method 0x82bfefc8.
//
// Solidity: function TOKEN() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCaller) TOKEN(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OrbiterGatewayCCTP.contract.Call(opts, &out, "TOKEN")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TOKEN is a free data retrieval call binding the contract method 0x82bfefc8.
//
// Solidity: function TOKEN() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) TOKEN() (common.Address, error) {
	return _OrbiterGatewayCCTP.Contract.TOKEN(&_OrbiterGatewayCCTP.CallOpts)
}

// TOKEN is a free data retrieval call binding the contract method 0x82bfefc8.
//
// Solidity: function TOKEN() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerSession) TOKEN() (common.Address, error) {
	return _OrbiterGatewayCCTP.Contract.TOKEN(&_OrbiterGatewayCCTP.CallOpts)
}

// TOKENMESSENGER is a free data retrieval call binding the contract method 0xb8b32ff7.
//
// Solidity: function TOKEN_MESSENGER() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCaller) TOKENMESSENGER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _OrbiterGatewayCCTP.contract.Call(opts, &out, "TOKEN_MESSENGER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TOKENMESSENGER is a free data retrieval call binding the contract method 0xb8b32ff7.
//
// Solidity: function TOKEN_MESSENGER() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) TOKENMESSENGER() (common.Address, error) {
	return _OrbiterGatewayCCTP.Contract.TOKENMESSENGER(&_OrbiterGatewayCCTP.CallOpts)
}

// TOKENMESSENGER is a free data retrieval call binding the contract method 0xb8b32ff7.
//
// Solidity: function TOKEN_MESSENGER() view returns(address)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerSession) TOKENMESSENGER() (common.Address, error) {
	return _OrbiterGatewayCCTP.Contract.TOKENMESSENGER(&_OrbiterGatewayCCTP.CallOpts)
}

// DestinationCaller is a free data retrieval call binding the contract method 0xb04b4f98.
//
// Solidity: function destinationCaller() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCaller) DestinationCaller(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _OrbiterGatewayCCTP.contract.Call(opts, &out, "destinationCaller")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DestinationCaller is a free data retrieval call binding the contract method 0xb04b4f98.
//
// Solidity: function destinationCaller() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) DestinationCaller() ([32]byte, error) {
	return _OrbiterGatewayCCTP.Contract.DestinationCaller(&_OrbiterGatewayCCTP.CallOpts)
}

// DestinationCaller is a free data retrieval call binding the contract method 0xb04b4f98.
//
// Solidity: function destinationCaller() view returns(bytes32)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPCallerSession) DestinationCaller() ([32]byte, error) {
	return _OrbiterGatewayCCTP.Contract.DestinationCaller(&_OrbiterGatewayCCTP.CallOpts)
}

// DepositForBurnWithOrbiter is a paid mutator transaction binding the contract method 0xecdbac94.
//
// Solidity: function depositForBurnWithOrbiter(uint256 amount, uint256 blocktimeDeadline, uint8 v, bytes32 r, bytes32 s, bytes orbiterPayload) returns()
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPTransactor) DepositForBurnWithOrbiter(opts *bind.TransactOpts, amount *big.Int, blocktimeDeadline *big.Int, v uint8, r [32]byte, s [32]byte, orbiterPayload []byte) (*types.Transaction, error) {
	return _OrbiterGatewayCCTP.contract.Transact(opts, "depositForBurnWithOrbiter", amount, blocktimeDeadline, v, r, s, orbiterPayload)
}

// DepositForBurnWithOrbiter is a paid mutator transaction binding the contract method 0xecdbac94.
//
// Solidity: function depositForBurnWithOrbiter(uint256 amount, uint256 blocktimeDeadline, uint8 v, bytes32 r, bytes32 s, bytes orbiterPayload) returns()
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPSession) DepositForBurnWithOrbiter(amount *big.Int, blocktimeDeadline *big.Int, v uint8, r [32]byte, s [32]byte, orbiterPayload []byte) (*types.Transaction, error) {
	return _OrbiterGatewayCCTP.Contract.DepositForBurnWithOrbiter(&_OrbiterGatewayCCTP.TransactOpts, amount, blocktimeDeadline, v, r, s, orbiterPayload)
}

// DepositForBurnWithOrbiter is a paid mutator transaction binding the contract method 0xecdbac94.
//
// Solidity: function depositForBurnWithOrbiter(uint256 amount, uint256 blocktimeDeadline, uint8 v, bytes32 r, bytes32 s, bytes orbiterPayload) returns()
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPTransactorSession) DepositForBurnWithOrbiter(amount *big.Int, blocktimeDeadline *big.Int, v uint8, r [32]byte, s [32]byte, orbiterPayload []byte) (*types.Transaction, error) {
	return _OrbiterGatewayCCTP.Contract.DepositForBurnWithOrbiter(&_OrbiterGatewayCCTP.TransactOpts, amount, blocktimeDeadline, v, r, s, orbiterPayload)
}

// OrbiterGatewayCCTPDepositForBurnWithOrbiterIterator is returned from FilterDepositForBurnWithOrbiter and is used to iterate over the raw logs and unpacked data for DepositForBurnWithOrbiter events raised by the OrbiterGatewayCCTP contract.
type OrbiterGatewayCCTPDepositForBurnWithOrbiterIterator struct {
	Event *OrbiterGatewayCCTPDepositForBurnWithOrbiter // Event containing the contract specifics and raw log

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
func (it *OrbiterGatewayCCTPDepositForBurnWithOrbiterIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OrbiterGatewayCCTPDepositForBurnWithOrbiter)
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
		it.Event = new(OrbiterGatewayCCTPDepositForBurnWithOrbiter)
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
func (it *OrbiterGatewayCCTPDepositForBurnWithOrbiterIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OrbiterGatewayCCTPDepositForBurnWithOrbiterIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OrbiterGatewayCCTPDepositForBurnWithOrbiter represents a DepositForBurnWithOrbiter event raised by the OrbiterGatewayCCTP contract.
type OrbiterGatewayCCTPDepositForBurnWithOrbiter struct {
	TransferNonce uint64
	PayloadNonce  uint64
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterDepositForBurnWithOrbiter is a free log retrieval operation binding the contract event 0xe8c9e52a56b4d3cf6127b4387ce50fa5b38c332d1686035eddef220ea068476b.
//
// Solidity: event DepositForBurnWithOrbiter(uint64 indexed transferNonce, uint64 indexed payloadNonce)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPFilterer) FilterDepositForBurnWithOrbiter(opts *bind.FilterOpts, transferNonce []uint64, payloadNonce []uint64) (*OrbiterGatewayCCTPDepositForBurnWithOrbiterIterator, error) {

	var transferNonceRule []interface{}
	for _, transferNonceItem := range transferNonce {
		transferNonceRule = append(transferNonceRule, transferNonceItem)
	}
	var payloadNonceRule []interface{}
	for _, payloadNonceItem := range payloadNonce {
		payloadNonceRule = append(payloadNonceRule, payloadNonceItem)
	}

	logs, sub, err := _OrbiterGatewayCCTP.contract.FilterLogs(opts, "DepositForBurnWithOrbiter", transferNonceRule, payloadNonceRule)
	if err != nil {
		return nil, err
	}
	return &OrbiterGatewayCCTPDepositForBurnWithOrbiterIterator{contract: _OrbiterGatewayCCTP.contract, event: "DepositForBurnWithOrbiter", logs: logs, sub: sub}, nil
}

// WatchDepositForBurnWithOrbiter is a free log subscription operation binding the contract event 0xe8c9e52a56b4d3cf6127b4387ce50fa5b38c332d1686035eddef220ea068476b.
//
// Solidity: event DepositForBurnWithOrbiter(uint64 indexed transferNonce, uint64 indexed payloadNonce)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPFilterer) WatchDepositForBurnWithOrbiter(opts *bind.WatchOpts, sink chan<- *OrbiterGatewayCCTPDepositForBurnWithOrbiter, transferNonce []uint64, payloadNonce []uint64) (event.Subscription, error) {

	var transferNonceRule []interface{}
	for _, transferNonceItem := range transferNonce {
		transferNonceRule = append(transferNonceRule, transferNonceItem)
	}
	var payloadNonceRule []interface{}
	for _, payloadNonceItem := range payloadNonce {
		payloadNonceRule = append(payloadNonceRule, payloadNonceItem)
	}

	logs, sub, err := _OrbiterGatewayCCTP.contract.WatchLogs(opts, "DepositForBurnWithOrbiter", transferNonceRule, payloadNonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OrbiterGatewayCCTPDepositForBurnWithOrbiter)
				if err := _OrbiterGatewayCCTP.contract.UnpackLog(event, "DepositForBurnWithOrbiter", log); err != nil {
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

// ParseDepositForBurnWithOrbiter is a log parse operation binding the contract event 0xe8c9e52a56b4d3cf6127b4387ce50fa5b38c332d1686035eddef220ea068476b.
//
// Solidity: event DepositForBurnWithOrbiter(uint64 indexed transferNonce, uint64 indexed payloadNonce)
func (_OrbiterGatewayCCTP *OrbiterGatewayCCTPFilterer) ParseDepositForBurnWithOrbiter(log types.Log) (*OrbiterGatewayCCTPDepositForBurnWithOrbiter, error) {
	event := new(OrbiterGatewayCCTPDepositForBurnWithOrbiter)
	if err := _OrbiterGatewayCCTP.contract.UnpackLog(event, "DepositForBurnWithOrbiter", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
