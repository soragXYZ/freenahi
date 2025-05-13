package financialassets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func NewFinancialAssetsScreen(app fyne.App, win fyne.Window) *fyne.Container {

	label := canvas.NewText("Work in progress", color.White)
	label.TextSize = 50
	label.TextStyle.Bold = true
	label.TextStyle.Italic = true
	return container.NewVBox(layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), label, layout.NewSpacer()), layout.NewSpacer())
}
