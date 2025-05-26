package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"
	"time"

	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Const which are used in order to increase code readability
// These consts are the name of the columns and used later in a switch case
const (
	pinnedColumn int = iota
	dateColumn
	valueColumn
	typeColumn
	detailsColumn
	deleteColumn
	numberOfColumn

	unselectTime = 200 * time.Millisecond
	detailsRegex = `^[A-Za-z0-9_-]{1,50}$`
)

// The struct which is returned by the backend
type Transaction struct {
	Id               int     `json:"id"`
	Pinned           bool    `json:"pinned"`
	Date             string  `json:"date"`
	Value            float32 `json:"value"`
	Transaction_type string  `json:"type"`
	Original_wording string  `json:"original_wording"`
}

// A standard table, but which has resizabled column width
type customTable struct {
	widget.Table

	pinnedColumnWidth, dateColumnWidth, valueColumnWidth, typeColumnWidth, detailsColumnWidth, deleteColumnWidth float32
}

func newCustomTable(pinnedColumnWidth, dateColumnWidth, valueColumnWidth, typeColumnWidth, detailsColumnWidth, deleteColumnWidth float32, length func() (rows int, cols int), create func() fyne.CanvasObject, update func(widget.TableCellID, fyne.CanvasObject)) *customTable {
	table := &customTable{
		pinnedColumnWidth:  pinnedColumnWidth,
		dateColumnWidth:    dateColumnWidth,
		valueColumnWidth:   valueColumnWidth,
		typeColumnWidth:    typeColumnWidth,
		detailsColumnWidth: detailsColumnWidth,
		deleteColumnWidth:  deleteColumnWidth,
	}
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
	// if the table is too big a scroller appears
	// No workaround ATM

	t.Table.SetColumnWidth(pinnedColumn, t.pinnedColumnWidth)
	t.Table.SetColumnWidth(dateColumn, t.dateColumnWidth)
	t.Table.SetColumnWidth(valueColumn, t.valueColumnWidth)
	t.Table.SetColumnWidth(typeColumn, t.typeColumnWidth)
	t.Table.SetColumnWidth(deleteColumn, t.deleteColumnWidth)

	// Get the remaining space for the column "details"
	value := t.Table.Size().Width - 5*4 - t.pinnedColumnWidth - t.dateColumnWidth - t.valueColumnWidth - t.typeColumnWidth - t.deleteColumnWidth

	if value > t.detailsColumnWidth {
		t.Table.SetColumnWidth(detailsColumn, value)
	} else {
		t.Table.SetColumnWidth(detailsColumn, t.detailsColumnWidth)
	}

	t.Table.Resize(size)
}

// Create the transaction screen
func NewTransactionScreen(app fyne.App, win fyne.Window) fyne.CanvasObject {

	transactionTable := createTransactionTable(app, win)

	return container.NewBorder(nil, nil, nil, nil, transactionTable)
}

