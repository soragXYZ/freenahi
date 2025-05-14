package tools

import (
	"fmt"
	"freenahiFront/internal/helper"
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
)

// Struct used to store the current valuation of the capital for the given period
type interest struct {
	duration, value int
}

// Create the "Tools" view
func NewToolsScreen(app fyne.App, win fyne.Window) *container.AppTabs {

	tabs := container.NewAppTabs(
		container.NewTabItem(lang.L("Simple interest"), newSimpleInterestView()),
		container.NewTabItem(lang.L("Compound interest"), newCompoundInterestView()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

// Create the "Simple Interest" view, used as a tab
func newSimpleInterestView() *fyne.Container {

	rateData, rateContainer, durationData, durationContainer, capitalData, capitalContainer := createUserInput()

	resultLabel := widget.NewLabel("X")
	resultLabel.TextStyle.Bold = true
	resultLabel.SizeName = theme.SizeNameHeadingText
	resultLabel.Alignment = fyne.TextAlignCenter

	multiplierLabel := widget.NewLabel("X")
	multiplierLabel.TextStyle.Bold = true
	multiplierLabel.Alignment = fyne.TextAlignCenter
	multiplierLabel.Importance = widget.SuccessImportance

	simpleInterestExplanation := widget.NewLabel(lang.L("Simple interest explanation"))

	durationHeaderLabel := widget.NewLabel(lang.L("Duration"))
	durationHeaderLabel.Alignment = fyne.TextAlignCenter

	valueHeaderLabel := widget.NewLabel(lang.L("Value"))
	valueHeaderLabel.Alignment = fyne.TextAlignCenter

	headerContainer := container.NewVBox(container.NewGridWithColumns(2, durationHeaderLabel, valueHeaderLabel))

	capitalTableContainer := container.NewScroll(container.NewHBox())
	capitalTableContainer.SetMinSize(fyne.NewSize(capitalTableContainer.MinSize().Width, 100))

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

		result := int(math.Round(float64(capital) + float64(capital)*rate/100*float64(duration)))
		multiplier := float64(result) / float64(capital)

		var array []interest

		for i := range duration + 1 {
			array = append(array, interest{
				value:    int(math.Round(float64(capital) + float64(capital)*rate/100*float64(i))),
				duration: i,
			})
		}

		// Recreate the table with new data
		capitalTableContainer.Content = createSimpleCapitalTable(array)
		capitalTableContainer.Refresh()

		resultLabel.SetText(fmt.Sprintf("%s : %s", lang.L("Final capital"), helper.IntValueSpacer(fmt.Sprintf("%d", result))))
		resultLabel.Refresh()
		multiplierLabel.SetText(fmt.Sprintf("%s: x%0.3f", lang.L("Multiplier"), multiplier))
		multiplierLabel.Refresh()
	})

	rateData.AddListener(dataListener)
	durationData.AddListener(dataListener)
	capitalData.AddListener(dataListener)

	return container.NewVBox(
		rateContainer,
		durationContainer,
		capitalContainer,
		widget.NewSeparator(),
		layout.NewSpacer(),
		resultLabel,
		multiplierLabel,
		layout.NewSpacer(),
		widget.NewSeparator(),
		headerContainer,
		capitalTableContainer,
		widget.NewSeparator(),
		container.NewHScroll(simpleInterestExplanation),
	)
}

// Create the compound interest view, used as a tab
func newCompoundInterestView() *fyne.Container {

	rateData, rateContainer, durationData, durationContainer, capitalData, capitalContainer := createUserInput()

	resultLabel := widget.NewLabel("X")
	resultLabel.TextStyle.Bold = true
	resultLabel.SizeName = theme.SizeNameHeadingText
	resultLabel.Alignment = fyne.TextAlignCenter

	multiplierLabel := widget.NewLabel("X")
	multiplierLabel.TextStyle.Bold = true
	multiplierLabel.Alignment = fyne.TextAlignCenter
	multiplierLabel.Importance = widget.SuccessImportance

	compoundInterestExplanation := widget.NewLabel(lang.L("Compound interest explanation"))

	durationHeaderLabel := widget.NewLabel(lang.L("Duration"))
	durationHeaderLabel.Alignment = fyne.TextAlignCenter

	valueHeaderLabel := widget.NewLabel(lang.L("Value"))
	valueHeaderLabel.Alignment = fyne.TextAlignCenter

	profitHeaderLabel := widget.NewLabel(lang.L("Profit"))
	profitHeaderLabel.Alignment = fyne.TextAlignCenter

	headerContainer := container.NewVBox(container.NewGridWithColumns(
		3,
		durationHeaderLabel,
		valueHeaderLabel,
		profitHeaderLabel,
	))

	capitalTableContainer := container.NewScroll(container.NewHBox())
	capitalTableContainer.SetMinSize(fyne.NewSize(capitalTableContainer.MinSize().Width, 100))

	// This function is called when there is a change in the rate, duration, or capital
	// => we update the result when the user modifies values
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

		result := int(math.Round(float64(capital) * math.Pow(1+rate/100, float64(duration))))
		multiplier := float64(result) / float64(capital)

		var array []interest

		for i := range duration + 1 {
			array = append(array, interest{
				value:    int(math.Round(float64(capital) * math.Pow(1+rate/100, float64(i)))),
				duration: i,
			})
		}

		// Recreate the table with new data
		capitalTableContainer.Content = createCompoundCapitalTable(array)
		capitalTableContainer.Refresh()

		resultLabel.SetText(fmt.Sprintf("%s : %d", lang.L("Final capital"), result))
		resultLabel.Refresh()
		multiplierLabel.SetText(fmt.Sprintf("%s: x%0.3f", lang.L("Multiplier"), multiplier))
		multiplierLabel.Refresh()
	})

	rateData.AddListener(dataListener)
	durationData.AddListener(dataListener)
	capitalData.AddListener(dataListener)

	return container.NewVBox(
		rateContainer,
		durationContainer,
		capitalContainer,
		widget.NewSeparator(),
		layout.NewSpacer(),
		resultLabel,
		multiplierLabel,
		layout.NewSpacer(),
		widget.NewSeparator(),
		headerContainer,
		capitalTableContainer,
		widget.NewSeparator(),
		container.NewHScroll(compoundInterestExplanation),
	)
}

