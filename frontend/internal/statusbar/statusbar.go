package statusbar

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/hashicorp/go-version"

	"freenahiFront/internal/github"
	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"
)

const (
	downloadURL = "https://github.com/soragXYZ/freenahi/releases"

	owner = "soragxyz"
	repo  = "freenahi"
)

// StatusBar struct, containing info for the application and the backend
type StatusBar struct {
	widget.BaseWidget

	newVersionAvailable *fyne.Container
	backendStatus       *StatusBarItem
}

// This struct holds info that need to be updated by the go routine
// ex: check if backend is available and update displayed info accordingly
type backendInfo struct {
	remoteBackendVersionItem, currentBackendVersionItem, backendStatusItem *widget.Label
	remoteSpinner, currentSpinner                                          *widget.Activity
}

// Create a new status bar located at the bottom of the main window
func NewStatusBar(app fyne.App, parentWin fyne.Window) *StatusBar {
	statusBar := &StatusBar{
		newVersionAvailable: container.NewHBox(),
	}
	statusBar.ExtendBaseWidget(statusBar)

	backendInfo := &backendInfo{
		remoteBackendVersionItem:  widget.NewLabel(""),
		currentBackendVersionItem: widget.NewLabel(""),
		backendStatusItem:         widget.NewLabel(""),
		remoteSpinner:             widget.NewActivity(),
		currentSpinner:            widget.NewActivity(),
	}

	backendInfo.remoteBackendVersionItem.Hide()
	backendInfo.currentBackendVersionItem.Hide()
	backendInfo.remoteSpinner.Start()
	backendInfo.currentSpinner.Start()

	statusBar.startGoroutines(app, parentWin, backendInfo)

	statusBar.backendStatus = NewStatusBarItem(theme.NewWarningThemedResource(theme.MediaRecordIcon()), lang.L("Contacting backend"), func() {
		statusBar.showBackendDialog(parentWin, backendInfo)
	})

	return statusBar
}

// Display a dialog box with backend versions and status (called when click on backend status bar)
func (a *StatusBar) showBackendDialog(parentWin fyne.Window, backendInfo *backendInfo) {

	content := container.New(
		layout.NewCustomPaddedVBoxLayout(0),
		container.New(
			layout.NewCustomPaddedVBoxLayout(0),
			container.NewHBox(
				widget.NewLabel(lang.L("Latest version")),
				layout.NewSpacer(),
				container.NewStack(backendInfo.remoteSpinner, backendInfo.remoteBackendVersionItem),
			),
			container.NewHBox(
				widget.NewLabel(lang.L("Current version")),
				layout.NewSpacer(),
				container.NewStack(backendInfo.currentSpinner, backendInfo.currentBackendVersionItem),
			),
		),
		widget.NewSeparator(),
		container.NewHBox(layout.NewSpacer(), backendInfo.backendStatusItem, layout.NewSpacer()),
	)

	d := dialog.NewCustom(lang.L("Backend version"), lang.L("Close"), content, parentWin)
	d.Resize(fyne.NewSize(d.MinSize().Width*1.3, d.MinSize().Height))
	d.Show()
}

// Display the status bar
func (a *StatusBar) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			layout.NewSpacer(),
			a.newVersionAvailable,
			widget.NewSeparator(),
			a.backendStatus,
			widget.NewSeparator(),
		))
	return widget.NewSimpleRenderer(c)
}

// StatusBarItem is a widget with a label and an optional icon, which can be tapped.
type StatusBarItem struct {
	widget.BaseWidget
	icon  *widget.Icon
	label *widget.Label

	// The function that is called when the label is tapped.
	OnTapped func()

	hovered bool
}

var _ fyne.Tappable = (*StatusBarItem)(nil)
var _ desktop.Hoverable = (*StatusBarItem)(nil)