// Create the transaction table
func createTransactionTable(app fyne.App, win fyne.Window) *customTable {

	pinnedLabel := widget.NewLabel(lang.L("Pinned"))
	pinnedLabel.TextStyle.Bold = true

	dateLabel := widget.NewLabel(lang.L("Date"))
	dateLabel.TextStyle.Bold = true
	testDateLabelSize := widget.NewLabel("XXXX-YY-ZZ").MinSize().Width

	valueLabel := widget.NewLabel(lang.L("Value"))
	valueLabel.TextStyle.Bold = true
	testValueLabelSize := widget.NewLabel("-123456123.00").MinSize().Width

	typeLabel := widget.NewLabel(lang.L("Type"))
	typeLabel.TextStyle.Bold = true
	testTypeLabelSize := widget.NewLabel(lang.L("loan_repayment")).MinSize().Width

	detailsLabel := widget.NewLabel(lang.L("Details"))
	detailsLabel.TextStyle.Bold = true
	testDetailsLabelSize := widget.NewLabel("CB DEBIT IMMEDIAT UBER EATS").MinSize().Width

	deleteLabel := widget.NewLabel(lang.L("Delete"))
	deleteLabel.TextStyle.Bold = true

	testIconSize := widget.NewIcon(theme.RadioButtonCheckedIcon()).MinSize().Width

	// Fill txs with the first page of txs.
	txs := []Transaction{}
	txs = append(txs, getTransactions(1, app)...)
	var txsPerPage = 50 // Default number of txs returned by the backend when querrying the endpoint "/transaction"
	var reachedDataEnd = false
	var threshold = 5 // Ask more data from the backend if we only have less than "threshold" txs left to display

	txList := newCustomTable(

		// We set the width of the columns, ie the max between the language name header size and actual value
		// For example, the max between "Value" and "-123456123.00", or "Montant" and "-123456123.00" in french
		float32(math.Max(float64(testIconSize), float64(pinnedLabel.MinSize().Width))),
		float32(math.Max(float64(testDateLabelSize), float64(dateLabel.MinSize().Width))),
		float32(math.Max(float64(testValueLabelSize), float64(valueLabel.MinSize().Width))),
		float32(math.Max(float64(testTypeLabelSize), float64(typeLabel.MinSize().Width))),
		float32(math.Max(float64(testDetailsLabelSize), float64(detailsLabel.MinSize().Width))),
		float32(math.Max(float64(testIconSize), float64(deleteLabel.MinSize().Width))),

		func() (int, int) {
			return len(txs), numberOfColumn
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("Template"))
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			// Clean the cell from the previous value
			o.(*fyne.Container).RemoveAll()

			// Update the cell by adding content according to its "type" (icon, date, value, type, details, delete)
			switch id.Col {
			case pinnedColumn:
				if txs[id.Row].Pinned {
					helper.AddHAligned(o, widget.NewIcon(theme.RadioButtonCheckedIcon()))
				} else {
					helper.AddHAligned(o, widget.NewIcon(theme.RadioButtonIcon()))
				}

			case dateColumn:
				parsedTxDate, err := time.Parse("2006-01-02 15:04:05", txs[id.Row].Date)
				if err != nil {
					helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", txs[id.Row].Date)
				}
				helper.AddHAligned(o, widget.NewLabel(parsedTxDate.Format("2006-01-02")))

			case valueColumn:
				value := fmt.Sprintf("%.2f", txs[id.Row].Value)
				valueText := widget.NewLabel(helper.ValueSpacer(value))

				if txs[id.Row].Value > 0 {
					valueText.Importance = widget.SuccessImportance
				}

				helper.AddHAligned(o, valueText)

			case typeColumn:
				// ToDo: display an icon instead of a text ? More user friendly
				// Each type is here: https://docs.powens.com/api-reference/products/data-aggregation/bank-transactions#transactiontype-values
				helper.AddHAligned(o, widget.NewLabel(lang.L(txs[id.Row].Transaction_type)))

			case detailsColumn:
				scroller := container.NewHScroll(widget.NewLabel(txs[id.Row].Original_wording))

				// To Do: calculate the size of txTable (itself?), get a scroller with cell size if not large enough
				// Otherwise, simple label and center text ?
				// value := win.Canvas().Size().Width - 5*4 - float32(math.Max(float64(testIconSize), float64(pinnedLabel.MinSize().Width))) - float32(math.Max(float64(testDateLabelSize), float64(dateLabel.MinSize().Width))) - float32(math.Max(float64(testValueLabelSize), float64(valueLabel.MinSize().Width))) - float32(math.Max(float64(testTypeLabelSize), float64(typeLabel.MinSize().Width))) - float32(math.Max(float64(testIconSize), float64(deleteLabel.MinSize().Width)))
				// zzz := float32(math.Max(math.Max(float64(testDetailsLabelSize), float64(detailsLabel.MinSize().Width)), float64(value)))
				// scroller.SetMinSize(fyne.NewSize(zzz, scroller.MinSize().Height))

				scroller.SetMinSize(fyne.NewSize(testDetailsLabelSize, scroller.MinSize().Height))
				helper.AddHAligned(o, scroller)

			case deleteColumn:
				helper.AddHAligned(o, widget.NewIcon(theme.DeleteIcon()))

			default:
				helper.Logger.Fatal().Msg("Too much column in the transaction grid")
			}

			// Load new items in the list when the user scrolled near the bottom of the page => infinite scrolling
			// We ask more data from the backend if we only have less than "threshold" txs left to display
			if id.Row > len(txs)-threshold && !reachedDataEnd {
				pageRequested := len(txs)/txsPerPage + 1
				newTxs := getTransactions(pageRequested, app)

				// We have retrieved every transaction if the backend sent less txs than the default number per page
				if len(newTxs) < txsPerPage {
					reachedDataEnd = true
				}
				txs = append(txs, newTxs...)
			}
		},
	)

	// Add header to the table
	txList.ShowHeaderRow = true
	txList.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {

		label := o.(*widget.Label)

		switch id.Col {
		case pinnedColumn:
			label.SetText(lang.L("Pinned"))
		case dateColumn:
			label.SetText(lang.L("Date"))
		case valueColumn:
			label.SetText(lang.L("Value"))
		case typeColumn:
			label.SetText(lang.L("Type"))
		case detailsColumn:
			label.SetText(lang.L("Details"))
		case deleteColumn:
			label.SetText(lang.L("Delete"))
		default:
			helper.Logger.Fatal().Msg("Too much column in the grid for tx header")
		}
	}

	txList.OnSelected = func(id widget.TableCellID) {

		// Dirty "workaround" for the customTable Resize issue
		txList.Resize(txList.Size())

		// Update the cell when selected by modifying content according to its "type" (icon, date, value, type, details, delete)
		switch id.Col {
		case pinnedColumn:
			txs[id.Row].Pinned = !txs[id.Row].Pinned
			txList.RefreshItem(widget.TableCellID{Row: id.Row, Col: pinnedColumn})
			updateTransaction(txs[id.Row], app)

		case dateColumn, valueColumn, typeColumn:
			// Nothing to do here, pass

		case detailsColumn:

			detailsItem := widget.NewEntry()
			detailsItem.SetText(txs[id.Row].Original_wording)
			detailsItem.Validator = validation.NewRegexp(detailsRegex, lang.L("Regex tx details"))

			items := []*widget.FormItem{widget.NewFormItem(lang.L("Details"), detailsItem)}

			d := dialog.NewForm(lang.L("Edit transaction"), lang.L("Update"), lang.L("Cancel"), items, func(b bool) {
				if !b {
					return
				}

				txs[id.Row].Original_wording = detailsItem.Text // replaced by the user input
				txList.RefreshItem(widget.TableCellID{Row: id.Row, Col: detailsColumn})
				updateTransaction(txs[id.Row], app)
			}, win)

			d.Resize(fyne.NewSize(d.MinSize().Width*2, d.MinSize().Height))
			d.Show()

		case deleteColumn:
			cnf := dialog.NewConfirm(lang.L("Delete"), lang.L("Delete confirmation"), func(b bool) {
				if !b {
					return
				}

				deleteTransaction(txs[id.Row], app)
				txs = slices.Delete(txs, id.Row, id.Row+1) // delete the selected row only
				txList.Refresh()
			}, win)
			cnf.SetDismissText(lang.L("Cancel"))
			cnf.SetConfirmText(lang.L("Delete"))
			cnf.Show()

		default:
			helper.Logger.Fatal().Msg("Too much column in the transaction grid")
		}

		go func() {
			time.Sleep(unselectTime)
			fyne.Do(func() {
				txList.Unselect(id)
			})
		}()

	}

	return txList
}

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/transaction" and retrieve txs of the selected page
func getTransactions(page int, app fyne.App) []Transaction {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	url := fmt.Sprintf("%s://%s:%s/transaction?page=%d", backendProtocol, backendIp, backendPort, page)
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

	var txs []Transaction
	if err := json.Unmarshal(body, &txs); err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot unmarshal transactions")
		return nil

	}

	return txs
}

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/transaction" and update the specified tx
func updateTransaction(tx Transaction, app fyne.App) {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	url := fmt.Sprintf("%s://%s:%s/transaction/%d", backendProtocol, backendIp, backendPort, tx.Id)

	jsonBody, err := json.Marshal(tx)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot marshal tx")
		return
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot create new request")
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot run http put request")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		helper.Logger.Error().Msg(resp.Status)
		return
	}
}

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/transaction" and delete the specified tx
func deleteTransaction(tx Transaction, app fyne.App) {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	url := fmt.Sprintf("%s://%s:%s/transaction/%d", backendProtocol, backendIp, backendPort, tx.Id)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot create new request")
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot run http delete request")
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		helper.Logger.Error().Msg(resp.Status)
		return
	}
}
