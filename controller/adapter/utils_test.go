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

package adapter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/controller/adapter"
)

func TestRecoverNativeDenom(t *testing.T) {
	testCases := []struct {
		name          string
		denom         string
		sourcePort    string
		sourceChannel string
		expErr        string
		expDenom      string
	}{
		{
			name:          "error - non native denom (multi-hop)",
			denom:         "transfer/channel-2/transfer/channel-3/uusdc",
			sourcePort:    "transfer",
			sourceChannel: "channel-1",
			expErr:        "coin is native of source chain",
		},
		{
			name:          "error - non native denom (multi-hop)",
			denom:         "transfer/channel-1/transfer/channel-2/uusdc",
			sourcePort:    "transfer",
			sourceChannel: "channel-1",
			expErr:        "orbiter supports only native coins",
		},
		{
			name:          "success - native denom",
			denom:         "transfer/channel-1/uusdc",
			sourcePort:    "transfer",
			sourceChannel: "channel-1",
			expDenom:      "uusdc",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			denom, err := adapter.RecoverNativeDenom(tC.denom, tC.sourcePort, tC.sourceChannel)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.expDenom, denom)
			}
		})
	}
}
