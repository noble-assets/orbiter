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
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/noble-assets/orbiter/v2/types"
	"github.com/noble-assets/orbiter/v2/types/core"
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
		return nil, core.ErrNilPointer.Wrap("codec cannot be nil for JSON parser")
	}

	return &JSONParser{
		cdc: cdc,
	}, nil
}

// Parse returns the orbiter payload from a JSON formatted
// string or an error.
func (p *JSONParser) Parse(jsonString string) (*core.Payload, error) {
	var jsonData map[string]any
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		return nil, core.ErrParsingPayload.Wrapf("not a valid json string: %s", err.Error())
	}

	if len(jsonData) != 1 {
		return nil, core.ErrParsingPayload.Wrapf(
			"json data contains multiple root level keys, accepted only %s",
			core.OrbiterPrefix,
		)
	}

	if jsonData[core.OrbiterPrefix] == nil {
		return nil, core.ErrParsingPayload.Wrapf(
			"json does not contain orbiter prefix: %s",
			core.OrbiterPrefix,
		)
	}

	pw := core.PayloadWrapper{}
	err = types.UnmarshalJSON(p.cdc, []byte(jsonString), &pw)
	if err != nil {
		return nil, core.ErrParsingPayload.Wrapf(
			"failed to cast json string into Payload: %s",
			err.Error(),
		)
	}

	return pw.Orbiter, nil
}
