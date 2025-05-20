package tools

import (
	"fmt"
	"freenahiFront/internal/helper"
	"log"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	simpleInterestType int = iota
	compoundInterestType
	loanType
)

// Create the "Tools" view
func NewToolsScreen(app fyne.App, win fyne.Window) *container.AppTabs {

	tabs := container.NewAppTabs(
		container.NewTabItem(lang.L("Simple interest"), createViewContainer(simpleInterestType)),
		container.NewTabItem(lang.L("Compound interest"), createViewContainer(compoundInterestType)),
		container.NewTabItem(lang.L("Amortizable loan"), createViewContainer(loanType)),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

// Create the table of interests for simple interest
func createSimpleCapitalTable(rate float64, duration int, capital int) *fyne.Container {

	mainContainer := container.NewVBox()

	for i := range duration + 1 {

		durationLabel := widget.NewLabel(strconv.Itoa(i))
		durationLabel.Alignment = fyne.TextAlignCenter

		// V = C + C*r*n
		value := int(math.Round(float64(capital) + float64(capital)*rate/100*float64(i)))
		valueLabel := widget.NewLabel(helper.IntValueSpacer(strconv.Itoa(value)))
		valueLabel.Alignment = fyne.TextAlignCenter

		mainContainer.Add(container.NewGridWithColumns(2, durationLabel, valueLabel))
	}

	return mainContainer
}

// Create the table of interests for compound interest
func createCompoundCapitalTable(rate float64, duration int, capital int) *fyne.Container {

	mainContainer := container.NewVBox()

	for i := range duration + 1 {
		durationLabel := widget.NewLabel(strconv.Itoa(i))
		durationLabel.Alignment = fyne.TextAlignCenter

		// V = C * (1+r)^n
		value := int(math.Round(float64(capital) * math.Pow(1+rate/100, float64(i))))
		valueLabel := widget.NewLabel(helper.IntValueSpacer(strconv.Itoa(value)))
		valueLabel.Alignment = fyne.TextAlignCenter

		var profitLabel *widget.Label
		if i == 0 {
			profitLabel = widget.NewLabel("0")
		} else {
			previousValue := int(math.Round(float64(capital) * math.Pow(1+rate/100, float64(i-1))))

			profitLabel = widget.NewLabel(helper.IntValueSpacer(strconv.Itoa(value - previousValue)))
		}
		profitLabel.Alignment = fyne.TextAlignCenter

		mainContainer.Add(container.NewGridWithColumns(3, durationLabel, valueLabel, profitLabel))
	}

	return mainContainer
}

// Create the table of costs for an amortizable loan
func createLoanTable(rate float64, duration int, capital int, insuranceType string) *fyne.Container {

	mainContainer := container.NewVBox()

	// m = [(C*r)/12] / [1-(1+(r/12))^-n]
	mensuality := ((float64(capital) * (rate / 100.0)) / 12.0) / (1 - math.Pow(1+(rate/100.0)/12.0, -float64(duration)*12))
	remainingCapital := float64(capital)

	for i := range duration * 12 {
		dueDateLabel := widget.NewLabel(strconv.Itoa(i + 1))
		dueDateLabel.Alignment = fyne.TextAlignCenter

		periodInterest := rate / 100 * float64(remainingCapital) / 12
		periodInterestLabel := widget.NewLabel(fmt.Sprintf("%0.2f", periodInterest))
		periodInterestLabel.Alignment = fyne.TextAlignCenter

		periodCapital := mensuality - periodInterest
		periodCapitalLabel := widget.NewLabel(fmt.Sprintf("%0.2f", periodCapital))
		periodCapitalLabel.Alignment = fyne.TextAlignCenter

		var interestCapitalLabel *widget.Label
		switch insuranceType {
		case lang.L("Outstanding capital"):
			interestCapitalLabel = widget.NewLabel("WIP")
		case lang.L("Initial capital"):
			interestCapitalLabel = widget.NewLabel("WIIIIP")
		}

		remainingCapital = remainingCapital - periodCapital

		mainContainer.Add(container.NewGridWithColumns(4,
			dueDateLabel,
			periodInterestLabel,
			periodCapitalLabel,
			interestCapitalLabel,
		))

	}

	return mainContainer
}

func createViewContainer(toolType int) *fyne.Container {

	const (
		defaultRateValue = 2.0
		minRate          = -10
		maxRate          = 20
		stepRate         = 0.1

		defaultDurationValue = 5
		minDuration          = 1
		maxDuration          = 50
		stepDuration         = 1

		defaultCapitalValue = 10000
		minCapital          = 0
		stepCapital         = 1000

		defaultInsuranceRateValue = 2.0
		minInsuranceRate          = -10
		maxInsuranceRate          = 20
		stepInsuranceRate         = 0.1
	)

	// ========================================================================
	// Rate
	rate := defaultRateValue
	rateData := binding.BindFloat(&rate)

	rateLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(rateData, lang.L("Capital interest rate")+": %0.2f %%"))
	rateLabel.Alignment = fyne.TextAlignCenter

	rateEntry := widget.NewEntryWithData(binding.FloatToString(rateData))

	rateSlide := widget.NewSliderWithData(minRate, maxRate, rateData)
	rateSlide.Step = stepRate

	rateIncreaseDecreaseButtons := container.NewGridWithColumns(2,
		widget.NewButton(fmt.Sprintf("- %0.2f %%", stepRate), func() {
			value, _ := rateData.Get()
			rateData.Set(value - rateSlide.Step)
		}),
		widget.NewButton(fmt.Sprintf("+ %0.2f %%", stepRate), func() {
			value, _ := rateData.Get()
			rateData.Set(value + stepRate)
		}),
	)

	rateContainer := container.NewVBox(
		container.NewGridWithColumns(3, rateLabel, rateEntry, rateIncreaseDecreaseButtons),
		rateSlide,
	)

	// ========================================================================
	// Duration
	duration := defaultDurationValue
	durationData := binding.BindInt(&duration)

	durationLabel := widget.NewLabelWithData(binding.IntToStringWithFormat(durationData, lang.L("Duration")+": %d "+lang.L("Years")))
	durationLabel.Alignment = fyne.TextAlignCenter

	durationEntry := widget.NewEntryWithData(binding.IntToString(durationData))

	durationIncreaseDecreaseButtons := container.NewGridWithColumns(2,
		widget.NewButton(fmt.Sprintf("- %d", stepDuration), func() {
			value, _ := durationData.Get()
			durationData.Set(value - stepDuration)
		}),
		widget.NewButton(fmt.Sprintf("+ %d", stepDuration), func() {
			value, _ := durationData.Get()
			durationData.Set(value + stepDuration)
		}),
	)

	durationContainer := container.NewGridWithColumns(3, durationLabel, durationEntry, durationIncreaseDecreaseButtons)

	// ========================================================================
	// Capital
	capital := defaultCapitalValue
	capitalData := binding.BindInt(&capital)

	var capitalLabel *widget.Label
	switch toolType {
	case simpleInterestType, compoundInterestType:
		capitalLabel = widget.NewLabelWithData(binding.IntToStringWithFormat(capitalData, lang.L("Capital")+": %d €"))

	case loanType:
		capitalLabel = widget.NewLabelWithData(binding.IntToStringWithFormat(capitalData, lang.L("Borrowed capital")+": %d €"))
	}

	capitalLabel.Alignment = fyne.TextAlignCenter

	capitalEntry := widget.NewEntryWithData(binding.IntToString(capitalData))

	capitalIncreaseDecreaseButtons := container.NewGridWithColumns(2,
		widget.NewButton(fmt.Sprintf("- %d", stepCapital), func() {
			value, _ := capitalData.Get()
			capitalData.Set(value - stepCapital)

		}),
		widget.NewButton(fmt.Sprintf("+ %d", stepCapital), func() {
			value, _ := capitalData.Get()
			capitalData.Set(value + stepCapital)
		}),
	)

	capitalContainer := container.NewGridWithColumns(3, capitalLabel, capitalEntry, capitalIncreaseDecreaseButtons)
	dataContainer := container.NewVBox(
		rateContainer,
		durationContainer,
		capitalContainer,
	)

	insuranceRate := defaultInsuranceRateValue
	insuranceRateData := binding.BindFloat(&insuranceRate)

	insuranceType := lang.L("Outstanding capital")
	insuranceTypeData := binding.BindString(&insuranceType)

	if toolType == loanType {

		insuranceRateLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(insuranceRateData, lang.L("Insurance interest rate")+": %0.2f"))
		insuranceRateLabel.Alignment = fyne.TextAlignCenter

		insuranceRateEntry := widget.NewEntryWithData(binding.FloatToString(insuranceRateData))

		insuranceRateSlide := widget.NewSliderWithData(minInsuranceRate, maxInsuranceRate, insuranceRateData)
		insuranceRateSlide.Step = stepInsuranceRate

		insuranceRateIncreaseDecreaseButtons := container.NewGridWithColumns(2,
			widget.NewButton(fmt.Sprintf("- %0.2f %%", stepInsuranceRate), func() {
				value, _ := insuranceRateData.Get()
				insuranceRateData.Set(value - insuranceRateSlide.Step)
			}),
			widget.NewButton(fmt.Sprintf("+ %0.2f %%", stepInsuranceRate), func() {
				value, _ := insuranceRateData.Get()
				insuranceRateData.Set(value + stepInsuranceRate)
			}),
		)

		radio := widget.NewRadioGroup([]string{lang.L("Outstanding capital"), lang.L("Initial capital")}, func(value string) {
			log.Println("Radio set to", value)
		})
		radio.Horizontal = true
		radio.Selected = lang.L("Outstanding capital")
		radio.OnChanged = func(value string) {
			switch value {
			case "":
				radio.Selected = lang.L("Outstanding capital")
				insuranceTypeData.Set(lang.L("Outstanding capital"))
			case lang.L("Outstanding capital"):
				insuranceTypeData.Set(lang.L("Outstanding capital"))
			case lang.L("Initial capital"):
				insuranceTypeData.Set(lang.L("Initial capital"))
			}
		}

		insuranceTypeLabel := widget.NewLabel(fmt.Sprintf("%s:", lang.L("Insurance type")))
		insuranceTypeLabel.Alignment = fyne.TextAlignCenter

		insuranceContainer := container.NewVBox(
			container.NewGridWithColumns(2, insuranceTypeLabel, radio),
			container.NewGridWithColumns(3, insuranceRateLabel, insuranceRateEntry, insuranceRateIncreaseDecreaseButtons),
			insuranceRateSlide,
		)

		dataContainer.Add(insuranceContainer)
	}

	// ========================================================================
	// Result

	var explanation *widget.Label
	var headerContainer *fyne.Container

	switch toolType {
	case simpleInterestType:
		durationHeaderLabel := widget.NewLabel(lang.L("Duration"))
		durationHeaderLabel.Alignment = fyne.TextAlignCenter
		durationHeaderLabel.TextStyle.Italic = true

		valueHeaderLabel := widget.NewLabel(lang.L("Value"))
		valueHeaderLabel.Alignment = fyne.TextAlignCenter
		valueHeaderLabel.TextStyle.Italic = true

		explanation = widget.NewLabel(lang.L("Simple interest explanation"))

		headerContainer = container.NewVBox(container.NewGridWithColumns(
			2,
			durationHeaderLabel,
			valueHeaderLabel,
		))

	case compoundInterestType:
		durationHeaderLabel := widget.NewLabel(lang.L("Duration"))
		durationHeaderLabel.Alignment = fyne.TextAlignCenter
		durationHeaderLabel.TextStyle.Italic = true

		valueHeaderLabel := widget.NewLabel(lang.L("Value"))
		valueHeaderLabel.Alignment = fyne.TextAlignCenter
		valueHeaderLabel.TextStyle.Italic = true

		explanation = widget.NewLabel(lang.L("Compound interest explanation"))

		profitHeaderLabel := widget.NewLabel(lang.L("Profit"))
		profitHeaderLabel.Alignment = fyne.TextAlignCenter
		profitHeaderLabel.TextStyle.Italic = true

		headerContainer = container.NewVBox(container.NewGridWithColumns(
			3,
			durationHeaderLabel,
			valueHeaderLabel,
			profitHeaderLabel,
		))

	case loanType:
		explanation = widget.NewLabel(lang.L("Amortizable loan explanation"))

		dueDateHeaderLabel := widget.NewLabel(lang.L("Due date"))
		dueDateHeaderLabel.Alignment = fyne.TextAlignCenter
		dueDateHeaderLabel.TextStyle.Italic = true

		interestHeaderLabel := widget.NewLabel(lang.L("Interests"))
		interestHeaderLabel.Alignment = fyne.TextAlignCenter
		interestHeaderLabel.TextStyle.Italic = true

		capitalHeaderLabel := widget.NewLabel(lang.L("Capital"))
		capitalHeaderLabel.Alignment = fyne.TextAlignCenter
		capitalHeaderLabel.TextStyle.Italic = true

		insuranceHeaderLabel := widget.NewLabel(lang.L("Insurance"))
		insuranceHeaderLabel.Alignment = fyne.TextAlignCenter
		insuranceHeaderLabel.TextStyle.Italic = true

		headerContainer = container.NewVBox(container.NewGridWithColumns(
			4,
			dueDateHeaderLabel,
			interestHeaderLabel,
			capitalHeaderLabel,
			insuranceHeaderLabel,
		))
	}

	capitalTableContainer := container.NewScroll(container.NewHBox())
	capitalTableContainer.SetMinSize(fyne.NewSize(capitalTableContainer.MinSize().Width, 100))

	resultContainer := container.NewVBox()

	// This function is called when there is a change in the rate, duration, or capital
	// => we update displayed results when the user modifies some values
	dataListener := binding.NewDataListener(func() {
		rate, _ := rateData.Get()
		duration, _ := durationData.Get()
		capital, _ := capitalData.Get()

		// Replace incorrect values if needed
		if rate < minRate {
			rateData.Set(minRate)
		} else if rate > maxRate {
			rateData.Set(maxRate)
		}

		if duration < minDuration {
			durationData.Set(minDuration)
		} else if duration > maxDuration {
			durationData.Set(maxDuration)
		}

		if capital < minCapital {
			capitalData.Set(minCapital)
		}

		var result int

		// Recreate the table with new data
		switch toolType {
		case simpleInterestType, compoundInterestType:
			switch toolType {
			case simpleInterestType:
				capitalTableContainer.Content = createSimpleCapitalTable(rate, duration, capital)
				result = int(math.Round(float64(capital) + float64(capital)*rate/100*float64(duration)))
			case compoundInterestType:
				capitalTableContainer.Content = createCompoundCapitalTable(rate, duration, capital)
				result = int(math.Round(float64(capital) * math.Pow(1+rate/100, float64(duration))))
			}

			resultLabel := widget.NewLabel(
				fmt.Sprintf("%s: %s", lang.L("Final capital"), helper.IntValueSpacer(fmt.Sprintf("%d", result))),
			)
			resultLabel.TextStyle.Bold = true
			resultLabel.SizeName = theme.SizeNameHeadingText
			resultLabel.Alignment = fyne.TextAlignCenter

			multiplierLabel := widget.NewLabel(
				fmt.Sprintf("%s: %0.2f %%", lang.L("Multiplier"), (float64(result)/float64(capital)-1)*100),
			)
			multiplierLabel.TextStyle.Bold = true
			multiplierLabel.Alignment = fyne.TextAlignCenter
			multiplierLabel.Importance = widget.SuccessImportance

			resultContainer.RemoveAll()
			resultContainer.Add(container.NewVBox(resultLabel, multiplierLabel))

		case loanType:

			capitalTableContainer.Content = createLoanTable(rate, duration, capital, insuranceType)

			// m = [(C*r)/12] / [1-(1+(r/12))^-n]
			mensuality := ((float64(capital) * (rate / 100.0)) / 12.0) / (1 - math.Pow(1+(rate/100.0)/12.0, -float64(duration)*12))
			mensualityLabel := widget.NewLabel(fmt.Sprintf("%s: %0.2f", lang.L("Mensuality"), mensuality))
			mensualityLabel.TextStyle.Bold = true
			mensualityLabel.Alignment = fyne.TextAlignCenter

			totalRefunded := helper.ValueSpacer(fmt.Sprintf("%0.2f", mensuality*12.0*float64(duration)))
			totalRefundedLabel := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Total refunded"), totalRefunded))
			totalRefundedLabel.TextStyle.Bold = true
			totalRefundedLabel.SizeName = theme.SizeNameHeadingText
			totalRefundedLabel.Alignment = fyne.TextAlignCenter

			paidInterest := mensuality*12.0*float64(duration) - float64(capital)
			paidInterestFormatted := helper.ValueSpacer(fmt.Sprintf("%0.2f", paidInterest))

			paidInterestLabel := widget.NewLabel(fmt.Sprintf("%s: %s", lang.L("Interest paid"), paidInterestFormatted))
			paidInterestLabel.TextStyle.Bold = true
			paidInterestLabel.SizeName = theme.SizeNameHeadingText
			paidInterestLabel.Alignment = fyne.TextAlignCenter

			loanCostLabel := widget.NewLabel(fmt.Sprintf("%s: %0.2f %%", lang.L("Loan cost"), 100*paidInterest/float64(capital)))
			loanCostLabel.TextStyle.Bold = true
			loanCostLabel.Alignment = fyne.TextAlignCenter

			resultContainer.RemoveAll()
			resultContainer.Add(container.NewVBox(totalRefundedLabel, paidInterestLabel, loanCostLabel, mensualityLabel))

		}

		capitalTableContainer.Refresh()

	})

	// Add the listener function to the binded data
	rateData.AddListener(dataListener)
	durationData.AddListener(dataListener)
	capitalData.AddListener(dataListener)
	insuranceRateData.AddListener(dataListener)
	insuranceTypeData.AddListener(dataListener)

	return container.NewVBox(
		dataContainer,
		widget.NewSeparator(),
		layout.NewSpacer(),
		resultContainer,
		layout.NewSpacer(),
		widget.NewSeparator(),
		headerContainer,
		capitalTableContainer,
		widget.NewSeparator(),
		container.NewHScroll(explanation),
	)

}
