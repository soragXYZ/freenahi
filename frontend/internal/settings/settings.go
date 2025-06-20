package settings

import (
	"fmt"
	"strconv"
	"time"

	"freenahiFront/internal/helper"
	customTheme "freenahiFront/internal/theme"
	"freenahiFront/resources"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	catppuccin "github.com/mbaklor/fyne-catppuccin"

	kxwidget "github.com/ErikKalkoken/fyne-kx/widget"
)

type settingVariant uint

const (
	settingUndefined settingVariant = iota
	settingText
	settingHeading
	settingSeperator
	settingSwitch
)

const (
	BackendIPDefault    = "localhost"
	PreferenceBackendIP = "currentBackendIP"

	BackendProtocolDefault    = "http"
	BackendProtocolSafe       = "https"
	PreferenceBackendProtocol = "currentBackendProtocol"

	BackendPortDefault    = "8080"
	PreferenceBackendPort = "currentBackendPort"

	BackendPollingIntervalDefault    = 10 // time in seconds
	BackendPollingIntervalMin        = 1
	BackendPollingIntervalMax        = 120
	PreferenceBackendPollingInterval = "currentBackendPollingInterval"

	FullscreenDefault    = false
	PreferenceFullscreen = "currentFullscreen"

	SystemTrayDefault    = false
	PreferenceSystemTray = "currentSystemTray"

	// ThemeDefault    = "Light" // using lang.L("Pastel") but lang not initialized
	PreferenceTheme = "currentTheme"
)

var ThemeDefault = lang.L("Pastel")

type ContextMenuButton struct {
	widget.Button
	menu *fyne.Menu
}

// NewContextMenuButtonWithIcon is an icon button that shows a context menu. The label is optional.
func NewContextMenuButtonWithIcon(icon fyne.Resource, label string, menu *fyne.Menu) *ContextMenuButton {
	b := &ContextMenuButton{menu: menu}
	b.Text = label
	b.Icon = icon

	b.ExtendBaseWidget(b)
	return b
}

// Open a menu when the button is clicked
func (b *ContextMenuButton) Tapped(e *fyne.PointEvent) {
	widget.ShowPopUpMenuAtPosition(
		b.menu,
		fyne.CurrentApp().Driver().CanvasForObject(b),
		e.AbsolutePosition,
	)
}

// SetMenuItems replaces the menu items.
func (b *ContextMenuButton) SetMenuItems(menuItems []*fyne.MenuItem) {
	b.menu.Items = menuItems
	b.menu.Refresh()
}

type SettingAction struct {
	Label  string
	Action func()
}

// SettingItem represents an item in a setting list.
type SettingItem struct {
	Hint   string      // optional hint text
	Label  string      // label
	Getter func() any  // returns the current value for this setting
	Setter func(v any) // sets the value for this setting

	onSelected func(it SettingItem, refresh func()) // action called when selected
	variant    settingVariant
}

func NewSettingItemOptions(
	label, hint string,
	options []string,
	defaultV string,
	getter func() string,
	setter func(v string),
	win fyne.Window,
) SettingItem {
	return SettingItem{
		Label: label,
		Hint:  hint,
		Getter: func() any {
			return getter()
		},
		Setter: func(v any) {
			setter(v.(string))
		},
		onSelected: func(it SettingItem, refresh func()) {
			sel := widget.NewRadioGroup(options, setter)
			sel.SetSelected(it.Getter().(string))
			d := makeSettingDialog(
				sel,
				it.Label,
				it.Hint,
				func() {
					sel.SetSelected(defaultV)
				},
				refresh,
				win,
			)
			d.Show()
		},
		variant: settingText,
	}
}

