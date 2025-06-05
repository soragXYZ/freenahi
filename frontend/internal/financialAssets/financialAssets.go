package financialassets

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"time"

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

type HistoryValuePoint struct {
	Valuation     float32
	DateValuation time.Time
}

const ( // for savings and bank account
	nameColumn int = iota
	valueColumn
	repartitionColumn
	numberOfColumns

	unselectTime = 200 * time.Millisecond
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

const (
	sortOff = iota // Used to sort data when clicking on a table header
	sortAsc
	sortDesc
	numberOfSorts
)

// Var holding the sort type for each
var columnSort = [numberOfColumns]int{}
var SFColumnSort = [SFnumberOfColumns]int{}

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

	// Set column header, sortable when taped https://fynelabs.com/2023/10/05/user-data-sorting-with-a-fyne-table-widget/
	bankingAccountTable.ShowHeaderRow = true
	bankingAccountTable.CreateHeader = func() fyne.CanvasObject {
		return widget.NewButton("000", func() {})
	}
	bankingAccountTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {

		b := o.(*widget.Button)

		switch id.Col {
		case nameColumn:
			b.SetText(lang.L("Account name"))
			helper.SetColumnHeaderIcon(columnSort[nameColumn], b, sortAsc, sortDesc)
		case valueColumn:
			b.SetText(lang.L("Value"))
			helper.SetColumnHeaderIcon(columnSort[valueColumn], b, sortAsc, sortDesc)
		case repartitionColumn:
			b.SetText(lang.L("Repartition"))
			helper.SetColumnHeaderIcon(columnSort[repartitionColumn], b, sortAsc, sortDesc)
		default:
			helper.Logger.Fatal().Msg("Too much column in the bank account assets grid for header")
		}

		b.OnTapped = func() {
			applySort(id.Col, &bankingAccountTable.Table, accounts)
		}
		b.Refresh()
	}

	bankingAccountTable.OnSelected = func(id widget.TableCellID) {
		go func() {
			time.Sleep(unselectTime)
			fyne.Do(func() {
				bankingAccountTable.Unselect(id)
			})
		}()

		w := app.NewWindow(fmt.Sprintf("%s: %s", accounts[id.Row].Bank_Original_name, accounts[id.Row].Original_name))
		w.CenterOnScreen()
		w.Resize(fyne.NewSize(600, 600))

		historicalData := GetHistoryValues(app, accounts[id.Row].Id, "all")

		// Sanitize data to create the graph
		var xLabel []string
		var yLabel []float64

		for _, point := range historicalData {
			xLabel = append(xLabel, point.DateValuation.Format("2006-01-02"))
			yLabel = append(yLabel, float64(point.Valuation))
		}

		graphItem := helper.DrawLine(
			xLabel,
			yLabel,
			fyne.NewSize(600, 600),
			"Line graph",
		)

		w.SetContent(container.NewHBox(graphItem))
		w.Show()

	}

	// Reload button reloads data by querying the backend
	reloadButton := widget.NewButton("", func() {
		accounts = account.GetBankAccounts(app, accountType)
		bankingAccountTable.Refresh()

		// Reset header sorting if any
		columnSort[0] = numberOfSorts
		applySort(0, &bankingAccountTable.Table, accounts)
	})

	reloadButton.Icon = theme.ViewRefreshIcon()

	bankingAccountItem := container.NewBorder(nil, container.NewBorder(nil, nil, nil, reloadButton, nil), nil, nil, bankingAccountTable)

	totalItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total"), helper.ValueSpacer(fmt.Sprintf("%.2f", total))))
	totalItem.Alignment = fyne.TextAlignCenter
	totalItem.SizeName = theme.SizeNameHeadingText

	totalContainer := container.NewBorder(
		nil,
		nil,
		widget.NewSeparator(),
		container.NewVBox(layout.NewSpacer(), totalItem, layout.NewSpacer()),
	)

	return container.NewBorder(nil, nil, nil, totalContainer, bankingAccountItem)
}

