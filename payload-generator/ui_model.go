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

package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"orbiter.dev/types/core"
)

// state is a toggle for the currently selected UI state.
type state int

const (
	actionSelection state = iota
	feeActionInput
	forwardingSelection
	cctpForwardingInput
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

var focusIndex int

// model contains all relevant information and state
// for the UI to interactively build an Orbiter payload.
type model struct {
	state state
	list  list.Model

	actionInputs     []textinput.Model
	forwardingInputs []textinput.Model

	actions    []*core.Action
	forwarding *core.Forwarding
	err        error
	payload    string

	windowWidth  int
	windowHeight int
}

// initialModel creates the default view for the payload generator,
// that is shown when starting the tool.
func initialModel() model {
	actionItems := []list.Item{
		item{title: core.ACTION_FEE.String(), desc: "Add fee payment action"},
		item{title: core.ACTION_SWAP.String(), desc: "Add token swap action"},
		item{title: "No more actions", desc: "Proceed to forwarding selection"},
	}

	l := list.New(actionItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select an action to add:"

	return model{
		state:   actionSelection,
		list:    l,
		actions: []*core.Action{},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles the different TUI states through the different
// selection modals.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 3)
		return m, nil
	}

	var cmd tea.Cmd
	switch m.state {
	case actionSelection, forwardingSelection:
		m.list, cmd = m.list.Update(msg)
	case feeActionInput:
		cmd = m.updateActionInputs(msg)
	case cctpForwardingInput:
		cmd = m.updateForwardingInputs(msg)
	default:
		panic(fmt.Errorf("unhandled state: %v", m.state))
	}

	return m, cmd
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case actionSelection:
		selected := m.list.SelectedItem().(item)
		switch selected.title {
		case core.ACTION_FEE.String():
			return m.initFeeActionInput(), nil
		case core.ACTION_SWAP.String():
			panic("not implemented yet: " + core.ACTION_SWAP.String())
		case "No more actions":
			return m.initForwardingSelection(), nil
		}
	case feeActionInput:
		return m.processFeeAction()
	case forwardingSelection:
		selected := m.list.SelectedItem().(item)
		switch selected.title {
		case core.PROTOCOL_CCTP.String():
			return m.initCCTPForwardingInput(), nil
		case core.PROTOCOL_IBC.String():
			panic("not supported yet: " + core.PROTOCOL_IBC.String())
		case core.PROTOCOL_HYPERLANE.String():
			panic("not implemented yet: " + core.PROTOCOL_HYPERLANE.String())
		}
	case cctpForwardingInput:
		return m.processCCTPForwarding()
	}
	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	switch m.state {
	case actionSelection:
		m.writeActionSelection(&s)
	case forwardingSelection:
		m.writeForwardingSelection(&s)
	case feeActionInput:
		m.writeFeeActionSelection(&s)
	case cctpForwardingInput:
		m.writeCCTPForwardingSelection(&s)
	}

	if m.err != nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("\nError: " + m.err.Error()))
	}

	return s.String()
}