func makeSettingDialog(
	setting fyne.CanvasObject,
	label, hint string,
	reset, refresh func(),
	w fyne.Window,
) dialog.Dialog {
	var d dialog.Dialog
	buttons := container.NewHBox(
		widget.NewButton(lang.L("Close"), func() {
			d.Hide()
		}),
		layout.NewSpacer(),
		widget.NewButton(lang.L("Reset"), func() {
			reset()
		}),
	)
	c := container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewLabel(hint),
			buttons,
		),
		nil,
		nil,
		setting,
	)

	d = dialog.NewCustomWithoutButtons(label, c, w)
	_, s := w.Canvas().InteractiveArea()
	d.Resize(fyne.NewSize(s.Width*0.8, 100))
	d.SetOnClosed(refresh)
	return d
}

func NewSettingItemHeading(label string) SettingItem {
	return SettingItem{Label: label, variant: settingHeading}
}

// NewSettingItemUserInput creates a user input setting in a setting list.
func NewSettingItemUserInput(
	label, hint, placeholder, regex, regexError string,
	defaultV string,
	getter func() string,
	setter func(v string),
	win fyne.Window,
) SettingItem {
	return SettingItem{
		Label: label,
		Hint:  hint,
		Getter: func() any {
			return getter()
		},
		Setter: func(v any) {
			setter(v.(string))
		},
		onSelected: func(it SettingItem, refresh func()) {

			userEntry := widget.NewEntry()
			userEntry.SetPlaceHolder(placeholder)
			userEntry.Validator = validation.NewRegexp(regex, regexError)

			items := []*widget.FormItem{
				widget.NewFormItem(hint, userEntry),
			}

			_, s := win.Canvas().InteractiveArea()
			d := dialog.NewForm(label, lang.L("Save"), lang.L("Cancel"), items, func(b bool) {
				if !b {
					return
				}
				it.Setter(userEntry.Text)
				refresh()
			}, win)
			d.Resize(fyne.NewSize(s.Width*0.7, 100))
			d.Show()
		},
		variant: settingText,
	}
}

// NewSettingItemSwitch creates a switch setting in a setting list.
func NewSettingItemSwitch(
	label, hint string,
	getter func() bool,
	onChanged func(bool),
) SettingItem {
	return SettingItem{
		Label: label,
		Hint:  hint,
		Getter: func() any {
			return getter()
		},
		Setter: func(v any) {
			onChanged(v.(bool))
		},
		onSelected: func(it SettingItem, refresh func()) {
			it.Setter(!it.Getter().(bool))
			refresh()
		},
		variant: settingSwitch,
	}
}

func NewSettingItemSlider(
	label, hint string,
	minV, maxV, defaultV float64,
	getter func() float64,
	setter func(v float64),
	win fyne.Window,
) SettingItem {
	return SettingItem{
		Label: label,
		Hint:  hint,
		Getter: func() any {
			return getter()
		},
		Setter: func(v any) {
			switch x := v.(type) {
			case float64:
				setter(x)
			case int:
				setter(float64(x))
			default:
				helper.Logger.Fatal().Msg("setting item: unsupported type")
			}
		},
		onSelected: func(it SettingItem, refresh func()) {
			sl := kxwidget.NewSlider(minV, maxV)
			sl.SetValue(float64(getter()))
			sl.OnChangeEnded = setter
			d := makeSettingDialog(
				sl,
				it.Label,
				it.Hint,
				func() {
					sl.SetValue(defaultV)
				},
				refresh,
				win,
			)
			d.Show()
		},
		variant: settingText,
	}
}

