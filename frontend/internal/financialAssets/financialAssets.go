package financialassets

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
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

const ( // for savings and bank account
	nameColumn int = iota
	valueColumn
	repartitionColumn
	numberOfColumns
)

const ( // SF = stocks and funds
	SFnameColumn int = iota
	SFquantityColumn
	SFunitCostColumn
	SFcurrentPriceColumn
	SFvalueColumn
	SFrepartitionColumn
	SFprofitColumn
	SFnumberOfColumns
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
		container.NewTabItem(lang.L("Bank accounts"), NewCheckingOrSavingsScreen(app, "checking")),
		container.NewTabItem(lang.L("Savings books"), NewCheckingOrSavingsScreen(app, "savings")),
		container.NewTabItem(lang.L("Stocks and funds"), NewStocksAndFundsScreen(app)),
		container.NewTabItem(lang.L("Real estate"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
		container.NewTabItem(lang.L("Crypto"), container.NewVBox(layout.NewSpacer(), wipItem, layout.NewSpacer())),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

func NewCheckingOrSavingsScreen(app fyne.App, accountType string) *fyne.Container {

	if accountType != "checking" && accountType != "savings" {
		helper.Logger.Fatal().Msgf("Wrong type '%s' used for GetBankAccounts. Must be checking or savings", accountType)
	}

	accounts := account.GetBankAccounts(app, accountType) // Fill accounts: backend call
	total := 0.0                                          // Used later to calculate the repartition of each individual asset

	for _, account := range accounts {
		total += float64(account.Balance)
	}

	bankingAccountTable := newCustomTable(
		func() (int, int) {
			return len(accounts), numberOfColumns
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

	// These values are used later to set column width sizes, which are the max between the header and an actual value
	assetNameHeader := widget.NewLabel(lang.L("Name"))
	assetNameHeader.TextStyle.Bold = true
	assetNameHeaderSize := assetNameHeader.MinSize().Width
	testAssetNameLabelSize := widget.NewLabel("ISH WLD SWP PEA EU").MinSize().Width

	quantityHeader := widget.NewLabel(lang.L("Quantity"))
	quantityHeader.TextStyle.Bold = true
	quantityHeaderSize := quantityHeader.MinSize().Width
	testQuantityLabelSize := widget.NewLabel("10 000 000").MinSize().Width

	unitCostHeader := widget.NewLabel(lang.L("Unit cost"))
	unitCostHeader.TextStyle.Bold = true
	unitCostHeaderSize := unitCostHeader.MinSize().Width
	testUnitCostLabelSize := widget.NewLabel("10 000 000").MinSize().Width

	currentPriceHeader := widget.NewLabel(lang.L("Current price"))
	currentPriceHeader.TextStyle.Bold = true
	currentPriceHeaderSize := currentPriceHeader.MinSize().Width
	testCurrentPriceLabelSize := widget.NewLabel("10 000 000").MinSize().Width

	valueHeader := widget.NewLabel(lang.L("Value"))
	valueHeader.TextStyle.Bold = true
	valueHeaderSize := valueHeader.MinSize().Width
	testValueLabelSize := widget.NewLabel("10 000 000").MinSize().Width

	repartitionHeader := widget.NewLabel(lang.L("Repartition"))
	repartitionHeader.TextStyle.Bold = true
	repartitionHeaderSize := repartitionHeader.MinSize().Width
	testRepartitionLabelSize := widget.NewLabel("100 %").MinSize().Width

	profitHeader := widget.NewLabel(lang.L("Profit"))
	profitHeader.TextStyle.Bold = true
	profitHeaderSize := profitHeader.MinSize().Width
	testProfitLabel := widget.NewLabel("10 000 000")
	testProfitLabel.SizeName = theme.SizeNameCaptionText
	testProfitLabelSize := testProfitLabel.MinSize().Width

	assetTable := widget.NewTable(
		func() (int, int) {
			return len(invests), SFnumberOfColumns
		},
		func() fyne.CanvasObject {
			scrollerLabel := widget.NewLabel("Template")
			scrollerLabel.Alignment = fyne.TextAlignCenter
			scroller := container.NewHScroll(scrollerLabel)
			scroller.SetMinSize(fyne.NewSize(float32(
				math.Max(float64(testAssetNameLabelSize), float64(assetNameHeaderSize))),
				scroller.MinSize().Height),
			)

			ibanLabel := widget.NewLabel("Template")
			ibanLabel.TextStyle.Italic = true
			ibanLabel.SizeName = theme.SizeNameCaptionText
			ibanLabel.Selectable = true

			assetNameItem := container.NewVBox(scroller, container.NewCenter(ibanLabel))

			quantityItem := widget.NewLabel("Template")
			quantityItem.Alignment = fyne.TextAlignCenter

			unitCostItem := widget.NewLabel("Template")
			unitCostItem.Alignment = fyne.TextAlignCenter

			currentPriceItem := widget.NewLabel("Template")
			currentPriceItem.Alignment = fyne.TextAlignCenter

			valueItem := widget.NewLabel("Template")
			valueItem.Alignment = fyne.TextAlignCenter

			repartitionItem := widget.NewLabel("Template")
			repartitionItem.Alignment = fyne.TextAlignCenter

			profitTotalLabel := widget.NewLabel("Template")
			profitTotalLabel.Alignment = fyne.TextAlignCenter
			profitTotalLabel.SizeName = theme.SizeNameCaptionText
			profitRelativeLabel := widget.NewLabel("Template")
			profitRelativeLabel.Alignment = fyne.TextAlignCenter
			profitRelativeLabel.SizeName = theme.SizeNameCaptionText
			profitItem := container.NewVBox(profitTotalLabel, profitRelativeLabel)

			return container.NewStack(
				container.NewCenter(assetNameItem),
				container.NewCenter(quantityItem),
				container.NewCenter(unitCostItem),
				container.NewCenter(currentPriceItem),
				container.NewCenter(valueItem),
				container.NewCenter(repartitionItem),
				profitItem,
			)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			assetNameItem := o.(*fyne.Container).Objects[0].(*fyne.Container)
			quantityItem := o.(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
			unitCostItem := o.(*fyne.Container).Objects[2].(*fyne.Container).Objects[0].(*widget.Label)
			currentPriceItem := o.(*fyne.Container).Objects[3].(*fyne.Container).Objects[0].(*widget.Label)
			valueItem := o.(*fyne.Container).Objects[4].(*fyne.Container).Objects[0].(*widget.Label)
			repartitionItem := o.(*fyne.Container).Objects[5].(*fyne.Container).Objects[0].(*widget.Label)
			profitItem := o.(*fyne.Container).Objects[6].(*fyne.Container)

			assetNameItem.Hide()
			quantityItem.Hide()
			unitCostItem.Hide()
			currentPriceItem.Hide()
			valueItem.Hide()
			repartitionItem.Hide()
			profitItem.Hide()

			switch id.Col {
			case SFnameColumn:
				assetNameItem.Show()
				name := assetNameItem.Objects[0].(*fyne.Container).Objects[0].(*container.Scroll).Content.(*widget.Label)
				name.SetText(invests[id.Row].Label)
				isin := assetNameItem.Objects[0].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
				if invests[id.Row].Code_type == "ISIN" {
					isin.Show()
					isin.SetText(invests[id.Row].Code)
				} else {
					isin.Hide()
				}

			case SFquantityColumn:
				quantityItem.Show()
				quantityItem.SetText(fmt.Sprintf("%.2f", invests[id.Row].Quantity))

			case SFunitCostColumn:
				unitCostItem.Show()
				unitCostItem.SetText(helper.ValueSpacer(fmt.Sprintf("%.2f", invests[id.Row].Unit_price)))

			case SFcurrentPriceColumn:
				currentPriceItem.Show()
				currentPriceItem.SetText(helper.ValueSpacer(fmt.Sprintf("%.2f", invests[id.Row].Unit_value)))

			case SFvalueColumn:
				valueItem.Show()
				valueItem.SetText(helper.ValueSpacer(fmt.Sprintf("%0.2f", invests[id.Row].Valuation)))

			case SFrepartitionColumn:
				repartitionItem.Show()
				repartitionItem.SetText(fmt.Sprintf("%0.2f %%", float64(invests[id.Row].Valuation)/total*100))

			case SFprofitColumn:
				profitItem.Show()

				totalProfit := profitItem.Objects[0].(*widget.Label)
				relativeProfit := profitItem.Objects[1].(*widget.Label)
				if invests[id.Row].Diff > 0 {
					totalProfit.Importance = widget.SuccessImportance
					relativeProfit.Importance = widget.SuccessImportance
				} else if invests[id.Row].Diff < 0 {
					totalProfit.Importance = widget.DangerImportance
					relativeProfit.Importance = widget.DangerImportance
				} else {
					totalProfit.Importance = widget.MediumImportance
					relativeProfit.Importance = widget.MediumImportance
				}
				totalProfit.SetText(fmt.Sprintf("%.2f", invests[id.Row].Diff))
				relativeProfit.SetText(fmt.Sprintf("%.2f %%", invests[id.Row].Diff_percent*100))
			}
		},
	)

	// We set the width of the columns, ie the max between the language name header size and actual value
	// For example, the max between string "Value" and "-123456123.00", or string "Montant" and "-123456123.00" in french
	assetTable.SetColumnWidth(SFnameColumn, float32(math.Max(float64(testAssetNameLabelSize), float64(assetNameHeaderSize))))
	assetTable.SetColumnWidth(SFquantityColumn, float32(math.Max(float64(testQuantityLabelSize), float64(quantityHeaderSize))))
	assetTable.SetColumnWidth(SFunitCostColumn, float32(math.Max(float64(testUnitCostLabelSize), float64(unitCostHeaderSize))))
	assetTable.SetColumnWidth(SFcurrentPriceColumn, float32(math.Max(float64(testCurrentPriceLabelSize), float64(currentPriceHeaderSize))))
	assetTable.SetColumnWidth(SFvalueColumn, float32(math.Max(float64(testValueLabelSize), float64(valueHeaderSize))))
	assetTable.SetColumnWidth(SFrepartitionColumn, float32(math.Max(float64(testRepartitionLabelSize), float64(repartitionHeaderSize))))
	assetTable.SetColumnWidth(SFprofitColumn, float32(math.Max(float64(testProfitLabelSize), float64(profitHeaderSize))))

	// Add header to the table
	assetTable.ShowHeaderRow = true
	assetTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {

		l := o.(*widget.Label)

		switch id.Col {
		case SFnameColumn:
			l.SetText(lang.L("Name"))
		case SFquantityColumn:
			l.SetText(lang.L("Quantity"))
		case SFunitCostColumn:
			l.SetText(lang.L("Unit cost"))
		case SFcurrentPriceColumn:
			l.SetText(lang.L("Current price"))
		case SFvalueColumn:
			l.SetText(lang.L("Value"))
		case SFrepartitionColumn:
			l.SetText(lang.L("Repartition"))
		case SFprofitColumn:
			l.SetText(lang.L("Profit"))
		default:
			helper.Logger.Fatal().Msg("Too much column in the stocks and funds assets grid for header")
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
