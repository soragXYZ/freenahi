package menu

import (
	"net/url"

	"fyne.io/fyne/v2"
	fyneSettings "fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"

	"freenahiFront/internal/account"
	financialassets "freenahiFront/internal/financialAssets"
	"freenahiFront/internal/loan"
	"freenahiFront/internal/settings"
	"freenahiFront/internal/tools"
	"freenahiFront/internal/topmenu"
	"freenahiFront/internal/transactions"
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

	parametersMenu := fyne.NewMenu(lang.L("Settings"),
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

	tutorialMenu := fyne.NewMenu(lang.L("First steps"),
		fyne.NewMenuItem(lang.L("Powens configuration"), func() {
			u, _ := url.Parse("https://soragxyz.github.io/freenahi/getStarted/powens/")
			_ = app.OpenURL(u)
		}),
		fyne.NewMenuItem(lang.L("Backend configuration"), func() {
			u, _ := url.Parse("https://soragxyz.github.io/freenahi/getStarted/backend/")
			_ = app.OpenURL(u)
		}),
	)

	// Add new entries here if needed
	return fyne.NewMainMenu(
		parametersMenu,
		tutorialMenu,
	)
}

func NewLeftMenu(app fyne.App, win fyne.Window) *container.AppTabs {
	tabs := container.NewAppTabs(
		container.NewTabItem(lang.L("Financial assets"), financialassets.NewFinancialAssetsScreen(app, win)),
		container.NewTabItem(lang.L("Accounts"), account.NewAccountScreen(app)),
		container.NewTabItem(lang.L("Transactions"), transactions.NewTransactionScreen(app, win)),
		container.NewTabItem(lang.L("Loans"), loan.NewLoanScreen(app, win)),
		container.NewTabItem(lang.L("Tools"), tools.NewToolsScreen(app, win)),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}