// NewSettingList returns a new SettingList widget.
func NewSettingList(items []SettingItem) *widget.List {
	w := &widget.List{}
	w.Length = func() int {
		return len(items)
	}
	w.CreateItem = func() fyne.CanvasObject {
		label := widget.NewLabel("")
		label.Truncation = fyne.TextTruncateClip
		hint := widget.NewLabel("")
		hint.Truncation = fyne.TextTruncateClip
		c := container.NewPadded(container.NewBorder(
			nil,
			container.New(layout.NewCustomPaddedLayout(0, 0, 0, 0), widget.NewSeparator()),
			nil,
			container.NewVBox(
				layout.NewSpacer(),
				container.NewStack(kxwidget.NewSwitch(nil), widget.NewLabel("")),
				layout.NewSpacer(),
			),
			container.New(
				layout.NewCustomPaddedVBoxLayout(0),
				layout.NewSpacer(),
				label, hint, layout.NewSpacer(),
			),
		))
		return c
	}
	w.UpdateItem = func(id widget.ListItemID, co fyne.CanvasObject) {
		if id >= len(items) {
			return
		}
		it := items[id]
		border := co.(*fyne.Container).Objects[0].(*fyne.Container).Objects
		right := border[2].(*fyne.Container).Objects[1].(*fyne.Container).Objects
		sw := right[0].(*kxwidget.Switch)
		value := right[1].(*widget.Label)
		main := border[0].(*fyne.Container).Objects
		hint := main[2].(*widget.Label)
		if it.Hint != "" {
			hint.SetText(it.Hint)
			hint.Show()
		} else {
			hint.Hide()
		}
		label := main[1].(*widget.Label)
		label.Text = it.Label
		label.TextStyle.Bold = false
		switch it.variant {
		case settingHeading:
			label.TextStyle.Bold = true
			value.Hide()
			sw.Hide()
		case settingSwitch:
			value.Hide()
			sw.OnChanged = func(v bool) {
				it.Setter(v)
			}
			sw.On = it.Getter().(bool)
			sw.Show()
			sw.Refresh()
		case settingText:
			value.SetText(fmt.Sprint(it.Getter()))
			value.Show()
			sw.Hide()
		}
		sep := border[1].(*fyne.Container)
		if it.variant == settingSeperator {
			sep.Show()
			value.Hide()
			sw.Hide()
			label.Hide()
		} else {
			sep.Hide()
			label.Show()
			label.Refresh()
		}
		w.SetItemHeight(id, co.(*fyne.Container).MinSize().Height)
	}
	w.OnSelected = func(id widget.ListItemID) {
		if id >= len(items) {
			w.UnselectAll()
			return
		}
		it := items[id]
		if it.onSelected == nil {
			w.Unselect(id)
			return
		}
		it.onSelected(it, func() {
			w.RefreshItem(id)
		})
		go func() {
			time.Sleep(200 * time.Millisecond)
			fyne.Do(func() {
				w.Unselect(id)
			})
		}()
	}
	w.HideSeparators = true
	w.ExtendBaseWidget(w)
	return w
}

// NewSettingItemSeperator creates a seperator in a setting list.
func NewSettingItemSeperator() SettingItem {
	return SettingItem{variant: settingSeperator}
}

func MakeSettingsPage(title string, content fyne.CanvasObject, actions []SettingAction) fyne.CanvasObject {
	t := widget.NewLabel(title)
	t.TextStyle.Bold = true
	items := make([]*fyne.MenuItem, 0)
	for _, action := range actions {
		items = append(items, fyne.NewMenuItem(action.Label, action.Action))
	}
	options := NewContextMenuButtonWithIcon(theme.MoreHorizontalIcon(), lang.L("More"), fyne.NewMenu("", items...))
	return container.NewBorder(
		container.NewVBox(container.NewHBox(t, layout.NewSpacer(), options), widget.NewSeparator()),
		nil,
		nil,
		nil,
		container.NewScroll(content),
	)
}

func SetTheme(value string, app fyne.App) {
	switch value {
	case lang.L("Light"):
		app.Settings().SetTheme(&customTheme.ForcedVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantLight})
	case lang.L("Dark"):
		app.Settings().SetTheme(&customTheme.ForcedVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantDark})
	case lang.L("Pastel"):
		ctp := catppuccin.New()
		ctp.SetFlavor(catppuccin.Frappe)
		app.Settings().SetTheme(ctp)
	default:
		helper.Logger.Fatal().Msgf("Unsupported value '%s' for theme. Should be %s, %s or %s", value, lang.L("Light"), lang.L("Dark"), lang.L("Pastel"))
	}
	app.Preferences().SetString(PreferenceTheme, value)
	helper.Logger.Info().Msgf("Theme set to %s", value)
}

