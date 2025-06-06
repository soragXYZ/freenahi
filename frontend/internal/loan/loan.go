package loan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	subscriptionDateColumn int = iota // Used to name the column
	valueColumn
	durationColumn
	numberOfColumns

	unselectTime = 200 * time.Millisecond
)

const (
	sortOff = iota // Used to sort data when clicking on a table header
	sortAsc
	sortDesc
	numberOfSorts
)

// Var holding the sort type for each
var columnSort = [numberOfColumns]int{}

type Loan struct {
	Loan_account_id      int     `json:"-"` // absent in base data, field added for simplicity
	Total_amount         float64 `json:"total_amount"`
	Available_amount     float64 `json:"available_amount"`
	Used_amount          float64 `json:"used_amount"`
	Subscription_date    string  `json:"subscription_date"`
	Maturity_date        string  `json:"maturity_date"`
	Start_repayment_date string  `json:"start_repayment_date"`
	Deferred             bool    `json:"deferred"`
	Next_payment_amount  float64 `json:"next_payment_amount"`
	Next_payment_date    string  `json:"next_payment_date"`
	Rate                 float64 `json:"rate"`
	Nb_payments_left     uint    `json:"nb_payments_left"`
	Nb_payments_done     uint    `json:"nb_payments_done"`
	Nb_payments_total    uint    `json:"nb_payments_total"`
	Last_payment_amount  float64 `json:"last_payment_amount"`
	Last_payment_date    string  `json:"last_payment_date"`
	Account_label        string  `json:"account_label"`
	Insurance_label      string  `json:"insurance_label"`
	Insurance_amount     float64 `json:"insurance_amount"`
	Insurance_rate       float64 `json:"insurance_rate"`
	Duration             uint    `json:"duration"`
	Loan_type            string  `json:"type"`
}

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

// Create the main view for loans
func NewLoanScreen(app fyne.App, win fyne.Window) *fyne.Container {

	loanTable := createLoanTable(app)

	return loanTable
}

