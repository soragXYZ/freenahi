package financialassets

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"freenahiFront/internal/account"
	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"
)

// https://docs.powens.com/api-reference/products/wealth-aggregation/investments#investment-object
type Investment struct {
	Invest_id    int     `json:"id"`
	Account_id   int     `json:"id_account"`
	Label        string  `json:"label"`
	Code         string  `json:"code"`
	Code_type    string  `json:"code_type"`
	Stock_symbol string  `json:"stock_symbol"`
	Quantity     float32 `json:"quantity"`
	Unit_price   float32 `json:"unitprice"`
	Unit_value   float32 `json:"unitvalue"`
	Valuation    float32 `json:"valuation"`
	Diff         float32 `json:"diff"`
	Diff_percent float32 `json:"diff_percent"`
	Last_update  string  `json:"last_update"`
}

const (
	nameColumn int = iota
	valueColumn
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
		container.NewTabItem(lang.L("Bank accounts"), NewScreen(app, "checking")),
		container.NewTabItem(lang.L("Savings books"), NewScreen(app, "savings")),
		container.NewTabItem(lang.L("Stocks and funds"), NewStocksAndFundsScreen(app)),
		container.NewTabItem(lang.L("Real estate"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
		container.NewTabItem(lang.L("Crypto"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

func NewScreen(app fyne.App, accountType string) *fyne.Container {

	if accountType != "checking" && accountType != "savings" {
		helper.Logger.Fatal().Msgf("Wrong type '%s' used for GetBankAccounts. Must be checking or savings", accountType)
	}

	// Fill accounts: backend call
	accounts := account.GetBankAccounts(app, accountType)
	total := 0.0

	for _, account := range accounts {
		total += float64(account.Balance)
	}

	bankingAccountTable := newCustomTable(
		func() (int, int) {
			return len(accounts), numberOfColumn
		},
		func() fyne.CanvasObject {
			scrollerLabel := widget.NewLabel("Template")
			scrollerLabel.Alignment = fyne.TextAlignCenter
			accountNameItem := container.NewHScroll(scrollerLabel)

			valueItem := widget.NewLabel("Template")
			valueItem.Alignment = fyne.TextAlignCenter

			repartitionItem := widget.NewLabel("Template")
			repartitionItem.Alignment = fyne.TextAlignCenter

			return container.NewStack(accountNameItem, valueItem, repartitionItem)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			accountNameItem := o.(*fyne.Container).Objects[0].(*container.Scroll)
			valueItem := o.(*fyne.Container).Objects[1].(*widget.Label)
			repartitionItem := o.(*fyne.Container).Objects[2].(*widget.Label)

			accountNameItem.Hide()
			valueItem.Hide()
			repartitionItem.Hide()

			switch id.Col {
			case nameColumn:
				accountNameItem.Show()
				accountNameItem.Content.(*widget.Label).SetText(accounts[id.Row].Original_name)

			case valueColumn:
				valueItem.Show()
				valueItem.SetText(helper.ValueSpacer(fmt.Sprintf("%0.2f", accounts[id.Row].Balance)))

			case repartitionColumn:
				repartitionItem.Show()
				repartitionItem.SetText(fmt.Sprintf("%0.2f %%", float64(accounts[id.Row].Balance)/total*100))
			}
		},
	)

	bankingAccountTable.ShowHeaderRow = true
	bankingAccountTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {

		l := o.(*widget.Label)

		switch id.Col {
		case nameColumn:
			l.SetText(lang.L("Account name"))
		case valueColumn:
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

	totalContainer := container.NewBorder(
		nil,
		nil,
		widget.NewSeparator(),
		container.NewVBox(layout.NewSpacer(), totalItem, layout.NewSpacer()),
	)

	return container.NewBorder(nil, nil, nil, totalContainer, bankingAccountTable)
}

func NewStocksAndFundsScreen(app fyne.App) *fyne.Container {

	// Fill invests: backend call
	invests := GetInvests(app)
	total := 0.0

	for _, invest := range invests {
		total += float64(invest.Valuation)
	}

	assetTable := newCustomTable(
		func() (int, int) {
			return len(invests), numberOfColumn
		},
		func() fyne.CanvasObject {
			scrollerLabel := widget.NewLabel("Template")
			scrollerLabel.Alignment = fyne.TextAlignCenter
			assetNameItem := container.NewHScroll(scrollerLabel)

			valueItem := widget.NewLabel("Template")
			valueItem.Alignment = fyne.TextAlignCenter

			repartitionItem := widget.NewLabel("Template")
			repartitionItem.Alignment = fyne.TextAlignCenter

			return container.NewStack(assetNameItem, valueItem, repartitionItem)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			assetNameItem := o.(*fyne.Container).Objects[0].(*container.Scroll)
			valueItem := o.(*fyne.Container).Objects[1].(*widget.Label)
			repartitionItem := o.(*fyne.Container).Objects[2].(*widget.Label)

			assetNameItem.Hide()
			valueItem.Hide()
			repartitionItem.Hide()

			switch id.Col {
			case nameColumn:
				assetNameItem.Show()
				assetNameItem.Content.(*widget.Label).SetText(invests[id.Row].Label)

			case valueColumn:
				valueItem.Show()
				valueItem.SetText(helper.ValueSpacer(fmt.Sprintf("%0.2f", invests[id.Row].Valuation)))

			case repartitionColumn:
				repartitionItem.Show()
				repartitionItem.SetText(fmt.Sprintf("%0.2f %%", float64(invests[id.Row].Valuation)/total*100))
			}
		},
	)

	assetTable.ShowHeaderRow = true
	assetTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {

		l := o.(*widget.Label)

		switch id.Col {
		case nameColumn:
			l.SetText(lang.L("Account name"))
		case valueColumn:
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

	totalContainer := container.NewBorder(
		nil,
		nil,
		widget.NewSeparator(),
		container.NewVBox(layout.NewSpacer(), totalItem, layout.NewSpacer()),
	)

	return container.NewBorder(nil, nil, nil, totalContainer, assetTable)
}

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/investment" and retrieve investments
func GetInvests(app fyne.App) []Investment {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	url := fmt.Sprintf("%s://%s:%s/investment/", backendProtocol, backendIp, backendPort)

	resp, err := http.Get(url)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot run http get request")
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("ReadAll error")
		return nil
	}

	var invests []Investment
	if err := json.Unmarshal(body, &invests); err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot unmarshal investments")
		return nil

	}

	return invests
}
