package financialassets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"freenahiFront/internal/account"
	"freenahiFront/internal/helper"
)

// Create the Financial asset tab view
func NewFinancialAssetsScreen(app fyne.App, win fyne.Window) *container.AppTabs {

	wipItem := widget.NewLabel("Work in progress")
	wipItem.SizeName = theme.SizeNameHeadingText
	wipItem.Alignment = fyne.TextAlignCenter
	wipItem.TextStyle.Bold = true
	wipItem.TextStyle.Italic = true

	tabs := container.NewAppTabs(
		container.NewTabItem(lang.L("General"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
		container.NewTabItem(lang.L("Bank accounts"), NewBankAccountsScreen(app)),
		container.NewTabItem(lang.L("Savings books"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
		container.NewTabItem(lang.L("Real estate"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
		container.NewTabItem(lang.L("Crypto"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
		container.NewTabItem(lang.L("Stocks and funds"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
		container.NewTabItem(lang.L("Crypto"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

func NewBankAccountsScreen(app fyne.App) *fyne.Container {

	accountNameHeaderLabel := widget.NewLabel(lang.L("Account name"))
	accountNameHeaderLabel.Alignment = fyne.TextAlignCenter
	accountNameHeaderLabel.TextStyle.Bold = true

	valueHeaderLabel := widget.NewLabel(lang.L("Value"))
	valueHeaderLabel.Alignment = fyne.TextAlignCenter
	valueHeaderLabel.TextStyle.Bold = true

	repartitionHeaderLabel := widget.NewLabel(lang.L("Repartition"))
	repartitionHeaderLabel.Alignment = fyne.TextAlignCenter
	repartitionHeaderLabel.TextStyle.Bold = true

	headerContainer := container.NewGridWithColumns(
		3,
		accountNameHeaderLabel,
		valueHeaderLabel,
		repartitionHeaderLabel,
	)

	// Fill accounts: backend call
	accounts := account.GetBankAccounts(app, "checking")
	total := 0.0

	for _, account := range accounts {
		total += float64(account.Balance)
	}

	tableData := widget.NewList(
		func() int {
			return len(accounts)
		},
		func() fyne.CanvasObject {
			item1 := widget.NewLabel("Template")
			item1.Alignment = fyne.TextAlignCenter

			item2 := widget.NewLabel("Template")
			item2.Alignment = fyne.TextAlignCenter

			item3 := widget.NewLabel("Template")
			item3.Alignment = fyne.TextAlignCenter
			return container.NewGridWithColumns(3, container.NewHScroll(item1), item2, item3)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {

			accountNameItem := o.(*fyne.Container).Objects[0].(*container.Scroll).Content.(*widget.Label)
			amountItem := o.(*fyne.Container).Objects[1].(*widget.Label)
			repartitionItem := o.(*fyne.Container).Objects[2].(*widget.Label)

			accountNameItem.SetText(accounts[i].Original_name)
			amountItem.SetText(helper.ValueSpacer(fmt.Sprintf("%0.2f", accounts[i].Balance)))
			repartitionItem.SetText(fmt.Sprintf("%0.2f %%", float64(accounts[i].Balance)/total*100))
		},
	)

	table := container.NewBorder(container.NewVBox(headerContainer), nil, nil, nil, tableData)

	totalItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total"), helper.ValueSpacer(fmt.Sprintf("%.2f", total))))
	totalItem.Alignment = fyne.TextAlignCenter
	totalItem.SizeName = theme.SizeNameHeadingText

	totalContainer := container.NewVBox(layout.NewSpacer(), totalItem, layout.NewSpacer())

	return container.NewBorder(nil, nil, nil, totalContainer, table)
}
