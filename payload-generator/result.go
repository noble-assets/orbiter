package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) writeFinalPayload(s *strings.Builder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Generated Orbiter Payload"))
	s.WriteString("\n\n")
	s.WriteString("Your payload has been successfully generated! This JSON can be used as an IBC memo\n")
	s.WriteString("or sent to the Orbiter module to execute cross-chain operations.\n\n")
	s.WriteString("Preview (truncated for display):\n")
	s.WriteString(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).Render(m.payload[:min(len(m.payload), 200)] + "..."))
	s.WriteString("\n\n💡 The full payload will be printed to your terminal when you exit")
	s.WriteString("\n\nPress Enter or Ctrl+C to exit")
}
