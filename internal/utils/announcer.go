package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-wordwrap"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

var (
	termWidth int
	decrement int = 15

	// adaptiveHeaderLowerBorderColor helps facilitate light-colored
	// terminal displays. As such header underline color must be
	// adaptive to terminal color for visibility.
	adaptiveHeaderLowerBorderColor lipgloss.AdaptiveColor
)

func init() {
	termWidth, _, _ = term.GetSize(int(os.Stdout.Fd()))
	adaptiveHeaderLowerBorderColor = lipgloss.AdaptiveColor{Light: "#3d3d40", Dark: "#FAFAFA"}
}

func ContentsPrinter(div string, contents ...string) {
	if len(contents) == 0 {
		return
	}

	contentsWithFormatting := []string{}

	for _, c := range contents {
		contentsWithFormatting = append(contentsWithFormatting,
			fmt.Sprintf("%s\n%s", div, c),
		)
	}

	total := lipgloss.JoinVertical(lipgloss.Left, contentsWithFormatting...)

	fmt.Println(total)
}

func Error(header string, err error) {
	description := err.Error()

	wrappedDesc := wordwrap.WrapString(description, uint(termWidth-decrement))

	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true).
		PaddingRight(4).
		BorderForeground(adaptiveHeaderLowerBorderColor).
		Render

	total := lipgloss.JoinVertical(lipgloss.Left, headerStyle(header), wrappedDesc)

	var style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(2).
		PaddingRight(2).
		Width(termWidth-decrement+5).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#EE4B2B"))

	fmt.Println(style.Render(total))
}

func Warning(header, message string) {
	wrappedDesc := wordwrap.WrapString(message, uint(termWidth-decrement))

	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true).
		PaddingRight(4).
		BorderForeground(adaptiveHeaderLowerBorderColor).
		Render

	total := lipgloss.JoinVertical(lipgloss.Left, headerStyle(header), wrappedDesc)

	var style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(2).
		PaddingRight(2).
		Width(termWidth-decrement+5).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#FBFF00"))

	fmt.Println(style.Render(total))
}

func Tell(header, message string) {
	wrappedDesc := wordwrap.WrapString(message, uint(termWidth-decrement))

	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true).
		PaddingRight(4).
		BorderForeground(adaptiveHeaderLowerBorderColor).
		Render

	total := lipgloss.JoinVertical(lipgloss.Left, headerStyle(header), wrappedDesc)

	var style = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(2).
		PaddingRight(2).
		Width(termWidth-decrement+5).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("#B2BEB5"))

	fmt.Println(style.Render(total))
}

func PrintTable(header []string, data [][]string) {
	sb := &strings.Builder{}

	table := tablewriter.NewWriter(sb)

	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.SetBorder(false)

	table.AppendBulk(data)
	table.Render()

	fmt.Println(sb.String())
}

func PrintTableV2(header []string, data [][]string, caption string) {
	sb := &strings.Builder{}

	table := tablewriter.NewWriter(sb)

	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)
	table.SetBorder(false)
	table.SetCaption(true, caption)

	table.AppendBulk(data)
	table.Render()

	fmt.Println(sb.String())
}