// Create the table of of loan
func createLoanTable(app fyne.App) *fyne.Container {

	// Fill loans. Backend call
	loans := getLoans(app)

	loanTable := newCustomTable(
		func() (int, int) {
			return len(loans), numberOfColumns
		},
		func() fyne.CanvasObject {
			item := widget.NewLabel("Template")
			item.Alignment = fyne.TextAlignCenter
			return item
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			item := o.(*widget.Label)

			switch id.Col {
			case subscriptionDateColumn:
				if loans[id.Row].Subscription_date != "" { // Parse the date and keep only YYYY-MM-DD
					parsedSubscriptionDate, err := time.Parse("2006-01-02 15:04:05", loans[id.Row].Subscription_date)
					if err != nil {
						helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", loans[id.Row].Subscription_date)
					}
					item.SetText(parsedSubscriptionDate.Format("2006-01-02"))
				} else {
					item.SetText(lang.L("Irrelevant"))
				}

			case valueColumn:
				item.SetText(helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id.Row].Total_amount)))

			case durationColumn:
				item.SetText(fmt.Sprintf("%d", loans[id.Row].Duration))
			default:
				helper.Logger.Fatal().Msg("Too much column in the account grid")
			}
		},
	)

	// Set column header, sortable when taped https://fynelabs.com/2023/10/05/user-data-sorting-with-a-fyne-table-widget/
	loanTable.ShowHeaderRow = true
	loanTable.CreateHeader = func() fyne.CanvasObject {
		return widget.NewButton("000", func() {})
	}
	loanTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		b := o.(*widget.Button)

		switch id.Col {
		case subscriptionDateColumn:
			b.SetText(lang.L("Subscription date"))
			helper.SetColumnHeaderIcon(columnSort[subscriptionDateColumn], b, sortAsc, sortDesc)

		case valueColumn:
			b.SetText(lang.L("Value"))
			helper.SetColumnHeaderIcon(columnSort[valueColumn], b, sortAsc, sortDesc)

		case durationColumn:
			b.SetText(lang.L("Duration"))
			helper.SetColumnHeaderIcon(columnSort[durationColumn], b, sortAsc, sortDesc)
		}

		b.OnTapped = func() {
			applySort(id.Col, &loanTable.Table, loans)
		}
		b.Refresh()
	}

	// Display additional details when a loan is clicked on
	loanTable.OnSelected = func(id widget.TableCellID) {

		// Dirty "workaround" for the customTable Resize issue
		loanTable.Resize(loanTable.Size())

		go func() { // Unselect the cell after some time
			time.Sleep(unselectTime)
			fyne.Do(func() {
				loanTable.Unselect(id)
			})
		}()

		w := app.NewWindow(fmt.Sprintf("%s : %d", lang.L("Loan"), id.Row))
		w.CenterOnScreen()

		// Calculate the interest and capital reimbursed for the current (n+1) mensuality
		remainingCapital := loans[id.Row].Total_amount

		var periodInterest float64    // The interest amount for the current (n+1) mensuality
		var sumPeriodInterest float64 // The sum of interests paid for this loan at the moment
		var periodCapital float64

		for range loans[id.Row].Nb_payments_done {

			periodInterest = loans[id.Row].Rate / 100 * float64(remainingCapital) / 12
			periodCapital = loans[id.Row].Next_payment_amount - loans[id.Row].Insurance_amount - periodInterest

			sumPeriodInterest += periodInterest
			remainingCapital = remainingCapital - periodCapital
		}

		// =======================================================================================
		// Top Left Box
		loanTypeItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Type"), lang.L(loans[id.Row].Loan_type)))
		loanTypeItem.Alignment = fyne.TextAlignCenter
		loanTypeItem.SizeName = theme.SizeNameSubHeadingText

		amountItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Amount"), helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id.Row].Total_amount))))
		amountItem.Alignment = fyne.TextAlignCenter
		amountItem.SizeName = theme.SizeNameSubHeadingText

		durationItem := widget.NewLabel(fmt.Sprintf("%s: %d", lang.L("Duration"), loans[id.Row].Nb_payments_total))
		durationItem.Alignment = fyne.TextAlignCenter
		durationItem.SizeName = theme.SizeNameSubHeadingText

		rateItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f %%", lang.L("Rate"), loans[id.Row].Rate))
		rateItem.Alignment = fyne.TextAlignCenter
		rateItem.SizeName = theme.SizeNameSubHeadingText

		mensualitiesLeftItem := widget.NewLabel(fmt.Sprintf("%s: %d", lang.L("Mensualities left"), loans[id.Row].Nb_payments_left))
		mensualitiesLeftItem.Alignment = fyne.TextAlignCenter

		mensualitiesPaidItem := widget.NewLabel(fmt.Sprintf("%s: %d", lang.L("Mensualities paid"), loans[id.Row].Nb_payments_done))
		mensualitiesPaidItem.Alignment = fyne.TextAlignCenter

		var endingDateItem *widget.Label
		if loans[id.Row].Maturity_date != "" {
			endingDate, err := time.Parse("2006-01-02 15:04:05", loans[id.Row].Maturity_date)
			if err != nil {
				helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", loans[id.Row].Maturity_date)
			}

			endingDateItem = widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Ending date"), endingDate.Format("2006-01-02")))
		} else {
			endingDateItem = widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Ending date"), lang.L("Unknown")))
		}
		endingDateItem.Alignment = fyne.TextAlignCenter

		topLeftBox := container.NewBorder(
			nil,
			widget.NewSeparator(),
			nil,
			widget.NewSeparator(),
			container.NewVBox(
				container.NewGridWithColumns(2,
					loanTypeItem,
					amountItem),
				widget.NewSeparator(),
				container.NewGridWithColumns(2,
					durationItem,
					rateItem),
				widget.NewSeparator(),
				container.NewGridWithColumns(2,
					mensualitiesPaidItem,
					mensualitiesLeftItem),
				widget.NewSeparator(),
				endingDateItem,
			),
		)

		// =======================================================================================
		// Top right box
		nextPeriodPaymentGraph := helper.DrawDoughnut(
			[]string{lang.L("Capital"), lang.L("Insurance"), lang.L("Interests")},
			[]float64{periodCapital, loans[id.Row].Insurance_amount, periodInterest},
			fyne.NewSize(120, 120),
			"Next period payment",
		)

		mensualityItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Next mensuality"), loans[id.Row].Next_payment_amount))
		mensualityItem.Alignment = fyne.TextAlignCenter
		mensualityItem.SizeName = theme.SizeNameSubHeadingText

		periodCapitalItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Capital"), periodCapital))
		periodCapitalItem.Alignment = fyne.TextAlignCenter
		periodCapitalItem.SizeName = theme.SizeNameCaptionText

		periodInterestItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Interests"), periodInterest))
		periodInterestItem.Alignment = fyne.TextAlignCenter

		periodInterestItem.SizeName = theme.SizeNameCaptionText

		periodInsuranceItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Insurance"), loans[id.Row].Insurance_amount))
		periodInsuranceItem.Alignment = fyne.TextAlignCenter

		periodInsuranceItem.SizeName = theme.SizeNameCaptionText

		topRightBox := container.NewBorder(
			nil,
			widget.NewSeparator(),
			widget.NewSeparator(),
			nil,
			container.NewVBox(
				mensualityItem,
				container.NewHBox(
					layout.NewSpacer(),
					container.NewVBox(
						layout.NewSpacer(),
						widget.NewSeparator(),
						periodCapitalItem,
						widget.NewSeparator(),
						periodInterestItem,
						widget.NewSeparator(),
						periodInsuranceItem,
						widget.NewSeparator(),
						layout.NewSpacer(),
					),
					nextPeriodPaymentGraph,
					layout.NewSpacer(),
				),
			),
		)

		// =======================================================================================
		// Bottom left box
		totalToRefund := (loans[id.Row].Next_payment_amount - loans[id.Row].Insurance_amount) * float64(loans[id.Row].Nb_payments_total)
		paidInterest := totalToRefund - loans[id.Row].Total_amount

		totalPaymentGraph := helper.DrawDoughnut(
			[]string{lang.L("Capital"), lang.L("Interests")},
			[]float64{loans[id.Row].Total_amount, paidInterest},
			fyne.NewSize(120, 120),
			"Next period payment",
		)

		totalToRefundItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total loan cost"), helper.ValueSpacer(fmt.Sprintf("%0.2f", totalToRefund))))
		totalToRefundItem.Alignment = fyne.TextAlignCenter
		totalToRefundItem.SizeName = theme.SizeNameSubHeadingText

		totalCapitalItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Capital"), helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id.Row].Total_amount))))
		totalCapitalItem.Alignment = fyne.TextAlignCenter
		totalCapitalItem.SizeName = theme.SizeNameCaptionText

		totalInterestItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total interest to pay"), helper.ValueSpacer(fmt.Sprintf("%0.2f", paidInterest))))
		totalInterestItem.Alignment = fyne.TextAlignCenter
		totalInterestItem.SizeName = theme.SizeNameCaptionText

		bottomLeftBox := container.NewBorder(
			widget.NewSeparator(),
			nil,
			nil,
			widget.NewSeparator(),
			container.NewVBox(
				totalToRefundItem,
				container.NewHBox(
					layout.NewSpacer(),
					container.NewVBox(
						layout.NewSpacer(),
						widget.NewSeparator(),
						totalCapitalItem,
						widget.NewSeparator(),
						totalInterestItem,
						widget.NewSeparator(),
						layout.NewSpacer(),
					),
					totalPaymentGraph,
					layout.NewSpacer(),
				),
			),
		)

		// =======================================================================================
		// Bottom right box
		remainingCapitalItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Outstanding capital"), helper.ValueSpacer(fmt.Sprintf("%0.2f", remainingCapital))))
		remainingCapitalItem.Alignment = fyne.TextAlignCenter
		remainingCapitalItem.SizeName = theme.SizeNameSubHeadingText

		remainingCapitalProgressItem := widget.NewProgressBar()
		remainingCapitalProgressItem.TextFormatter = func() string {
			return fmt.Sprintf(
				"%s %.2f%% %s",
				lang.L("You have refunded"),
				remainingCapitalProgressItem.Value*100, lang.L("Of the capital"),
			)
		}
		remainingCapitalProgressItem.SetValue(1 - remainingCapital/loans[id.Row].Total_amount)

		bottomRightBox := container.NewBorder(
			widget.NewSeparator(),
			nil,
			widget.NewSeparator(),
			nil,
			container.NewVBox(
				remainingCapitalItem,
				layout.NewSpacer(),
				remainingCapitalProgressItem,
				layout.NewSpacer(),
			),
		)

		// =======================================================================================
		// Bottom mid box

		currentPaymentGraph := helper.DrawDoughnut(
			[]string{lang.L("Capital"), lang.L("Interests")},
			[]float64{loans[id.Row].Total_amount - remainingCapital, sumPeriodInterest},
			fyne.NewSize(120, 120),
			"Current period payment",
		)

		totalRefundedItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total refunded"), helper.ValueSpacer(fmt.Sprintf("%0.2f", float64(loans[id.Row].Nb_payments_done)*(loans[id.Row].Next_payment_amount-loans[id.Row].Insurance_amount)))))
		totalRefundedItem.Alignment = fyne.TextAlignCenter
		totalRefundedItem.SizeName = theme.SizeNameSubHeadingText

		capitalRefundedItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Capital"), helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id.Row].Total_amount-remainingCapital))))
		capitalRefundedItem.Alignment = fyne.TextAlignCenter
		capitalRefundedItem.SizeName = theme.SizeNameCaptionText

		interestRefundedItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Interests"), helper.ValueSpacer(fmt.Sprintf("%0.2f", sumPeriodInterest))))
		interestRefundedItem.Alignment = fyne.TextAlignCenter
		interestRefundedItem.SizeName = theme.SizeNameCaptionText

		bottomMidBox := container.NewBorder(
			widget.NewSeparator(),
			nil,
			nil,
			widget.NewSeparator(),
			container.NewVBox(
				totalRefundedItem,
				container.NewHBox(
					layout.NewSpacer(),
					container.NewVBox(
						layout.NewSpacer(),
						widget.NewSeparator(),
						capitalRefundedItem,
						widget.NewSeparator(),
						interestRefundedItem,
						widget.NewSeparator(),
						layout.NewSpacer(),
					),
					currentPaymentGraph,
					layout.NewSpacer(),
				),
			),
		)
		// =======================================================================================

		w.SetContent(container.NewVBox(
			container.NewGridWithColumns(2,
				topLeftBox,
				topRightBox,
			),
			container.NewGridWithColumns(3,
				bottomLeftBox,
				bottomMidBox,
				bottomRightBox,
			),
		))
		w.Show()
	}

	// Reload button reloads data by querying the backend
	reloadButton := widget.NewButton("", func() {
		loans = getLoans(app)
		loanTable.Refresh()

		// Reset header sorting if any
		columnSort[0] = numberOfSorts
		applySort(0, &loanTable.Table, loans)
	})

	reloadButton.Icon = theme.ViewRefreshIcon()

	return container.NewBorder(nil, container.NewBorder(nil, nil, nil, reloadButton, nil), nil, nil, loanTable)
}

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/loan" and retrieve loans
func getLoans(app fyne.App) []Loan {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	url := fmt.Sprintf("%s://%s:%s/loan/", backendProtocol, backendIp, backendPort)
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

	var loans []Loan
	if err := json.Unmarshal(body, &loans); err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot unmarshal loans")
		return nil

	}

	return loans
}

// Sort table data
func applySort(col int, t *widget.Table, data []Loan) {

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
			return a.Loan_account_id < b.Loan_account_id
		}

		switch col {
		case subscriptionDateColumn:
			if order == sortAsc {
				return a.Subscription_date < b.Subscription_date
			}
			return a.Subscription_date > b.Subscription_date

		case valueColumn:
			if order == sortAsc {
				return a.Total_amount > b.Total_amount
			}
			return a.Total_amount < b.Total_amount

		case durationColumn:
			if order == sortDesc {
				return a.Duration > b.Duration
			}
			return a.Duration < b.Duration

		default:
			return false
		}
	})
	t.Refresh()
}
