package menu

import (
	"net/url"

	"fyne.io/fyne/v2"
	fyneSettings "fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"

	"freenahiFront/internal/account"
	"freenahiFront/internal/settings"
	"freenahiFront/internal/topmenu"
	"freenahiFront/internal/transactions"
	"freenahiFront/internal/welcome"
)

// Tutorial defines the data structure for a tutorial
type Tutorial struct {
	Title string
	View  func(w fyne.Window) fyne.CanvasObject
}

func NewTopMenu(app fyne.App, win fyne.Window) *fyne.MainMenu {
	uiFyneSettings := func() {
		w := app.NewWindow(lang.L("Interface Settings"))
		w.SetContent(fyneSettings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(440, 520))
		w.Show()
	}

	helpMenu := fyne.NewMenu(lang.L("Settings"),
		fyne.NewMenuItem(lang.L("Interface Settings"), uiFyneSettings),
		fyne.NewMenuItem(lang.L("General Settings"), func() { settings.NewSettings(app, win) }),
		fyne.NewMenuItem(lang.L("User data"), func() { topmenu.ShowUserDataDialog(app, win) }),
		fyne.NewMenuItem(lang.L("About"), func() { topmenu.ShowAboutDialog(app, win) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(lang.L("Documentation"), func() {
			u, _ := url.Parse("https://soragxyz.github.io/freenahi/")
			_ = app.OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(lang.L("Contribute"), func() {
			u, _ := url.Parse("https://soragxyz.github.io/freenahi/other/contribute/")
			_ = app.OpenURL(u)
		}),
		// a quit item will be appended to our first menu, cannot remove it
	)

	// Add new entries here if needed
	return fyne.NewMainMenu(
		helpMenu,
	)
}

func NewLeftMenu(app fyne.App, win fyne.Window) *container.AppTabs {
	tabs := container.NewAppTabs(
		container.NewTabItem("Welcome", welcome.NewWelcomeScreen()),
		container.NewTabItem(lang.L("Accounts"), account.NewAccountScreen(app, win)),
		container.NewTabItem(lang.L("Transactions"), transactions.NewTransactionScreen(app, win)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}
