package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"orbiter.dev/types/controller/action"
	"orbiter.dev/types/core"
	"strconv"
	"strings"
)

func (m model) writeActionSelection(s *strings.Builder) {
	// Header
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Orbiter Payload Generator"))
	s.WriteString("\n\n")

	// Explanation
	if len(m.actions) == 0 {
		s.WriteString("\nWelcome! This tool helps you build payloads for cross-chain operations.\n")
		s.WriteString("Actions are optional operations that run before forwarding (like fee payments).\n")
		s.WriteString("The selected actions will be run sequentially, so bear that in mind.\n\n")
	} else {
		s.WriteString("Add another action or continue to forwarding selection.\n")
		s.WriteString("Current actions: ")
		for i, act := range m.actions {
			if i > 0 {
				s.WriteString(", ")
			}
			s.WriteString(act.Id.String())
		}
		s.WriteString("\n\n")
	}

	// List
	s.WriteString(m.list.View())
}

func (m model) writeFeeActionSelection(s *strings.Builder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Configure Fee Action"))
	s.WriteString("\n\n")
	s.WriteString("Fee actions allow you to collect a percentage of the transaction amount.\n")
	s.WriteString("The recipient will receive the specified percentage as a fee.\n\n")

	for _, input := range m.actionInputs {
		s.WriteString(input.View() + "\n")
	}

	s.WriteString("\nUse Tab/Shift+Tab to navigate fields, Enter to add action, Ctrl+C to quit")
}

func (m model) initFeeActionInput() model {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Fee recipient address"
	inputs[0].CharLimit = 100
	inputs[0].Width = 50

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Basis points (e.g. 100 for 1%)"
	inputs[1].CharLimit = 5
	inputs[1].Width = 30

	m.actionInputs = inputs
	actionFocusIndex = 0
	m.state = feeActionInput

	// Focus the first input
	m.actionInputs[0].Focus()

	return m
}

func (m model) processFeeAction() (tea.Model, tea.Cmd) {
	recipientAddr := m.actionInputs[0].Value()
	basisPointsStr := m.actionInputs[1].Value()

	if recipientAddr == "" {
		m.err = fmt.Errorf("recipient address is required")
		return m, nil
	}
	if basisPointsStr == "" {
		m.err = fmt.Errorf("basis points is required")
		return m, nil
	}

	basisPoints, err := strconv.ParseUint(basisPointsStr, 10, 32)
	if err != nil {
		m.err = fmt.Errorf("invalid basis points: %v", err)
		return m, nil
	}

	feeAttr := action.FeeAttributes{
		FeesInfo: []*action.FeeInfo{
			{
				Recipient:   recipientAddr,
				BasisPoints: uint32(basisPoints),
			},
		},
	}

	feeAction := core.Action{
		Id: core.ACTION_FEE,
	}
	err = feeAction.SetAttributes(&feeAttr)
	if err != nil {
		m.err = fmt.Errorf("failed to set action attributes: %v", err)
		return m, nil
	}

	if err = feeAction.Validate(); err != nil {
		m.err = fmt.Errorf("invalid fee action: %w", err)
	}

	m.actions = append(m.actions, feeAction)
	return m.initActionSelection(), nil
}

func (m model) initActionSelection() model {
	actionItems := []list.Item{
		item{title: core.ACTION_FEE.String(), desc: "Add fee payment action"},
		item{title: core.ACTION_SWAP.String(), desc: "Add token swap action"},
		item{title: "No more actions", desc: "Proceed to forwarding selection"},
	}

	l := list.New(actionItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select an action to add:"

	// Apply stored window dimensions if we have them
	if m.windowWidth > 0 && m.windowHeight > 0 {
		l.SetWidth(m.windowWidth)
		l.SetHeight(m.windowHeight - 3)
	}

	m.list = l
	m.state = actionSelection

	return m
}

func (m model) updateActionInputs(msg tea.Msg) tea.Cmd {
	if len(m.actionInputs) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Update focus position
			if s == "up" || s == "shift+tab" {
				if actionFocusIndex > 0 {
					actionFocusIndex--
				}
			} else {
				if actionFocusIndex < len(m.actionInputs)-1 {
					actionFocusIndex++
				}
			}

			// Update focus for all inputs
			cmds := make([]tea.Cmd, len(m.actionInputs))
			for i := 0; i < len(m.actionInputs); i++ {
				if i == actionFocusIndex {
					cmds[i] = m.actionInputs[i].Focus()
				} else {
					m.actionInputs[i].Blur()
				}
			}

			return tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking for all inputs
	cmds := make([]tea.Cmd, len(m.actionInputs))
	for i := range m.actionInputs {
		m.actionInputs[i], cmds[i] = m.actionInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