func SetBackendIP(value string, app fyne.App) {
	app.Preferences().SetString(PreferenceBackendIP, value)
	helper.Logger.Info().Msgf("Backend IP set to %s", value)
}

func SetBackendPort(value string, app fyne.App) {
	app.Preferences().SetString(PreferenceBackendPort, value)
	helper.Logger.Info().Msgf("Backend port set to %s", value)
}

func SetBackendPollingInterval(value int, app fyne.App) {
	app.Preferences().SetInt(PreferenceBackendPollingInterval, value)
	helper.Logger.Info().Msgf("Backend polling interval set to %d", value)
}

// GetFullscreen returns the PreferenceFullscreen app preference value
func GetFullscreen(app fyne.App) bool {
	return app.Preferences().BoolWithFallback(PreferenceFullscreen, FullscreenDefault)
}

func SetFullScreen(value bool, app fyne.App, topWin fyne.Window, currentWin fyne.Window) {
	app.Preferences().SetBool(PreferenceFullscreen, value)
	topWin.SetFullScreen(app.Preferences().BoolWithFallback(PreferenceFullscreen, FullscreenDefault))
	currentWin.Show()
	helper.Logger.Info().Msgf("Fullscreen set to %s", strconv.FormatBool(value))
}

func SetSystemTray(value bool, app fyne.App) {
	app.Preferences().SetBool(PreferenceSystemTray, value)
	helper.Logger.Info().Msgf("System tray set to %s", strconv.FormatBool(value))
}

func SetBackendProtocol(value string, app fyne.App) {
	app.Preferences().SetString(PreferenceBackendProtocol, value)
	helper.Logger.Info().Msgf("Backend protocol set to %s", value)
}

