package helper

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"github.com/go-analyze/charts"
	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

const (
	LogLevelDefault    = "info"
	PreferenceLogLevel = "currentLogLevel"
)

func InitLogger(app fyne.App) {
	logFilePath := filepath.Join(app.Storage().RootURI().Path(), app.Metadata().Name+".log")
	runLogFile, _ := os.OpenFile(
		logFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	stdOut := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}
	multi := zerolog.MultiLevelWriter(stdOut, runLogFile)

	Logger = zerolog.New(multi).With().Timestamp().Logger()
}

var logLevelName2Level = map[string]zerolog.Level{
	"trace": zerolog.TraceLevel,
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
	"panic": zerolog.PanicLevel,
}

func LogLevelNames() []string {
	x := slices.Collect(maps.Keys(logLevelName2Level))
	return x
}

func SetLogLevel(logLevel string, app fyne.App) {
	switch logLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		Logger.Fatal().Msgf("Unsupported value '%s' for log level. Should be trace, debug, info, warn, error, fatal or panic", logLevel)
	}
	app.Preferences().SetString(PreferenceLogLevel, logLevel)
	Logger.Info().Msgf("Log level set to %s", logLevel)
}

// Watch events
func LogLifecycle(app fyne.App) {
	app.Lifecycle().SetOnStarted(func() {
		Logger.Trace().Msg("Started application")
	})
	app.Lifecycle().SetOnStopped(func() {
		Logger.Trace().Msg("Stopped application")
	})
	app.Lifecycle().SetOnEnteredForeground(func() {
		Logger.Trace().Msg("Entered foreground")
	})
	app.Lifecycle().SetOnExitedForeground(func() {
		Logger.Trace().Msg("Exited foreground")
	})
}

// Set system tray if desktop (mini icon like wifi, shield, notifs, etc...)
func MakeTray(app fyne.App, win fyne.Window) {
	if desk, isDesktop := app.(desktop.App); isDesktop {
		comebackItem := fyne.NewMenuItem("Open app", func() {})
		comebackItem.Icon = theme.HomeIcon()
		comebackItem.Action = func() {
			win.Show()
			Logger.Trace().Msg("Going back to main menu with system tray")
		}
		appName := app.Metadata().Name
		titleItem := fyne.NewMenuItem(appName, nil)
		titleItem.Disabled = true
		menu := fyne.NewMenu("SystemTrayMenu",
			titleItem,
			fyne.NewMenuItemSeparator(),
			comebackItem,
		)
		desk.SetSystemTrayMenu(menu)
	}
}

// Center align the objectToAdd by adding 2 spacers. To be used with an horizontal box
func AddHAligned(object fyne.CanvasObject, objectToAdd fyne.CanvasObject) {
	object.(*fyne.Container).Add(layout.NewSpacer())
	object.(*fyne.Container).Add(objectToAdd)
	object.(*fyne.Container).Add(layout.NewSpacer())
}

// Add spacing to value to make it more easily readable
// Ex: From 123456 to 123 456
func IntValueSpacer(value string) string {

	if len(value) <= 4 {
		return value
	}

	var modifiedValue string
	for pos, char := range reverse(value) {
		if pos%3 == 0 && pos >= 3 {
			modifiedValue = modifiedValue + " "
		}
		modifiedValue = modifiedValue + string(char)
	}
	return reverse(modifiedValue)
}

// Add spacing to value to make it more easily readable
// Ex: From 123456.78 to 123 456.78
func ValueSpacer(value string) string {

	if len(value) < 7 {
		return value
	}

	var modifiedValue string
	for pos, char := range reverse(value) {
		if pos%3 == 0 && pos > 5 {
			modifiedValue = modifiedValue + " "
		}
		modifiedValue = modifiedValue + string(char)
	}
	return reverse(modifiedValue)
}

// reverse a string: from "abc" to "cba"
func reverse(s string) string {
	rns := []rune(s)
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		rns[i], rns[j] = rns[j], rns[i]
	}

	return string(rns)
}

// This function creates an doughnut graph image from the specified data
func DrawDoughnut(xData []string, yData []float64, minSize fyne.Size, name string) *canvas.Image {

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

	// Deactivate legends if needed
	// for i := range opt.SeriesList {
	// 	opt.SeriesList[i].Label.Show = charts.Ptr(false)
	// }

	fontSize := 20
	opt.CenterValues = "labels"
	opt.CenterValuesFontStyle = charts.NewFontStyleWithSize(float64(fontSize))

	p := charts.NewPainter(charts.PainterOptions{
		OutputFormat: charts.ChartOutputPNG,
		Width:        450,
		Height:       450,
	})
	err := p.DoughnutChart(opt)
	if err != nil {
		Logger.Error().Err(err).Msg("Cannot create doughnut chart")
		return nil
	}
	buf, err := p.Bytes()
	if err != nil {
		Logger.Error().Err(err).Msg("Cannot convert doughnut chart to bytes")
		return nil
	}
	image := canvas.NewImageFromReader(bytes.NewReader(buf), name)
	image.SetMinSize(minSize)
	image.FillMode = canvas.ImageFillContain

	return image
}
