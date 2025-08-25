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

package executor_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/keeper/component/executor"
	"github.com/noble-assets/orbiter/testutil"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	executortypes "github.com/noble-assets/orbiter/types/component/executor"
	"github.com/noble-assets/orbiter/types/core"
)

func TestMsgServerPauseAction(t *testing.T) {
	testCases := []struct {
		name   string
		msg    *executortypes.MsgPauseAction
		expErr string
	}{
		{
			name: "error - unauthorized signer",
			msg: &executortypes.MsgPauseAction{
				Signer:   "noble1invalid",
				ActionId: core.ACTION_FEE.String(),
			},
			expErr: core.ErrUnauthorized.Error(),
		},
		{
			name: "error - invalid action ID",
			msg: &executortypes.MsgPauseAction{
				Signer:   testutil.Authority,
				ActionId: core.ActionID(99).String(),
			},
			expErr: "invalid action ID",
		},
		{
			name: "success - valid pause request",
			msg: &executortypes.MsgPauseAction{
				Signer:   testutil.Authority,
				ActionId: core.ACTION_FEE.String(),
			},
			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			msgServer := executor.NewMsgServer(k.Executor(), k)

			resp, err := msgServer.PauseAction(ctx, tC.msg)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				events := ctx.EventManager().Events()
				require.Len(t, events, 1)
				require.Contains(t, events[0].Type, "EventPaused")
			}
		})
	}
}

func TestMsgServerUnpauseAction(t *testing.T) {
	testCases := []struct {
		name   string
		msg    *executortypes.MsgUnpauseAction
		expErr string
	}{
		{
			name: "error - unauthorized signer",
			msg: &executortypes.MsgUnpauseAction{
				Signer:   "noble1invalid",
				ActionId: core.ACTION_FEE.String(),
			},
			expErr: core.ErrUnauthorized.Error(),
		},
		{
			name: "error - invalid action ID",
			msg: &executortypes.MsgUnpauseAction{
				Signer:   testutil.Authority,
				ActionId: core.ActionID(99).String(),
			},
			expErr: "invalid action ID",
		},
		{
			name: "success - valid unpause request",
			msg: &executortypes.MsgUnpauseAction{
				Signer:   testutil.Authority,
				ActionId: core.ACTION_FEE.String(),
			},
			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			msgServer := executor.NewMsgServer(k.Executor(), k)

			resp, err := msgServer.UnpauseAction(ctx, tC.msg)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// ASSERT: should only contain the unpause event
				events := ctx.EventManager().Events()
				require.Len(t, events, 1)
				// TODO: since the messages etc. are also going to be developer / user facing --
				// should this rather be `Pause` and `Resume`? Or `Disable`/`Enable`? Would be
				// ebtter english..
				require.Contains(t, events[0].Type, "EventUnpaused")
			}
		})
	}
}
