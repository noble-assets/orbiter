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

package core

import (
	"fmt"

	"cosmossdk.io/collections"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const ModuleName = "orbiter"

const (
	ComponentPrefix = "component"
)

var (
	ModuleAddress = authtypes.NewModuleAddress(ModuleName)

	dustCollectorName    = fmt.Sprintf("%s/%s", ModuleName, "dust_collector")
	DustCollectorAddress = authtypes.NewModuleAddress(dustCollectorName)
)

// ====================================================================================================
// Forwarding
// ====================================================================================================.
const (
	ForwarderName            = "forwarder"
	ForwardingControllerName = "forwarding_controller"

	// Maps names.
	PausedProtocolCounterpartiesName = "paused_protocol_counterparties"
	PausedProtocolsName              = "paused_protocols"
)

var (
	PausedProtocolCounterpartiesPrefix = collections.NewPrefix(10)
	PausedProtocolsPrefix              = collections.NewPrefix(11)
)

// ====================================================================================================
// Action
// ====================================================================================================.
const (
	ExecutorName         = "executor"
	ActionControllerName = "action_controller"

	// Maps names.
	PausedActionsName = "paused_action_controllers"

	// Controllers constants.

	// BPSNormalizer is used to normalize the basis points
	// defined in a fee action execution.
	BPSNormalizer = 10_000
)

var PausedActionsPrefix = collections.NewPrefix(20)

// ====================================================================================================
// Dispatcher
// ====================================================================================================.
const (
	DispatcherName = "dispatcher"

	// Maps names.
	DispatchedAmountsName = "dispatched_amounts"
	DispatchedCountsName  = "dispatched_counts"
)

var (
	DispatchedAmountsPrefix                        = collections.NewPrefix(30)
	DispatchedAmountsPrefixByDestinationProtocolID = collections.NewPrefix(31)
	DispatchedAmountsPrefixByDestinationCCID       = collections.NewPrefix(32)

	DispatchedCountsPrefix                        = collections.NewPrefix(33)
	DispatchedCountsPrefixByDestinationProtocolID = collections.NewPrefix(34)
)

// ====================================================================================================
// Adapter
// ====================================================================================================.
const (
	AdapterName           = "adapter"
	AdapterControllerName = "adapter_controller"

	AdapterParamsName = "adapter_params"

	// CCTPNobleDomain is the identifier of the Noble domain
	// in the CCTP protocol.
	CCTPNobleDomain = 4

	crosschainIDSeparator = ":"
)

var OrbiterPrefix = ModuleName

var AdapterParamsPrefix = collections.NewPrefix(40)
