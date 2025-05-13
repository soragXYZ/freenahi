package tools

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-analyze/charts"
)

func NewToolsScreen(app fyne.App, win fyne.Window) *container.AppTabs {

	tabs := container.NewAppTabs(
		container.NewTabItem(lang.L("Simple interest"), newSimpleInterestView()),
		container.NewTabItem(lang.L("Compound interest"), widget.NewLabel("ok")),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return tabs
}

// To Do: handle incorrect user input
func newSimpleInterestView() *fyne.Container {

	const (
		defaultRateSliderValue = 2.0
		minRate                = -10
		maxRate                = 20
		minDuration            = 1
	)

	// ========================================================================
	// Rate

	rateText := widget.NewLabel(lang.L("Interest rate"))
	rateText.Resize(fyne.NewSize(rateText.MinSize().Width, rateText.MinSize().Height))
	rateText.TextStyle.Italic = true
	rateText.TextStyle.Bold = true

	rate := defaultRateSliderValue
	rateData := binding.BindFloat(&rate)

	rateLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(rateData, lang.L("Interest rate")+": %0.2f"))
	rateLabel.Alignment = fyne.TextAlignCenter

	rateEntry := widget.NewEntryWithData(binding.FloatToString(rateData))

	rateSlide := widget.NewSliderWithData(minRate, maxRate, rateData)
	rateSlide.Step = 0.1

	rateIncreaseDecreaseButtons := container.NewGridWithColumns(2,
		widget.NewButton("- 0.1 %", func() {
			value, _ := rateData.Get()
			if value-0.1 >= minRate {
				rateData.Set(value - 0.1)
			}
		}),
		widget.NewButton("+ 0.1 %", func() {
			value, _ := rateData.Get()
			if value+0.1 <= maxRate {
				rateData.Set(value + 0.1)
			}
		}),
	)

	rateContainer := container.NewVBox(
		container.NewGridWithColumns(3, rateLabel, rateEntry, rateIncreaseDecreaseButtons),
		rateSlide,
	)

	// ========================================================================
	// Duration

	durationText := widget.NewLabel(lang.L("Duration"))
	durationText.Resize(fyne.NewSize(durationText.MinSize().Width, durationText.MinSize().Height))
	durationText.TextStyle.Italic = true
	durationText.TextStyle.Bold = true

	duration := 5
	durationData := binding.BindInt(&duration)

	durationLabel := widget.NewLabelWithData(binding.IntToStringWithFormat(durationData, lang.L("Duration")+": %d années"))
	durationLabel.Alignment = fyne.TextAlignCenter

	durationEntry := widget.NewEntryWithData(binding.IntToString(durationData))

	durationIncreaseDecreaseButtons := container.NewGridWithColumns(2,
		widget.NewButton("- 1", func() {
			value, _ := durationData.Get()
			if value-1 >= minDuration {
				durationData.Set(value - 1)
			}
		}),
		widget.NewButton("+ 1", func() {
			value, _ := durationData.Get()
			durationData.Set(value + 1)
		}),
	)

	durationContainer := container.NewGridWithColumns(3, durationLabel, durationEntry, durationIncreaseDecreaseButtons)

	// ========================================================================
	// Capital

	capitalText := widget.NewLabel(lang.L("Capital"))
	capitalText.Resize(fyne.NewSize(capitalText.MinSize().Width*2, capitalText.MinSize().Height*2))
	capitalText.TextStyle.Italic = true
	capitalText.TextStyle.Bold = true

	capital := 10000
	capitalData := binding.BindInt(&capital)

	capitalLabel := widget.NewLabelWithData(binding.IntToStringWithFormat(capitalData, lang.L("Capital")+": %d €"))
	capitalLabel.Alignment = fyne.TextAlignCenter

	capitalEntry := widget.NewEntryWithData(binding.IntToString(capitalData))

	capitalIncreaseDecreaseButtons := container.NewGridWithColumns(2,
		widget.NewButton("- 1 000", func() {
			value, _ := capitalData.Get()
			capitalData.Set(value - 1000)
		}),
		widget.NewButton("+ 1 000", func() {
			value, _ := capitalData.Get()
			capitalData.Set(value + 1000)
		}),
	)

	capitalContainer := container.NewGridWithColumns(3, capitalLabel, capitalEntry, capitalIncreaseDecreaseButtons)

	// ========================================================================
	// Result

	resultLabel := widget.NewLabel("X")
	resultLabel.TextStyle.Bold = true
	resultLabel.Alignment = fyne.TextAlignCenter
	resultLabel.Resize(fyne.NewSize(300, 300))

	graphBox := container.NewHBox(layout.NewSpacer(), drawLineGraph(rate, duration, capital), layout.NewSpacer())

	// This function is called when there is a change in the rate, duration, or capital
	// => we update the result when the user modifies values
	dataListener := binding.NewDataListener(func() {
		rate, _ = rateData.Get()
		duration, _ = durationData.Get()
		capital, _ = capitalData.Get()

		result := int(math.Round(float64(capital) + float64(capital)*rate/100*float64(duration)))

		resultLabel.SetText(fmt.Sprintf("%s %d", lang.L("Final capital"), result))
		resultLabel.Refresh()

		graphBox.Objects[1] = container.NewHBox(layout.NewSpacer(), drawLineGraph(rate, duration, capital), layout.NewSpacer())
		graphBox.Refresh()
	})

	rateData.AddListener(dataListener)
	durationData.AddListener(dataListener)
	capitalData.AddListener(dataListener)

	return container.NewVBox(
		rateContainer,
		durationContainer,
		capitalContainer,
		widget.NewSeparator(),
		// graphBox,
		resultLabel,
	)
}

// To Do: handle incorrect user input
func drawLineGraph(rate float64, duration int, capital int) *canvas.Image {

	var axisLabel []string
	var yLabel []float64
	var constant []float64

	for i := range duration + 1 {
		axisLabel = append(axisLabel, strconv.Itoa(i))
		yLabel = append(yLabel, float64(capital)+float64(capital)*rate/100*float64(i))
		constant = append(constant, float64(capital))
	}

	values := [][]float64{yLabel, constant}

	opt := charts.NewLineChartOptionWithData(values)
	opt.Title.Text = "Line"
	opt.XAxis.Labels = axisLabel

	// opt.Legend.Padding = charts.Box{
	// 	Top:    5,
	// 	Bottom: 10,
	// }
	opt.YAxis[0].Min = charts.Ptr(values[0][0] / 2) // Ensure y-axis starts at 0

	// Setup fill styling below
	opt.FillArea = charts.Ptr(true) // Enable fill area
	opt.FillOpacity = 150           // Set fill opacity
	// opt.XAxis.BoundaryGap = charts.Ptr(false) // Disable boundary gap

	p := charts.NewPainter(charts.PainterOptions{
		OutputFormat: charts.ChartOutputPNG,
		Width:        600,
		Height:       600,
	})
	if err := p.LineChart(opt); err != nil {
		panic(err)
	}
	buf, err := p.Bytes()
	if err != nil {
		panic(err)
	}

	image := canvas.NewImageFromReader(bytes.NewReader(buf), "image")
	image.FillMode = canvas.ImageFillOriginal

	return image
}