// Create a new statusBar item
func NewStatusBarItem(res fyne.Resource, text string, tapped func()) *StatusBarItem {
	w := &StatusBarItem{OnTapped: tapped, label: widget.NewLabel(text)}
	if res != nil {
		w.icon = widget.NewIcon(res)
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *StatusBarItem) Tapped(_ *fyne.PointEvent) {
	if w.OnTapped != nil {
		w.OnTapped()
	}
}

// Cursor returns the cursor type of this widget
func (w *StatusBarItem) Cursor() desktop.Cursor {
	if w.hovered {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

// MouseIn is a hook that is called if the mouse pointer enters the element.
func (w *StatusBarItem) MouseIn(e *desktop.MouseEvent) {
	w.hovered = true
}

func (w *StatusBarItem) MouseMoved(*desktop.MouseEvent) {
	// needed to satisfy the interface only
}

// MouseOut is a hook that is called if the mouse pointer leaves the element.
func (w *StatusBarItem) MouseOut() {
	w.hovered = false
}

// Display the statusBarItem
func (w *StatusBarItem) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewHBox()
	if w.icon != nil {
		c.Add(w.icon)
	}
	c.Add(w.label)
	return widget.NewSimpleRenderer(c)
}

// Start asynchronous jobs:
// - check if an application update is available
// - check if the backend is reachable and update data accordingly
func (a *StatusBar) startGoroutines(app fyne.App, parentWin fyne.Window, backendInfo *backendInfo) {

	go a.showApplicationUpdateInStatusBar(app, parentWin)
	go a.showBackendInStatusBar(app, backendInfo)

}

// Modify the bottom status bar to indicate (or not) that an update for the application is available
func (a *StatusBar) showApplicationUpdateInStatusBar(app fyne.App, parentWin fyne.Window) {
	currentVersion := app.Metadata().Version
	remoteVersion, isRemoteNewer, err := github.AvailableUpdate("ErikKalkoken", "evebuddy", currentVersion) // ToDo: use correct repo values

	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot fetch github version")
	}

	// If no update available, show nothing
	if !isRemoteNewer {
		return
	}

	// If an update is available, create a clickable hyperlink and display versions
	hyperlink := widget.NewHyperlink(lang.L("Application update available"), nil)
	hyperlink.OnTapped = func() {
		c := container.NewVBox(
			container.NewHBox(widget.NewLabel(lang.L("Latest version")), layout.NewSpacer(), widget.NewLabel(remoteVersion)),
			container.NewHBox(widget.NewLabel(lang.L("Current version")), layout.NewSpacer(), widget.NewLabel(currentVersion)),
		)

		d := dialog.NewCustomConfirm(lang.L("Application update available"), lang.L("Downloads"), lang.L("Close"), c, func(ok bool) {
			if !ok {
				return
			}
			download, _ := url.Parse(downloadURL)
			if err := app.OpenURL(download); err != nil {
				helper.Logger.Error().Err(err).Msg("Cannot open the URL")
			}
		}, parentWin,
		)
		d.Show()
	}
	fyne.Do(func() {
		a.newVersionAvailable.Add(widget.NewSeparator())
		a.newVersionAvailable.Add(widget.NewIcon(theme.DownloadIcon()))
		a.newVersionAvailable.Add(hyperlink)
	})

}

// Check the backend status at regular time interval, every xxx seconds
func (a *StatusBar) showBackendInStatusBar(app fyne.App, backendInfo *backendInfo) {

	var statusBarText string

	for {
		backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
		backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
		backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

		statusBarSuccessIcon := theme.NewSuccessThemedResource(theme.MediaRecordIcon())
		statusBarWarnIcon := theme.NewWarningThemedResource(theme.MediaRecordIcon())
		statusBarErrorIcon := theme.NewErrorThemedResource(theme.MediaRecordIcon())

		backendInfo.backendStatusItem.Text = ""

		// Get remote backend version (ie the latest version available on github / dockerHub)
		imageLatestVersion, err := github.FetchDockerLatest(owner, repo)
		if err != nil {
			helper.Logger.Error().Err(err).Msg("Cannot fetch latest docker image version")
			return
		}

		remoteBackendVersion, err := version.NewVersion(imageLatestVersion)
		if err != nil {
			helper.Logger.Error().Err(err).Msg("Version error")
			return
		}
		backendInfo.backendStatusItem.Text = lang.L("Got latest backend version available")
		// ToDo: add error message
		// backendInfo.backendStatusItem.Text = "Could not get the latest backend version available\n"

		helper.Logger.Trace().Str("Remote backend version", remoteBackendVersion.String()).Msg("Latest backend version obtained")
		backendInfo.remoteBackendVersionItem.Text = remoteBackendVersion.String()

		// Update the UI if some changes are detected
		if len(backendInfo.remoteBackendVersionItem.Text) != 0 {
			fyne.Do(func() {
				backendInfo.remoteBackendVersionItem.Refresh()
				backendInfo.remoteSpinner.Hide()
				backendInfo.remoteBackendVersionItem.Show()
			})
		}

		// Get current backend version (ie the version you are currently using)
		url := fmt.Sprintf("%s://%s:%s/version/", backendProtocol, backendIp, backendPort)
		resp, err := http.Get(url)

		if e, ok := err.(net.Error); ok && e.Timeout() { // Backend unreachable
			helper.Logger.Error().Err(err).Msg("Timeout")
			statusBarText = lang.L("Backend unreachable")
			backendInfo.backendStatusItem.Text += "Timeout: " + err.Error() + "\n\n"
			backendInfo.backendStatusItem.Text += lang.L("Is IP correct")
			backendInfo.backendStatusItem.Importance = widget.DangerImportance

			fyne.Do(func() {
				backendInfo.backendStatusItem.Refresh()
				a.backendStatus.icon.SetResource(statusBarErrorIcon)
			})

		} else if err != nil { // Backend Error
			helper.Logger.Error().Err(err).Msg("Cannot run http get request")
			statusBarText = lang.L("Backend Error")
			backendInfo.backendStatusItem.Text += lang.L("Backend Error") + ": " + err.Error()
			backendInfo.backendStatusItem.Importance = widget.DangerImportance

			fyne.Do(func() {
				backendInfo.backendStatusItem.Refresh()
				a.backendStatus.icon.SetResource(statusBarErrorIcon)
			})

		} else { // Backend reachable

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				helper.Logger.Error().Err(err).Msg("ReadAll error")
			}

			currentBackendVersion, err := version.NewVersion(string(body))
			if err != nil {
				helper.Logger.Error().Err(err).Msg("Version error")
			}

			helper.Logger.Trace().Str("current backend version", currentBackendVersion.String()).Msg("Backend is reachable")
			backendInfo.currentBackendVersionItem.Text = currentBackendVersion.String()
			backendInfo.backendStatusItem.Text += lang.L("Got current backend version")

			// ToDo: when latest version is implemented, modify this condition because
			// we can get the currentVersion but not the remote
			if currentBackendVersion.LessThan(remoteBackendVersion) { // backend can be updated
				statusBarText = lang.L("Backend update available")
				backendInfo.backendStatusItem.Text += lang.L("Backend update available")
				backendInfo.backendStatusItem.Importance = widget.MediumImportance

				fyne.Do(func() {
					backendInfo.backendStatusItem.Refresh()
					a.backendStatus.icon.SetResource(statusBarWarnIcon)
				})
			} else { // backend is up to date
				statusBarText = lang.L("Backend up-to-date")
				backendInfo.backendStatusItem.Text += lang.L("Backend up-to-date")
				backendInfo.backendStatusItem.Importance = widget.SuccessImportance
				fyne.Do(func() {
					backendInfo.backendStatusItem.Refresh()
					a.backendStatus.icon.SetResource(statusBarSuccessIcon)
				})
			}
		}

		// Update the UI if some changes are detected
		if len(backendInfo.currentBackendVersionItem.Text) != 0 {
			fyne.Do(func() {
				backendInfo.currentBackendVersionItem.Refresh()
				backendInfo.currentSpinner.Hide()
				backendInfo.currentBackendVersionItem.Show()
			})
		}

		fyne.Do(func() { a.backendStatus.label.SetText(statusBarText) })

		secondsToWait := app.Preferences().IntWithFallback(settings.PreferenceBackendPollingInterval, settings.BackendPollingIntervalDefault)
		time.Sleep(time.Duration(secondsToWait) * time.Second)

	}
}