func NewStocksAndFundsScreen(app fyne.App) *fyne.Container {

	// Fill invests: backend call
	invests := GetInvests(app)
	total := 0.0

	for _, invest := range invests {
		total += float64(invest.Valuation)
	}

	// These values are used later to set column width sizes, which are the max between the header and an actual value
	testIconSize := widget.NewIcon(theme.RadioButtonCheckedIcon()).MinSize().Width

	assetNameHeader := widget.NewLabel(lang.L("Name"))
	assetNameHeader.TextStyle.Bold = true
	assetNameHeaderSize := assetNameHeader.MinSize().Width + testIconSize
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
	valueHeaderSize := valueHeader.MinSize().Width + testIconSize
	testValueLabelSize := widget.NewLabel("10 000 000").MinSize().Width

	repartitionHeader := widget.NewLabel(lang.L("Repartition"))
	repartitionHeader.TextStyle.Bold = true
	repartitionHeaderSize := repartitionHeader.MinSize().Width + testIconSize
	testRepartitionLabelSize := widget.NewLabel("100 %").MinSize().Width

	profitHeader := widget.NewLabel(lang.L("Profit"))
	profitHeader.TextStyle.Bold = true
	profitHeaderSize := profitHeader.MinSize().Width + testIconSize
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

	// Set column header, sortable when taped https://fynelabs.com/2023/10/05/user-data-sorting-with-a-fyne-table-widget/
	assetTable.ShowHeaderRow = true
	assetTable.CreateHeader = func() fyne.CanvasObject {
		return widget.NewButton("000", func() {})
	}

	assetTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {

		b := o.(*widget.Button)

		switch id.Col {
		case SFnameColumn:
			b.SetText(lang.L("Name"))
			helper.SetColumnHeaderIcon(SFColumnSort[SFnameColumn], b, sortAsc, sortDesc)

		case SFquantityColumn:
			b.SetText(lang.L("Quantity"))

		case SFunitCostColumn:
			b.SetText(lang.L("Unit cost"))

		case SFcurrentPriceColumn:
			b.SetText(lang.L("Current price"))

		case SFvalueColumn:
			b.SetText(lang.L("Value"))
			helper.SetColumnHeaderIcon(SFColumnSort[SFvalueColumn], b, sortAsc, sortDesc)

		case SFrepartitionColumn:
			b.SetText(lang.L("Repartition"))
			helper.SetColumnHeaderIcon(SFColumnSort[SFrepartitionColumn], b, sortAsc, sortDesc)

		case SFprofitColumn:
			b.SetText(lang.L("Profit"))
			helper.SetColumnHeaderIcon(SFColumnSort[SFprofitColumn], b, sortAsc, sortDesc)

		default:
			helper.Logger.Fatal().Msg("Too much column in the stocks and funds assets grid for header")
		}

		b.OnTapped = func() {
			applySortSF(id.Col, assetTable, invests)
		}
		b.Refresh()
	}

	assetTable.OnSelected = func(id widget.TableCellID) {
		go func() {
			time.Sleep(unselectTime)
			fyne.Do(func() {
				assetTable.Unselect(id)
			})
		}()

	}

	// Reload button reloads data by querying the backend
	reloadButton := widget.NewButton("", func() {
		invests = GetInvests(app)
		assetTable.Refresh()

		// Reset header sorting if any
		SFColumnSort[0] = numberOfSorts
		applySortSF(0, assetTable, invests)
	})

	reloadButton.Icon = theme.ViewRefreshIcon()

	assetTableItem := container.NewBorder(nil, container.NewBorder(nil, nil, nil, reloadButton, nil), nil, nil, assetTable)

	totalItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total"), helper.ValueSpacer(fmt.Sprintf("%.2f", total))))
	totalItem.Alignment = fyne.TextAlignCenter
	totalItem.SizeName = theme.SizeNameHeadingText

	totalContainer := container.NewBorder(
		nil,
		nil,
		widget.NewSeparator(),
		container.NewVBox(layout.NewSpacer(), totalItem, layout.NewSpacer()),
	)

	return container.NewBorder(nil, nil, nil, totalContainer, assetTableItem)
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

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/history" and retrieve value/date pairs for an account.
// Data are used later to draw line graphs
func GetHistoryValues(app fyne.App, account int, period string) []HistoryValuePoint {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	url := fmt.Sprintf("%s://%s:%s/history/%d?period=%s", backendProtocol, backendIp, backendPort, account, period)

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

	var values []HistoryValuePoint
	if err := json.Unmarshal(body, &values); err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot unmarshal HistoryValuePoint")
		return nil

	}

	return values
}

// Sort table data for banking and checking
func applySort(col int, t *widget.Table, data []account.BankAccount) {

	// Circle sorting: off => asc => desc => off => etc...
	order := columnSort[col]
	order++
	if order >= numberOfSorts {
		order = sortOff
	}

	// Reset all and assign tapped sort
	for i := range numberOfColumns {
		columnSort[i] = sortOff
	}

	columnSort[col] = order

	sort.Slice(data, func(i, j int) bool {
		a := data[i]
		b := data[j]

		// re-sort with no sort selected
		if order == sortOff {
			return a.Balance > b.Balance
		}

		switch col {
		case nameColumn:
			if order == sortAsc {
				return a.Original_name < b.Original_name
			}
			return a.Original_name > b.Original_name

		case valueColumn, repartitionColumn:
			if order == sortAsc {
				return a.Balance > b.Balance
			}
			return a.Balance < b.Balance

		default:
			return false
		}
	})
	t.Refresh()
}

// Sort table data for stocks and funds
func applySortSF(col int, t *widget.Table, data []Investment) {

	// Circle sorting: off => asc => desc => off => etc...
	order := SFColumnSort[col]
	order++
	if order >= numberOfSorts {
		order = sortOff
	}

	// Reset all and assign tapped sort
	for i := range SFnumberOfColumns {
		SFColumnSort[i] = sortOff
	}

	SFColumnSort[col] = order

	sort.Slice(data, func(i, j int) bool {
		a := data[i]
		b := data[j]

		// re-sort with no sort selected
		if order == sortOff {
			return a.Valuation > b.Valuation
		}

		switch col {
		case SFnameColumn:
			if order == sortAsc {
				return a.Label < b.Label
			}
			return a.Label > b.Label

		case SFvalueColumn, SFrepartitionColumn:
			if order == sortAsc {
				return a.Valuation < b.Valuation
			}
			return a.Valuation > b.Valuation

		case SFprofitColumn:
			if order == sortAsc {
				return a.Diff < b.Diff
			}
			return a.Diff > b.Diff

		default:
			return false
		}
	})
	t.Refresh()
}
