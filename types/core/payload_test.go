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

package core_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/testutil"
	"github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

func TestPayloadRoundTrip(t *testing.T) {
	pp := createPendingPayload(t)

	hash, err := pp.SHA256Hash()
	require.NoError(t, err, "failed to hash payload")

	parsedHash, err := core.NewPayloadHash(hash.String())
	require.NoError(t, err, "failed to parse payload hash")

	require.Equal(t, hash.String(), parsedHash.String(), "payload hash mismatch")
}

func TestNewPayloadHash(t *testing.T) {
	t.Parallel()

	pp := createPendingPayload(t)

	hash, err := pp.SHA256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:  "success - valid hash",
			input: hash.String(),
		},
		{
			name:        "error - invalid hash",
			input:       "abcdefg",
			errContains: "invalid payload hash",
		},
		{
			name:        "error - too short hash",
			input:       "0123ab",
			errContains: "malformed payload hash",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := core.NewPayloadHash(tc.input)
			if tc.errContains == "" {
				require.NoError(t, err, "failed to parse payload hash")
				require.Equal(t, tc.input, parsed.String(), "payload hash mismatch")
			} else {
				require.ErrorContains(t, err, tc.errContains)
			}
		})
	}
}

func createPendingPayload(t *testing.T) *core.PendingPayload {
	t.Helper()

	fw, err := forwarding.NewCCTPForwarding(
		1,
		testutil.RandomBytes(32),
		nil,
		nil,
	)
	require.NoError(t, err, "failed to create forwarding")

	p, err := core.NewPayload(fw)
	require.NoError(t, err, "failed to create payload")

	return &core.PendingPayload{
		Sequence: 0,
		Payload:  p,
	}
}
