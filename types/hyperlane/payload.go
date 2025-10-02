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
)

// GetPayloadHashFromWarpMessageBody grabs the orbiter payload hash from a Hyperlane message body.
// This hash is stored in the last 32 bytes of the passed byte slice.
func GetPayloadHashFromWarpMessageBody(body []byte) ([]byte, error) {
	if len(body) != ORBITER_PAYLOAD_SIZE {
		return nil, fmt.Errorf(
			"malformed orbiter payload; expected %d bytes, got %d",
			ORBITER_PAYLOAD_SIZE,
			len(body),
		)
	}

	return body[len(body)-hyperlaneutil.HEX_ADDRESS_LENGTH:], nil
}

// GetReducedWarpMessageFromOrbiterMessage removes the extra payload bytes from the formatted
// message body
// which turns the custom Orbiter Hyperlane message format into a Warp compatible one.
func GetReducedWarpMessageFromOrbiterMessage(
	message hyperlaneutil.HyperlaneMessage,
) (hyperlaneutil.HyperlaneMessage, error) {
	payload := message.Body
	if len(payload) != ORBITER_PAYLOAD_SIZE {
		return hyperlaneutil.HyperlaneMessage{}, fmt.Errorf(
			"malformed orbiter payload; expected %d bytes, got %d",
			ORBITER_PAYLOAD_SIZE,
			len(payload),
		)
	}

	// NOTE: this operation just leaves the first two body entries in (recipient & amount)
	// and cuts the orbiter payload hash.
	message.Body = payload[:2*hyperlaneutil.HEX_ADDRESS_LENGTH]

	return message, nil
}
