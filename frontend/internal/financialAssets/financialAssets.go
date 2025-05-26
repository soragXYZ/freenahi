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

const (
	accountNameColumn int = iota
	amountColumn
	repartitionColumn
	numberOfColumn
)

// A standard table, but which has resizabled column width
type customTable struct {
	widget.Table
}

func newCustomTable(length func() (rows int, cols int), create func() fyne.CanvasObject, update func(widget.TableCellID, fyne.CanvasObject)) *customTable {
	table := &customTable{}
	table.ExtendBaseWidget(table)

	table.Length = length
	table.CreateCell = create
	table.UpdateCell = update

	return table
}

// Function called when the table is resized: auto adjust the column width
func (t *customTable) Resize(size fyne.Size) {

	// Note that sometimes this function is not called even if it should
	// For example, when you quickly reduce the window size
	// Thus, the table width is not correctly auto adjusted,
	// the table is too big a scroller appears
	// No workaround ATM

	_, columns := t.Length()
	for i := range columns {
		// Make columns equally spaced
		t.Table.SetColumnWidth(i, t.Table.Size().Width/float32(columns)-3)
	}

	t.Table.Resize(size)
}

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

	// Fill accounts: backend call
	accounts := account.GetBankAccounts(app, "checking")
	total := 0.0

	for _, account := range accounts {
		total += float64(account.Balance)
	}

	bankingAccountTable := newCustomTable(
		func() (int, int) {
			return len(accounts), numberOfColumn
		},
		func() fyne.CanvasObject {
			label1 := widget.NewLabel("Template")
			label1.Alignment = fyne.TextAlignCenter
			label2 := widget.NewLabel("Template")
			label2.Alignment = fyne.TextAlignCenter
			return container.NewStack(
				container.NewHScroll(widget.NewLabel("Template")),
				label1,
				label2,
			)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			accountNameItem := o.(*fyne.Container).Objects[0].(*container.Scroll)
			amountItem := o.(*fyne.Container).Objects[1].(*widget.Label)
			repartitionItem := o.(*fyne.Container).Objects[2].(*widget.Label)

			switch id.Col {

			case accountNameColumn:
				accountNameItem.Show()
				amountItem.Hide()
				repartitionItem.Hide()
				accountNameItem.Content.(*widget.Label).SetText(accounts[id.Row].Original_name)

			case amountColumn:
				accountNameItem.Hide()
				amountItem.Show()
				repartitionItem.Hide()
				amountItem.SetText(helper.ValueSpacer(fmt.Sprintf("%0.2f", accounts[id.Row].Balance)))

			case repartitionColumn:
				accountNameItem.Hide()
				amountItem.Hide()
				repartitionItem.Show()
				repartitionItem.SetText(fmt.Sprintf("%0.2f %%", float64(accounts[id.Row].Balance)/total*100))
			}
		},
	)

	bankingAccountTable.ShowHeaderRow = true
	bankingAccountTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {

		l := o.(*widget.Label)

		switch id.Col {
		case accountNameColumn:
			l.SetText(lang.L("Account name"))
		case amountColumn:
			l.SetText(lang.L("Value"))
		case repartitionColumn:
			l.SetText(lang.L("Repartition"))
		default:
			helper.Logger.Fatal().Msg("Too much column in the bank account assets grid for header")
		}
	}

	totalItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total"), helper.ValueSpacer(fmt.Sprintf("%.2f", total))))
	totalItem.Alignment = fyne.TextAlignCenter
	totalItem.SizeName = theme.SizeNameHeadingText

	totalContainer := container.NewVBox(layout.NewSpacer(), totalItem, layout.NewSpacer())

	return container.NewBorder(nil, nil, nil, totalContainer, bankingAccountTable)
}
