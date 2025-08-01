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

package adapters

import (
	"encoding/json"
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/types"
)

// JSONParser is an utility type capable of parsing
// a JSON representation of the orbiter payload into
// the data transfer type.
type JSONParser struct {
	cdc codec.Codec
}

// NewJSONParser returns a reference to a JSONParser instance.
func NewJSONParser(cdc codec.Codec) (*JSONParser, error) {
	if cdc == nil {
		return nil, errors.New("codec cannot be nil for JSON parser")
	}

	return &JSONParser{
		cdc: cdc,
	}, nil
}

// Parse returns the orbiter payload from a JSON formatted
// string or an error.
func (p *JSONParser) Parse(jsonString string) (*types.Payload, error) {
	var jsonData map[string]any
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		return nil, types.ErrParsingPayload.Wrapf("not a valid json string: %s", err.Error())
	}

	if len(jsonData) != 1 {
		return nil, types.ErrParsingPayload.Wrapf(
			"json data contains multiple root level keys, accepted only %s",
			types.OrbiterPrefix,
		)
	}

	if jsonData[types.OrbiterPrefix] == nil {
		return nil, types.ErrParsingPayload.Wrapf(
			"json does not contain orbiter prefix: %s",
			types.OrbiterPrefix,
		)
	}

	pw := types.PayloadWrapper{}
	err = types.UnmarshalJSON(p.cdc, []byte(jsonString), &pw)
	if err != nil {
		return nil, types.ErrParsingPayload.Wrapf(
			"failed to cast json string into Payload: %s",
			err.Error(),
		)
	}

	return pw.Orbiter, nil
}
