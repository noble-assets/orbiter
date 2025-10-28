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

package adapter

import (
	"errors"
	"fmt"
	"strings"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

func RecoverNativeDenom(denom, sourcePort, sourceChannel string) (string, error) {
	if transfertypes.SenderChainIsSource(sourcePort, sourceChannel, denom) {
		return "", errors.New("coin is native of source chain")
	}

	voucherPrefix := transfertypes.GetDenomPrefix(sourcePort, sourceChannel)

	// Remove from the denom the prefix created on the source chain when it received
	// the coin from Noble.
	if !strings.HasPrefix(denom, voucherPrefix) {
		return "", fmt.Errorf(
			"denom %q missing expected IBC prefix %q",
			denom,
			voucherPrefix,
		)
	}
	unprefixedDenom := strings.TrimPrefix(denom, voucherPrefix)

	// The denomination used to send the coins is either the native denom or the hash of the path
	// if the denomination is not native.
	denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
	if !denomTrace.IsNativeDenom() {
		return "", errors.New("orbiter supports only native coins")
	}

	return unprefixedDenom, nil
}
