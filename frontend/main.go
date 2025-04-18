package main

import (
	"fmt"
	"log"
	"net/url"

	"freenahiFront/helper"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	appID = "github.soragXYZ.freenahi"
)

// Set system tray if desktop (mini icon like wifi, shield, notifs, etc...)
func makeTray(app fyne.App) {
	if desk, isDesktop := app.(desktop.App); isDesktop {
		h := fyne.NewMenuItem("Come back to Freenahi", func() {})
		h.Icon = theme.HomeIcon()
		h.Action = func() { log.Println("System tray menu tapped for Welcome") }
		menu := fyne.NewMenu("SystemTrayMenu", h)
		desk.SetSystemTrayMenu(menu)
	}
}

// Watch events
func logLifecycle(app fyne.App) {
	app.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	app.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	app.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	app.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

// Create the option menu
func makeMenu(fyneApp fyne.App, w fyne.Window) *fyne.MainMenu {
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://soragxyz.github.io/freenahi/")
			_ = fyneApp.OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Contribute", func() {
			u, _ := url.Parse("https://soragxyz.github.io/freenahi/other/contribute/")
			_ = fyneApp.OpenURL(u)
		}),
		// a quit item will be appended to our first menu, cannot remove it ?
	)

	// Add new entries here if needed
	main := fyne.NewMainMenu(
		helpMenu,
	)
	return main
}

func makeNav() fyne.CanvasObject {
	app := fyne.CurrentApp()

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			app.Settings().SetTheme(&helper.ForcedVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantDark})
		}),
		widget.NewButton("Light", func() {
			app.Settings().SetTheme(&helper.ForcedVariant{Theme: theme.DefaultTheme(), Variant: theme.VariantLight})
		}),
	)

	return container.NewBorder(nil, themes, nil, nil, nil)
}

func main() {
	fyneApp := app.NewWithID(appID)
	logLifecycle(fyneApp)
	makeTray(fyneApp)

	w := fyneApp.NewWindow("Main Window")

	w.SetMainMenu(makeMenu(fyneApp, w))
	w.SetContent(widget.NewLabel("Hello World! Messing with the front"))
	w.Resize(fyne.NewSize(400, 400))

	content := container.NewStack()
	tutorial := container.NewBorder(nil, nil, nil, nil, content)
	if fyne.CurrentDevice().IsMobile() {
		w.SetContent(makeNav())
	} else {
		split := container.NewHSplit(makeNav(), tutorial)
		split.Offset = 0.2
		w.SetContent(split)
	}

	// Exit cross on the window (with reduce and fullscreen)
	w.SetCloseIntercept(func() {
		fmt.Println("Tried to quit")
		fyneApp.Quit()
	})

	w.ShowAndRun()
}
