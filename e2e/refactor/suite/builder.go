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

package suite

import (
	"context"
	"testing"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type SuiteBuilder struct {
	t      *testing.T
	ctx    context.Context
	logger *zap.Logger
	docker Docker

	suite *Suite

	relayerReporter *testreporter.RelayerExecReporter

	chainSpecs []*interchaintest.ChainSpec
}

func NewBuilder(t *testing.T, opts ...BuilderOpt) *SuiteBuilder {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	client, network := interchaintest.DockerSetup(t)

	sb := &SuiteBuilder{
		t:      t,
		ctx:    ctx,
		logger: logger,
		docker: Docker{
			client:    client,
			networkID: network,
		},
		suite: &Suite{},

		chainSpecs: make([]*interchaintest.ChainSpec, 0),
	}

	for _, opt := range opts {
		opt(sb)
	}

	return sb
}

type BuilderOpt func(b *SuiteBuilder)

func (b *SuiteBuilder) WithOptions(...BuilderOpt)

func (b *SuiteBuilder) WithDefaultNobleSpec() {
	// create the Noble specs and append them
	b.chainSpecs = append(
		b.chainSpecs,
	)
}

func (sb *SuiteBuilder) AppendChainSpec(spec *interchaintest.ChainSpec) {
	sb.chainSpecs = append(sb.chainSpecs, spec)
}

func (b *SuiteBuilder) Build() {
	ctx := b.ctx
	suite := b.suite
	t := b.t

	factory := interchaintest.NewBuiltinChainFactory(
		b.logger,
		b.chainSpecs,
	)

	chains, err := factory.Chains(b.t.Name())
	require.NoError(b.t, err)

	interchain := interchaintest.NewInterchain()

	for _, spec := range b.chainSpecs {
		interchain.AddChain(&spec)
	}

	require.NoError(
		t,
		interchain.Build(ctx, b.relayerReporter, interchaintest.InterchainBuildOptions{
			TestName:         t.Name(),
			Client:           b.docker.client,
			NetworkID:        b.docker.networkID,
			SkipPathCreation: true,
		}),
	)

	t.Cleanup(func() {
		err := interchain.Close()
		if err != nil {
			t.Logf("failed to close interchain: %v", err)
		}
	})
}
