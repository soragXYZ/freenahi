package loan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

const (
	// Possible value for loan type
	// https://docs.powens.com/api-reference/products/data-aggregation/bank-accounts#loan-object
	simple         = "loan"
	revolving      = "revolvingcredit"
	mortgage       = "mortgage"
	consumercredit = "consumercredit"

	unselectTime = 200 * time.Millisecond
)

type Loan struct {
	Loan_account_id      int     `json:"-"` // absent in base data, field added for simplicity
	Total_amount         float32 `json:"total_amount"`
	Available_amount     float32 `json:"available_amount"`
	Used_amount          float32 `json:"used_amount"`
	Subscription_date    string  `json:"subscription_date"`
	Maturity_date        string  `json:"maturity_date"`
	Start_repayment_date string  `json:"start_repayment_date"`
	Deferred             bool    `json:"deferred"`
	Next_payment_amount  float32 `json:"next_payment_amount"`
	Next_payment_date    string  `json:"next_payment_date"`
	Rate                 float32 `json:"rate"`
	Nb_payments_left     uint    `json:"nb_payments_left"`
	Nb_payments_done     uint    `json:"nb_payments_done"`
	Nb_payments_total    uint    `json:"nb_payments_total"`
	Last_payment_amount  float32 `json:"last_payment_amount"`
	Last_payment_date    string  `json:"last_payment_date"`
	Account_label        string  `json:"account_label"`
	Insurance_label      string  `json:"insurance_label"`
	Insurance_amount     float32 `json:"insurance_amount"`
	Insurance_rate       float32 `json:"insurance_rate"`
	Duration             uint    `json:"duration"`
	Loan_type            string  `json:"type"`
}

func NewLoanScreen(app fyne.App, win fyne.Window) *fyne.Container {

	loanTable := createLoanTable(app)

	return loanTable
}

// Create the table of of loan
func createLoanTable(app fyne.App) *fyne.Container {

	subscriptionDateLabel := widget.NewLabel(lang.L("Subscription date"))
	subscriptionDateLabel.Alignment = fyne.TextAlignCenter

	valueHeaderLabel := widget.NewLabel(lang.L("Value"))
	valueHeaderLabel.Alignment = fyne.TextAlignCenter

	durationHeaderLabel := widget.NewLabel(lang.L("Duration"))
	durationHeaderLabel.Alignment = fyne.TextAlignCenter

	headerContainer := container.NewGridWithColumns(
		3,
		subscriptionDateLabel,
		valueHeaderLabel,
		durationHeaderLabel,
	)

	loans := getLoans(app)

	loanTable := widget.NewList(
		func() int {
			return len(loans)
		},
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(3, widget.NewLabel("Template"), widget.NewLabel("Template"), widget.NewLabel("Template"))
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {

			// Clean the cell from the previous value
			item := o.(*fyne.Container)
			item.RemoveAll()

			var subscriptionDateItem *widget.Label
			if loans[i].Subscription_date != "" {
				parsedSubscriptionDate, err := time.Parse("2006-01-02 15:04:05", loans[i].Subscription_date)
				if err != nil {
					helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", loans[i].Subscription_date)
				}
				subscriptionDateItem = widget.NewLabel(parsedSubscriptionDate.Format("2006-01-02"))
			} else {
				subscriptionDateItem = widget.NewLabel(lang.L("Irrelevant"))
			}

			subscriptionDateItem.Alignment = fyne.TextAlignCenter

			amountItem := widget.NewLabel(helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[i].Total_amount)))
			amountItem.Alignment = fyne.TextAlignCenter

			durationItem := widget.NewLabel(fmt.Sprintf("%d", loans[i].Duration))
			durationItem.Alignment = fyne.TextAlignCenter

			item.Add(subscriptionDateItem)
			item.Add(amountItem)
			item.Add(durationItem)
		},
	)
	loanTable.OnSelected = func(id widget.ListItemID) {

		go func() {
			time.Sleep(unselectTime)
			fyne.Do(func() {
				loanTable.Unselect(id)
			})
		}()

		w := app.NewWindow(fmt.Sprintf("%s : %d", lang.L("Loan"), id))
		w.CenterOnScreen()

		// credit numero | date souscription | type | montant | durée | taux | mensualité | mensualité restantes | montant assurance
		// ajouter cout total crédit, capital restant a rembourser, capital remboursé %
		// https://youtu.be/P3v06IYFu_A?si=hur964b0YJWiKlRJ&t=522

		var totalNbPayments int
		if loans[id].Nb_payments_total == 0 { // Manually calculate how much payments are left to pay

			// Manually calculate duration the ugly way
			t1, err := time.Parse("2006-01-02 15:04:05", loans[id].Maturity_date)
			if err != nil {
				helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", loans[id].Maturity_date)
			}

			t2, err := time.Parse("2006-01-02 15:04:05", loans[id].Subscription_date)
			if err != nil {
				helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", loans[id].Subscription_date)
			}

			yearT1, _ := strconv.Atoi(t1.Format("2006"))
			monthT1, _ := strconv.Atoi(t1.Format("01"))
			yearT2, _ := strconv.Atoi(t2.Format("2006"))
			monthT2, _ := strconv.Atoi(t2.Format("01"))

			totalNbPayments = (yearT1-yearT2)*12 + monthT1 - monthT2

		} else {
			totalNbPayments = int(loans[id].Nb_payments_total)
		}

		// Calculate the credit and capital reimbursed for the current (n+1) period
		remainingCapital := loans[id].Total_amount
		var periodInterest float32
		var periodCapital float32

		for j := range totalNbPayments {

			// Loop until we reach the current period
			if totalNbPayments-j == int(loans[id].Nb_payments_left) {
				break
			}

			periodInterest = loans[id].Rate / 100 * float32(remainingCapital) / 12
			periodCapital = loans[id].Next_payment_amount - periodInterest

			remainingCapital = remainingCapital - periodCapital
		}

		content := container.NewVBox(
			widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Amount"), helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id].Total_amount)))),
			widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Type"), lang.L(loans[id].Loan_type))),
			widget.NewLabel(fmt.Sprintf("%s: %0.2f %%", lang.L("Rate"), loans[id].Rate)),
			widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Mensuality"), loans[id].Next_payment_amount)),
			widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Insurance"), loans[id].Insurance_amount)),
			widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Interests"), periodInterest)),
			widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Capital"), periodCapital)),
		)

		switch loans[id].Loan_type {
		case simple:
			content.Add(widget.NewLabel("simple"))

		case revolving:
			content.Add(widget.NewLabel("revolving"))

		case mortgage:
			content.Add(widget.NewLabel("mortgage"))

		case consumercredit:
			content.Add(widget.NewLabel("consumer"))

		default:
			helper.Logger.Fatal().Msg("Loan type: unsupported type")
		}

		w.SetContent(content)
		w.Resize(fyne.NewSize(800, 800))
		w.Show()
	}

	return container.NewBorder(headerContainer, nil, nil, nil, loanTable)
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