// Create the table of transaction for simple interest
func createSimpleCapitalTable(array []interest) *fyne.Container {

	mainContainer := container.NewVBox()

	for i := range len(array) {
		durationLabel := widget.NewLabel(strconv.Itoa(array[i].duration))
		durationLabel.Alignment = fyne.TextAlignCenter

		valueLabel := widget.NewLabel(helper.IntValueSpacer(strconv.Itoa(array[i].value)))
		valueLabel.Alignment = fyne.TextAlignCenter

		mainContainer.Add(container.NewGridWithColumns(2, durationLabel, valueLabel))
	}

	return mainContainer
}

// Create the table of transaction for compound interest
func createCompoundCapitalTable(array []interest) *fyne.Container {

	mainContainer := container.NewVBox()

	for i := range len(array) {
		durationLabel := widget.NewLabel(strconv.Itoa(array[i].duration))
		durationLabel.Alignment = fyne.TextAlignCenter

		valueLabel := widget.NewLabel(helper.IntValueSpacer(strconv.Itoa(array[i].value)))
		valueLabel.Alignment = fyne.TextAlignCenter

		var profitLabel *widget.Label
		if i == 0 {
			profitLabel = widget.NewLabel("0")
		} else {
			profitLabel = widget.NewLabel(helper.IntValueSpacer(strconv.Itoa(array[i].value - array[i-1].value)))
		}
		profitLabel.Alignment = fyne.TextAlignCenter

		mainContainer.Add(container.NewGridWithColumns(3, durationLabel, valueLabel, profitLabel))
	}

	return mainContainer
}

func createUserInput() (binding.ExternalFloat, *fyne.Container, binding.ExternalInt, *fyne.Container, binding.ExternalInt, *fyne.Container) {

	// ========================================================================
	// Rate
	rate := defaultRateValue
	rateData := binding.BindFloat(&rate)

	rateLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(rateData, lang.L("Interest rate")+": %0.2f"))
	rateLabel.Alignment = fyne.TextAlignCenter

	rateEntry := widget.NewEntryWithData(binding.FloatToString(rateData))

	rateSlide := widget.NewSliderWithData(minRate, maxRate, rateData)
	rateSlide.Step = stepRate

	rateIncreaseDecreaseButtons := container.NewGridWithColumns(2,
		widget.NewButton(fmt.Sprintf("- %0.2f", stepRate), func() {
			value, _ := rateData.Get()
			rateData.Set(value - rateSlide.Step)
		}),
		widget.NewButton(fmt.Sprintf("+ %0.2f", stepRate), func() {
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

	durationLabel := widget.NewLabelWithData(binding.IntToStringWithFormat(durationData, lang.L("Duration")+": %d années"))
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

	capitalLabel := widget.NewLabelWithData(binding.IntToStringWithFormat(capitalData, lang.L("Capital")+": %d €"))
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

	return rateData, rateContainer, durationData, durationContainer, capitalData, capitalContainer
}
