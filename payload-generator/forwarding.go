package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"orbiter.dev/testutil"
	"orbiter.dev/types/controller/forwarding"
	"orbiter.dev/types/core"
)

func (m model) writeForwardingSelection(s *strings.Builder) {
	// Header
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Select Forwarding Protocol"))
	s.WriteString("\n\n")

	// Explanation
	s.WriteString("Now choose how to forward your transaction to the destination chain.\n")
	s.WriteString("Each protocol supports different chains and tokens:\n\n")

	// List
	s.WriteString(m.list.View())
}

func (m model) writeCCTPForwardingSelection(s *strings.Builder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Configure CCTP Forwarding"))
	s.WriteString("\n\n")
	s.WriteString("CCTP enables USDC transfers across chains. Configure the destination details:\n")
	s.WriteString("• Domain: Chain identifier (0=Ethereum, 1=Avalanche, 2=OP, 3=Arbitrum, 7=Base)\n")
	s.WriteString("• Mint Recipient: Address that receives USDC on destination\n")
	s.WriteString("• Destination Caller: Address that can call functions on destination\n")
	s.WriteString("• Passthrough Payload: Additional data to pass through (optional)\n\n")

	for _, input := range m.forwardingInputs {
		s.WriteString(input.View() + "\n")
	}

	s.WriteString("\nUse Tab/Shift+Tab to navigate fields, Enter to create payload, Ctrl+C to quit")
}

func (m model) initForwardingSelection() model {
	forwardingItems := []list.Item{
		item{title: core.PROTOCOL_CCTP.String(), desc: "Circle's Cross-Chain Transfer Protocol (USDC transfers)"},
		item{title: core.PROTOCOL_IBC.String(), desc: "Inter-Blockchain Communication (Cosmos ecosystem)"},
		item{title: core.PROTOCOL_HYPERLANE.String(), desc: "Hyperlane interchain protocol"},
	}

	l := list.New(forwardingItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a protocol:"

	// Apply stored window dimensions if we have them
	if m.windowWidth > 0 && m.windowHeight > 0 {
		l.SetWidth(m.windowWidth)
		l.SetHeight(m.windowHeight - 3)
	}

	m.list = l
	m.state = forwardingSelection

	return m
}

func (m model) initCCTPForwardingInput() model {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Destination domain (e.g. 0)"
	inputs[0].CharLimit = 10
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Mint recipient (put 'r' for random)"
	inputs[1].CharLimit = 128
	inputs[1].Width = 70

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Destination caller (put 'r' for random)"
	inputs[2].CharLimit = 128
	inputs[2].Width = 70

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Passthrough payload (can be left empty)"
	inputs[3].CharLimit = 256
	inputs[3].Width = 70

	m.forwardingInputs = inputs
	forwardingFocusIndex = 0
	m.state = cctpForwardingInput

	// Focus the first input
	m.forwardingInputs[0].Focus()

	return m
}

func (m model) processCCTPForwarding() (tea.Model, tea.Cmd) {
	domainStr := m.forwardingInputs[0].Value()
	mintRecipientStr := m.forwardingInputs[1].Value()
	destCallerStr := m.forwardingInputs[2].Value()
	passthroughStr := m.forwardingInputs[3].Value()

	if domainStr == "" {
		return m, nil
	}

	domain, err := strconv.ParseUint(domainStr, 10, 32)
	if err != nil {
		m.err = fmt.Errorf("invalid destination domain: %v", err)
		return m, nil
	}

	var mintRecipient []byte
	if mintRecipientStr == "" {
		m.err = errors.New("mint recipient cannot be empty")
	} else if mintRecipientStr == "r" {
		mintRecipient = testutil.RandomBytes(32)
	} else {
		mintRecipient = []byte(mintRecipientStr)
	}

	var destCaller []byte
	if destCallerStr == "r" {
		destCaller = testutil.RandomBytes(32)
	} else if destCallerStr != "" {
		destCaller = []byte(destCallerStr)
	}

	var passthroughPayload []byte
	if passthroughStr != "" {
		passthroughPayload = []byte(passthroughStr)
	}

	cctpForwarding, err := forwarding.NewCCTPForwarding(
		uint32(domain),
		mintRecipient,
		destCaller,
		passthroughPayload,
	)
	if err != nil {
		m.err = fmt.Errorf("failed to create CCTP forwarding: %v", err)
		return m, nil
	}

	m.forwarding = cctpForwarding
	return m.buildFinalPayload()
}

func (m model) updateForwardingInputs(msg tea.Msg) tea.Cmd {
	if len(m.forwardingInputs) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Update focus position
			if (s == "up" || s == "shift+tab") && forwardingFocusIndex > 0 {
				forwardingFocusIndex--
			} else if forwardingFocusIndex < len(m.forwardingInputs)-1 {
				forwardingFocusIndex++
			}

			// Update focus for all inputs
			cmds := make([]tea.Cmd, len(m.forwardingInputs))
			for i := 0; i < len(m.forwardingInputs); i++ {
				if i == forwardingFocusIndex {
					cmds[i] = m.forwardingInputs[i].Focus()
				} else {
					m.forwardingInputs[i].Blur()
				}
			}

			return tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking for all inputs
	cmds := make([]tea.Cmd, len(m.forwardingInputs))
	for i := range m.forwardingInputs {
		m.forwardingInputs[i], cmds[i] = m.forwardingInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
