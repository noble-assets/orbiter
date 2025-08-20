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

	orbiter "orbiter.dev"
	"orbiter.dev/testutil"
	"orbiter.dev/types"
	"orbiter.dev/types/controller/action"
	"orbiter.dev/types/controller/forwarding"
	"orbiter.dev/types/core"
)

// state is a toggle for the currently selected UI state.
type state int

const (
	actionSelection state = iota
	feeActionInput
	forwardingSelection
	cctpForwardingInput
	finalPayload
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// model contains all relevant information and state
// for the UI to interactively build an Orbiter payload.
type model struct {
	state      state
	list       list.Model
	inputs     []textinput.Model
	focusIndex int

	actions    []core.Action
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
	l.Title = "Orbiter Payload Generator\n\n" +
		"Welcome! This tool helps you build payloads for cross-chain operations.\n" +
		"Actions are optional operations that run before forwarding (like fee payments).\n" +
		"The selected actions will be run sequentially, so bear that in mind.\n\n" +
		"Select an action to add:"

	return model{
		state:   actionSelection,
		list:    l,
		actions: []core.Action{},
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
	case feeActionInput, cctpForwardingInput:
		cmd = m.updateInputs(msg)
	case finalPayload:
		// no action required, will quit the program after
	default:
		panic(fmt.Errorf("unhandled state: %v", m.state))
	}

	return m, cmd
}

func (m model) updateInputs(msg tea.Msg) tea.Cmd {
	if len(m.inputs) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Update focus position
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			// Update focus for all inputs
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking for all inputs
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case actionSelection:
		selected := m.list.SelectedItem().(item)
		switch selected.title {
		case core.ACTION_FEE.String():
			return m.initFeeActionInput(), nil
		case core.ACTION_SWAP.String():
			panic(fmt.Errorf("%s is not implemented yet", core.ACTION_SWAP))
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
			panic(fmt.Errorf("%s is not implemented yet", core.PROTOCOL_IBC))
		case core.PROTOCOL_HYPERLANE.String():
			panic(fmt.Errorf("%s is not implemented yet", core.PROTOCOL_HYPERLANE))
		}
	case cctpForwardingInput:
		return m.processCCTPForwarding()
	case finalPayload:
		return m, tea.Quit
	}
	return m, nil
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

	m.inputs = inputs
	m.focusIndex = 0
	m.state = feeActionInput

	// Focus the first input
	m.inputs[0].Focus()

	return m
}

func (m model) processFeeAction() (tea.Model, tea.Cmd) {
	recipientAddr := m.inputs[0].Value()
	basisPointsStr := m.inputs[1].Value()

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
	l.Title = "Add another action or continue to forwarding:\n\nSelect an action to add:"

	// Apply stored window dimensions if we have them
	if m.windowWidth > 0 && m.windowHeight > 0 {
		l.SetWidth(m.windowWidth)
		l.SetHeight(m.windowHeight - 3)
	}

	m.list = l
	m.state = actionSelection
	m.inputs = nil
	m.focusIndex = 0

	return m
}

func (m model) initForwardingSelection() model {
	forwardingItems := []list.Item{
		item{title: core.PROTOCOL_CCTP.String(), desc: "Circle's Cross-Chain Transfer Protocol (USDC transfers)"},
		item{title: core.PROTOCOL_IBC.String(), desc: "Inter-Blockchain Communication (Cosmos ecosystem)"},
		item{title: core.PROTOCOL_HYPERLANE.String(), desc: "Hyperlane interchain protocol"},
	}

	l := list.New(forwardingItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Forwarding Protocol\n\n" +
		"Now choose how to forward your transaction to the destination chain.\n" +
		"Each protocol supports different chains and tokens:\n\n" +
		"Select a protocol:"

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

	m.inputs = inputs
	m.focusIndex = 0
	m.state = cctpForwardingInput

	// Focus the first input
	m.inputs[0].Focus()

	return m
}

func (m model) processCCTPForwarding() (tea.Model, tea.Cmd) {
	domainStr := m.inputs[0].Value()
	mintRecipientStr := m.inputs[1].Value()
	destCallerStr := m.inputs[2].Value()
	passthroughStr := m.inputs[3].Value()

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

func (m model) buildFinalPayload() (tea.Model, tea.Cmd) {
	var actions []*core.Action
	for i := range m.actions {
		actions = append(actions, &m.actions[i])
	}

	payload, err := core.NewPayloadWrapper(m.forwarding, actions)
	if err != nil {
		m.err = fmt.Errorf("failed to create payload wrapper: %v", err)
		return m, nil
	}

	encCfg := testutil.MakeTestEncodingConfig("noble")
	orbiter.RegisterInterfaces(encCfg.InterfaceRegistry)
	payloadStr, err := types.MarshalJSON(encCfg.Codec, payload)
	if err != nil {
		m.err = fmt.Errorf("failed to marshal payload: %v", err)
		return m, nil
	}

	m.payload = string(payloadStr)
	m.state = finalPayload

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	switch m.state {
	case actionSelection, forwardingSelection:
		s.WriteString(m.list.View())
	case feeActionInput:
		s.WriteString("💰 Configure Fee Action\n\n")
		s.WriteString("Fee actions allow you to collect a percentage of the transaction amount.\n")
		s.WriteString("The recipient will receive the specified percentage as a fee.\n\n")
		for i, input := range m.inputs {
			if i == m.focusIndex {
				s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("▶ "))
			} else {
				s.WriteString("  ")
			}
			s.WriteString(input.View() + "\n")
		}
		s.WriteString("\nUse Tab/Shift+Tab to navigate fields, Enter to add action, Ctrl+C to quit")
	case cctpForwardingInput:
		s.WriteString("🔗 Configure CCTP Forwarding\n\n")
		s.WriteString("CCTP enables USDC transfers across chains. Configure the destination details:\n")
		s.WriteString("• Domain: Chain identifier (0=Ethereum, 1=Avalanche, 2=OP, 3=Arbitrum, 7=Base)\n")
		s.WriteString("• Mint Recipient: Address that receives USDC on destination\n")
		s.WriteString("• Destination Caller: Address that can call functions on destination\n")
		s.WriteString("• Passthrough Payload: Additional data to pass through (optional)\n\n")
		for i, input := range m.inputs {
			if i == m.focusIndex {
				s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("▶ "))
			} else {
				s.WriteString("  ")
			}
			s.WriteString(input.View() + "\n")
		}
		s.WriteString("\nUse Tab/Shift+Tab to navigate fields, Enter to create payload, Ctrl+C to quit")
	case finalPayload:
		s.WriteString("✅ Generated Orbiter Payload\n\n")
		s.WriteString("Your payload has been successfully generated! This JSON can be used as an IBC memo\n")
		s.WriteString("or sent to the Orbiter module to execute cross-chain operations.\n\n")
		s.WriteString("Preview (truncated for display):\n")
		s.WriteString(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Render(m.payload[:min(len(m.payload), 200)] + "..."))
		s.WriteString("\n\n💡 The full payload will be printed to your terminal when you exit")
		s.WriteString("\n\nPress Enter or Ctrl+C to exit")
	}

	if m.err != nil {
		s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("\nError: " + m.err.Error()))
	}

	return s.String()
}