func NewSettings(app fyne.App, topWindow fyne.Window) {

	win := app.NewWindow(lang.L("General Settings"))

	///////////////////////////////////////////////////////////////////////////
	// General Tab
	theme := NewSettingItemOptions(
		lang.L("Theme"),
		lang.L("Theme details"),
		[]string{lang.L("Light"), lang.L("Dark"), lang.L("Pastel")},
		ThemeDefault,
		func() string {
			return app.Preferences().StringWithFallback(PreferenceTheme, ThemeDefault)
		},
		func(v string) {
			SetTheme(v, app)
		},
		win,
	)
	fullscreen := NewSettingItemSwitch(
		lang.L("Fullscreen"),
		lang.L("Fullscreen details"),
		func() bool {
			return app.Preferences().BoolWithFallback(PreferenceFullscreen, FullscreenDefault)
		},
		func(v bool) {
			SetFullScreen(v, app, topWindow, win)
		},
	)
	closeButton := NewSettingItemSwitch(
		lang.L("Close button"),
		lang.L("Close button details"),
		func() bool {
			return app.Preferences().BoolWithFallback(PreferenceSystemTray, SystemTrayDefault)
		},
		func(v bool) {
			SetSystemTray(v, app)
		},
	)
	logLevel := NewSettingItemOptions(
		lang.L("Log level"),
		lang.L("Log level details"),
		helper.LogLevelNames(),
		helper.LogLevelDefault,
		func() string {
			return app.Preferences().StringWithFallback(helper.PreferenceLogLevel, helper.LogLevelDefault)
		},
		func(v string) {
			helper.SetLogLevel(v, app)
		},
		win,
	)
	language := NewSettingItemOptions(
		lang.L("Language"),
		lang.L("Language option"),
		resources.GetTranslationNames(),
		resources.LanguageDefault,
		func() string {
			return app.Preferences().StringWithFallback(resources.PreferenceLanguage, resources.LanguageDefault)
		},
		func(v string) {
			resources.SetLanguage(v, app)
		},
		win,
	)

	generalItems := []SettingItem{
		NewSettingItemHeading(lang.L("Interface")),
		theme,
		fullscreen,
		NewSettingItemSeperator(),
		NewSettingItemHeading(lang.L("Application")),
		closeButton,
		logLevel,
		language,
	}

	generalSettingsList := NewSettingList(generalItems)

	generalReset := SettingAction{
		Label: lang.L("Reset to default values"),
		Action: func() {
			SetTheme(ThemeDefault, app)
			SetFullScreen(FullscreenDefault, app, topWindow, win)
			SetSystemTray(SystemTrayDefault, app)
			helper.SetLogLevel(helper.LogLevelDefault, app)
			resources.SetLanguage(resources.LanguageDefault, app)
			generalSettingsList.Refresh()
		},
	}

	generalActionsList := []SettingAction{generalReset}

	///////////////////////////////////////////////////////////////////////////
	// Backend Tab
	backendIP := NewSettingItemUserInput(
		lang.L("Backend IP"),
		lang.L("Backend IP details"),
		lang.L("Backend IP placeholder"),
		`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$|localhost$`, // IPv4 or localhost regex
		"userEntry can only contain letters, numbers, '.', and ':'",
		BackendIPDefault,
		func() string {
			return app.Preferences().StringWithFallback(PreferenceBackendIP, BackendIPDefault)
		},
		func(v string) {
			SetBackendIP(v, app)
		},
		win,
	)
	backendProtocol := NewSettingItemOptions(
		lang.L("Backend protocol"),
		lang.L("Backend protocol details"),
		[]string{BackendProtocolDefault, BackendProtocolSafe},
		BackendProtocolDefault,
		func() string {
			return app.Preferences().StringWithFallback(PreferenceBackendProtocol, BackendProtocolDefault)
		},
		func(v string) {
			SetBackendProtocol(v, app)
		},
		win,
	)
	backendPort := NewSettingItemUserInput(
		lang.L("Backend Port"),
		lang.L("Backend Port details"),
		lang.L("Backend Port placeholder"),
		"^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$", // 0-65535 port regex
		"userEntry can only contain letters, numbers, '.', and ':'",
		BackendPortDefault,
		func() string {
			return app.Preferences().StringWithFallback(PreferenceBackendPort, BackendPortDefault)
		},
		func(v string) {
			SetBackendPort(v, app)
		},
		win,
	)

	backendPollingInterval := NewSettingItemSlider(
		lang.L("Backend polling"),
		lang.L("Backend polling details"),
		float64(BackendPollingIntervalMin),
		float64(BackendPollingIntervalMax),
		float64(BackendPollingIntervalDefault),
		func() float64 {
			return float64(app.Preferences().IntWithFallback(PreferenceBackendPollingInterval, BackendPollingIntervalDefault))
		},
		func(v float64) {
			SetBackendPollingInterval(int(v), app)
		},
		win,
	)

	backendItems := []SettingItem{
		backendIP,
		backendProtocol,
		backendPort,
		backendPollingInterval,
	}

	backendSettingsList := NewSettingList(backendItems)

	backendReset := SettingAction{
		Label: lang.L("Reset to default values"),
		Action: func() {
			SetBackendIP(BackendIPDefault, app)
			SetBackendProtocol(BackendProtocolDefault, app)
			SetBackendPort(BackendPortDefault, app)
			SetBackendPollingInterval(BackendPollingIntervalDefault, app)
			backendSettingsList.Refresh()
		},
	}

	backendActionsList := []SettingAction{backendReset}

	tabs := container.NewAppTabs(
		container.NewTabItem(lang.L("General"), MakeSettingsPage(lang.L("General"), generalSettingsList, generalActionsList)),
		container.NewTabItem(lang.L("Backend"), MakeSettingsPage(lang.L("Backend"), backendSettingsList, backendActionsList)),
	)

	tabs.SetTabLocation(container.TabLocationLeading)
	win.SetContent(tabs)
	win.Resize(fyne.NewSize(750, 600))
	win.CenterOnScreen()
	win.Show()
}
