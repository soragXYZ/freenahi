package financialassets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewFinancialAssetsScreen(app fyne.App, win fyne.Window) *fyne.Container {

	label := widget.NewLabel("Work in progress")
	label.SizeName = theme.SizeNameHeadingText
	label.TextStyle.Bold = true
	label.TextStyle.Italic = true
	return container.NewHBox(label)
}
