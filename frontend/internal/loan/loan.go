package loan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-analyze/charts"
)

const (
	unselectTime = 200 * time.Millisecond
)

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

// Create the main view for loans
func NewLoanScreen(app fyne.App, win fyne.Window) *fyne.Container {

	loanTable := createLoanTable(app)

	return loanTable
}

// Create the table of of loan
func createLoanTable(app fyne.App) *fyne.Container {

	subscriptionDateLabel := widget.NewLabel(lang.L("Subscription date"))
	subscriptionDateLabel.Alignment = fyne.TextAlignCenter
	subscriptionDateLabel.TextStyle.Bold = true

	valueHeaderLabel := widget.NewLabel(lang.L("Value"))
	valueHeaderLabel.Alignment = fyne.TextAlignCenter
	valueHeaderLabel.TextStyle.Bold = true

	durationHeaderLabel := widget.NewLabel(lang.L("Duration"))
	durationHeaderLabel.Alignment = fyne.TextAlignCenter
	durationHeaderLabel.TextStyle.Bold = true

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
			if loans[i].Subscription_date != "" { // Parse the date and keep only YYYY-MM-DD
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

	// Display additional details when a loan is clicked on
	loanTable.OnSelected = func(id widget.ListItemID) {

		go func() { // Unselect the cell after some time
			time.Sleep(unselectTime)
			fyne.Do(func() {
				loanTable.Unselect(id)
			})
		}()

		w := app.NewWindow(fmt.Sprintf("%s : %d", lang.L("Loan"), id))
		w.CenterOnScreen()

		// ToDo: to be moved in backend
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

		// Calculate the interest and capital reimbursed for the current (n+1) mensuality
		remainingCapital := loans[id].Total_amount

		var periodInterest float64    // The interest amount for the current (n+1) mensuality
		var sumPeriodInterest float64 // The sum of interests paid for this loan at the moment
		var periodCapital float64

		for j := range totalNbPayments {

			// Loop until we reach the current mensuality, ie today
			if totalNbPayments-j == int(loans[id].Nb_payments_left) {
				break
			}

			periodInterest = loans[id].Rate / 100 * float64(remainingCapital) / 12
			periodCapital = loans[id].Next_payment_amount - periodInterest

			sumPeriodInterest += periodInterest
			remainingCapital = remainingCapital - periodCapital
		}

		// =======================================================================================
		// Top Left Box
		loanTypeItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Type"), lang.L(loans[id].Loan_type)))
		loanTypeItem.Alignment = fyne.TextAlignCenter
		loanTypeItem.SizeName = theme.SizeNameSubHeadingText

		amountItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Amount"), helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id].Total_amount))))
		amountItem.Alignment = fyne.TextAlignCenter
		amountItem.SizeName = theme.SizeNameSubHeadingText

		durationItem := widget.NewLabel(fmt.Sprintf("%s: %d", lang.L("Duration"), totalNbPayments))
		durationItem.Alignment = fyne.TextAlignCenter
		durationItem.SizeName = theme.SizeNameSubHeadingText

		rateItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f %%", lang.L("Rate"), loans[id].Rate))
		rateItem.Alignment = fyne.TextAlignCenter
		rateItem.SizeName = theme.SizeNameSubHeadingText

		mensualitiesLeftItem := widget.NewLabel(fmt.Sprintf("%s: %d", lang.L("Mensualities left"), loans[id].Nb_payments_left))
		mensualitiesLeftItem.Alignment = fyne.TextAlignCenter

		mensualitiesPaidItem := widget.NewLabel(fmt.Sprintf("%s: %d", lang.L("Mensualities paid"), loans[id].Nb_payments_done))
		mensualitiesPaidItem.Alignment = fyne.TextAlignCenter

		var endingDateItem *widget.Label
		if loans[id].Maturity_date != "" {
			endingDate, err := time.Parse("2006-01-02 15:04:05", loans[id].Maturity_date)
			if err != nil {
				helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", loans[id].Maturity_date)
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
		nextPeriodPaymentGraph := drawDoughnut(
			[]string{lang.L("Capital"), lang.L("Insurance"), lang.L("Interests")},
			[]float64{periodCapital, loans[id].Insurance_amount, periodInterest},
			fyne.NewSize(120, 120),
			"Next period payment",
		)

		mensualityItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Next mensuality"), loans[id].Next_payment_amount))
		mensualityItem.Alignment = fyne.TextAlignCenter
		mensualityItem.SizeName = theme.SizeNameSubHeadingText

		periodCapitalItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Capital"), periodCapital))
		periodCapitalItem.Alignment = fyne.TextAlignCenter
		periodCapitalItem.SizeName = theme.SizeNameCaptionText

		periodInterestItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Interests"), periodInterest))
		periodInterestItem.Alignment = fyne.TextAlignCenter

		periodInterestItem.SizeName = theme.SizeNameCaptionText

		periodInsuranceItem := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Insurance"), loans[id].Insurance_amount))
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
		totalToRefund := loans[id].Next_payment_amount * float64(totalNbPayments)
		paidInterest := totalToRefund - loans[id].Total_amount

		totalPaymentGraph := drawDoughnut(
			[]string{lang.L("Capital"), lang.L("Interests")},
			[]float64{loans[id].Total_amount, paidInterest},
			fyne.NewSize(120, 120),
			"Next period payment",
		)

		totalToRefundItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total loan cost"), helper.ValueSpacer(fmt.Sprintf("%0.2f", totalToRefund))))
		totalToRefundItem.Alignment = fyne.TextAlignCenter
		totalToRefundItem.SizeName = theme.SizeNameSubHeadingText

		totalCapitalItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Capital"), helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id].Total_amount))))
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
		remainingCapitalProgressItem.SetValue(1 - remainingCapital/loans[id].Total_amount)

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

		currentPaymentGraph := drawDoughnut(
			[]string{lang.L("Capital"), lang.L("Interests")},
			[]float64{loans[id].Total_amount - remainingCapital, sumPeriodInterest},
			fyne.NewSize(120, 120),
			"Current period payment",
		)

		totalRefundedItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total refunded"), helper.ValueSpacer(fmt.Sprintf("%0.2f", float64(totalNbPayments-int(loans[id].Nb_payments_left))*loans[id].Next_payment_amount))))
		totalRefundedItem.Alignment = fyne.TextAlignCenter
		totalRefundedItem.SizeName = theme.SizeNameSubHeadingText

		capitalRefundedItem := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Capital"), helper.ValueSpacer(fmt.Sprintf("%0.2f", loans[id].Total_amount-remainingCapital))))
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

	return container.NewBorder(container.NewVBox(headerContainer, widget.NewSeparator()), nil, nil, nil, loanTable)
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

// This function creates an doughnut graph image from the specified data
func drawDoughnut(xData []string, yData []float64, size fyne.Size, name string) *canvas.Image {

	var finalXData []string
	var finalYData []float64

	// Remove incorrect values from data set
	for index, element := range yData {

		if element > 0 {
			finalXData = append(finalXData, xData[index])
			finalYData = append(finalYData, element)

		}
	}

	opt := charts.NewDoughnutChartOptionWithData(finalYData)

	opt.Theme = charts.GetTheme(charts.ThemeSummer).WithBackgroundColor(charts.ColorTransparent)

	opt.Legend = charts.LegendOption{
		SeriesNames: finalXData,
		Show:        charts.Ptr(false),
	}

	fontSize := 30
	opt.CenterValues = "labels"
	opt.CenterValuesFontStyle = charts.NewFontStyleWithSize(float64(fontSize))

	p := charts.NewPainter(charts.PainterOptions{
		OutputFormat: charts.ChartOutputPNG,
		Width:        15 * fontSize,
		Height:       15 * fontSize,
	})
	err := p.DoughnutChart(opt)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot create doughnut chart")
		return nil
	}
	buf, err := p.Bytes()
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot convert doughnut chart to bytes")
		return nil
	}
	image := canvas.NewImageFromReader(bytes.NewReader(buf), name)
	image.SetMinSize(size)
	image.FillMode = canvas.ImageFillContain

	return image
}
