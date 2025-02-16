package formcreator

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func themeFP() *huh.Theme {
	t := huh.ThemeBase()

	var (
		normalFg = lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
		fpBlue   = lipgloss.AdaptiveColor{Light: "#002F6C", Dark: "#0076A8"}
		cream    = lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}
		fpGreen  = lipgloss.AdaptiveColor{Light: "#FE9E1B", Dark: "#D57800"}
		fpOrange = lipgloss.AdaptiveColor{Light: "#658D1B", Dark: "#A8AD00"}
		fpRed    = lipgloss.AdaptiveColor{Light: "#DA291C", Dark: "#DA291C"}
	)

	t.Focused.Title = t.Focused.Title.Foreground(fpBlue).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(fpBlue).Bold(true).MarginBottom(1)
	t.Focused.Directory = t.Focused.Directory.Foreground(fpBlue)
	t.Focused.Description = t.Focused.Description.Foreground(fpGreen)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(fpRed)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(fpRed)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(fpOrange)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(fpOrange)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(fpOrange)
	t.Focused.Option = t.Focused.Option.Foreground(normalFg)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(fpOrange)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(fpOrange)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(fpGreen).SetString("✓ ")

	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(normalFg)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(cream).Background(fpOrange)
	t.Focused.Next = t.Focused.FocusedButton

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(fpGreen)

	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(fpOrange)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	return t
}
