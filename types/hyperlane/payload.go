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

package hyperlane

import (
	"fmt"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"

	errorsmod "cosmossdk.io/errors"

	"github.com/noble-assets/orbiter/types/core"
)

// GetPayloadFromWarpMessageBody parses the Orbiter payload from a Hyperlane message body.
// This hash is stored after the default Warp payload contents and consists of the Proto-marshalled
// data.
func GetPayloadFromWarpMessageBody(body []byte) (*core.Payload, error) {
	if len(body) <= 64 {
		return nil, fmt.Errorf(
			"malformed orbiter payload; expected more than 64 bytes; got: %d",
			len(body),
		)
	}

	var payload core.Payload
	if err := payload.Unmarshal(body[64:]); err != nil {
		return nil, errorsmod.Wrap(err, "failed to unmarshal payload")
	}

	return &payload, nil
}

// GetReducedWarpMessageFromOrbiterMessage removes the extra payload bytes from the formatted
// message body
// which turns the custom Orbiter Hyperlane message format into a Warp compatible one.
//
// TODO: move into hyperlane adapter types
func GetReducedWarpMessageFromOrbiterMessage(
	message hyperlaneutil.HyperlaneMessage,
) (hyperlaneutil.HyperlaneMessage, error) {
	payload := message.Body
	if len(payload) <= 64 {
		return hyperlaneutil.HyperlaneMessage{}, fmt.Errorf(
			"malformed orbiter payload; expected more than 64 bytes, got %d",
			len(payload),
		)
	}

	// NOTE: this operation just leaves the first two body entries in (recipient & amount)
	// and cuts the orbiter payload hash.
	message.Body = payload[:64]

	return message, nil
}
