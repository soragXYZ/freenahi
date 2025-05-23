package account

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Const which are used in order to increase code readability
// These consts are the name of the columns and used later in a switch case
const (
	accountNameColumn int = iota
	valueColumn
	currencyColumn
	lastUpdateColumn
	typeColumn
	usageColumn
	IBANColumn
	accountNumberColumn
	numberOfColumn

	unselectTime = 200 * time.Millisecond
)

type BankAccount struct {
	Number        string  `json:"number"`
	Original_name string  `json:"original_name"`
	Balance       float32 `json:"balance"`
	Last_update   string  `json:"last_update"`
	Iban          string  `json:"iban"`
	Currency      string  `json:"currency"`
	Account_type  string  `json:"type"`
	Usage         string  `json:"usage"`
}

// Create the transaction screen
func NewAccountScreen(app fyne.App) fyne.CanvasObject {

	accountTable := createAccountTable(app)

	manageButton := widget.NewButton(lang.L("Manage account with Powens"), func() {

		webviewURL := getWebviewManageConnexionLink(app)

		if err := app.OpenURL(webviewURL); err != nil {
			helper.Logger.Error().Err(err).Msg("Cannot open the URL")
		}
	})

	screen := container.NewBorder(
		container.NewVBox(manageButton, widget.NewSeparator()),
		nil,
		nil,
		nil,
		accountTable,
	)

	return screen
}

// Create the table of transaction
func createAccountTable(app fyne.App) *widget.Table {

	accountNameLabel := widget.NewLabel(lang.L("Account name"))
	accountNameLabel.TextStyle.Bold = true
	testAccountNameLabelSize := widget.NewLabel("COMPTE COURANT NUMERO XXX").MinSize().Width

	valueLabel := widget.NewLabel(lang.L("Value"))
	valueLabel.TextStyle.Bold = true
	testValueLabelSize := widget.NewLabel("-123456123.00").MinSize().Width

	currencyLabel := widget.NewLabel(lang.L("Currency"))
	currencyLabel.TextStyle.Bold = true
	testCurrencyLabelSize := widget.NewLabel("EUR").MinSize().Width

	lastUpdateLabel := widget.NewLabel(lang.L("Last update"))
	lastUpdateLabel.TextStyle.Bold = true
	testLastUpdateLabelSize := widget.NewLabel("XXXX-YY-ZZ").MinSize().Width

	typeLabel := widget.NewLabel(lang.L("Type"))
	typeLabel.TextStyle.Bold = true
	testTypeLabelSize := widget.NewLabel(lang.L("capitalisation")).MinSize().Width

	usageLabel := widget.NewLabel(lang.L("Usage"))
	usageLabel.TextStyle.Bold = true
	testUsageLabelSize := widget.NewLabel(lang.L("PRIV")).MinSize().Width

	IBANLabel := widget.NewLabel(lang.L("IBAN"))
	IBANLabel.TextStyle.Bold = true
	testIBANLabelSize := widget.NewLabel("FR76 3000 1007 9412 3456 7890 185").MinSize().Width

	numberLabel := widget.NewLabel(lang.L("Account number"))
	numberLabel.TextStyle.Bold = true
	testNumberLabelSize := widget.NewLabel("550e8400-e29b-41d4-a716-446655440000").MinSize().Width

	// Fill bank accounts. The first row is a special item only used for the table header (no real data)
	bankAccounts := []BankAccount{{
		Iban: "columnHeader",
	}}
	bankAccounts = append(bankAccounts, GetBankAccounts(app, "")...)

	accountList := widget.NewTable(
		func() (int, int) {
			return len(bankAccounts), numberOfColumn
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("template"))
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			// Clean the cell from the previous value
			o.(*fyne.Container).RemoveAll()

			// If we are on the first row, we set special values and we will pin this row to create a header with it
			if id.Row == 0 {
				switch id.Col {
				case accountNameColumn:
					helper.AddHAligned(o, accountNameLabel)
				case valueColumn:
					helper.AddHAligned(o, valueLabel)
				case currencyColumn:
					helper.AddHAligned(o, currencyLabel)
				case lastUpdateColumn:
					helper.AddHAligned(o, lastUpdateLabel)
				case typeColumn:
					helper.AddHAligned(o, typeLabel)
				case usageColumn:
					helper.AddHAligned(o, usageLabel)
				case IBANColumn:
					helper.AddHAligned(o, IBANLabel)
				case accountNumberColumn:
					helper.AddHAligned(o, numberLabel)
				default:
					helper.Logger.Fatal().Msg("Too much column in the grid")
				}
				return
			}

			// Update the cell by adding content according to its "type" (IBAN, accountName, value, currency, lastUpdate, type, usage, number)
			switch id.Col {
			case accountNameColumn:
				scroller := container.NewHScroll(container.NewHBox(
					layout.NewSpacer(),
					widget.NewLabel(bankAccounts[id.Row].Original_name),
					layout.NewSpacer(),
				))
				scroller.SetMinSize(fyne.NewSize(testAccountNameLabelSize, scroller.MinSize().Height))
				helper.AddHAligned(o, scroller)

			case valueColumn:
				value := fmt.Sprintf("%.2f", bankAccounts[id.Row].Balance)
				valueText := widget.NewLabel(helper.ValueSpacer(value))

				if bankAccounts[id.Row].Balance < 0 {
					valueText.Importance = widget.WarningImportance
				}

				helper.AddHAligned(o, valueText)

			case currencyColumn:
				helper.AddHAligned(o, widget.NewLabel(bankAccounts[id.Row].Currency))

			case lastUpdateColumn:
				parsedTxDate, err := time.Parse("2006-01-02 15:04:05", bankAccounts[id.Row].Last_update)
				if err != nil {
					helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", bankAccounts[id.Row].Last_update)
				}
				helper.AddHAligned(o, widget.NewLabel(parsedTxDate.Format("2006-01-02")))

			case typeColumn:
				// Each type is here: https://docs.powens.com/api-reference/products/data-aggregation/bank-account-types#accounttypename-values
				helper.AddHAligned(o, widget.NewLabel(lang.L(bankAccounts[id.Row].Account_type)))

			case usageColumn:
				if bankAccounts[id.Row].Usage != "" {
					helper.AddHAligned(o, widget.NewLabel(lang.L(bankAccounts[id.Row].Usage)))
				}

			case IBANColumn:
				helper.AddHAligned(o, widget.NewLabel(ibanSpacer(bankAccounts[id.Row].Iban)))

			case accountNumberColumn:
				helper.AddHAligned(o, widget.NewLabel(bankAccounts[id.Row].Number))

			default:
				helper.Logger.Fatal().Msg("Too much column in the grid")
			}
		},
	)

	accountList.OnSelected = func(id widget.TableCellID) {

		// Do nothing if the user selected the column header
		if id.Row == 0 {
			return
		}

		go func() {
			time.Sleep(unselectTime)
			fyne.Do(func() {
				accountList.Unselect(id)
			})
		}()
	}

	// We set the width of the columns, ie the max between the language name header size and actual value
	// For example, the max between "Value" and "-123456123.00", or "Montant" and "-123456123.00" in french
	accountList.SetColumnWidth(accountNameColumn, float32(math.Max(
		float64(testAccountNameLabelSize),
		float64(accountNameLabel.MinSize().Width))),
	)
	accountList.SetColumnWidth(valueColumn, float32(math.Max(
		float64(testValueLabelSize),
		float64(valueLabel.MinSize().Width))),
	)
	accountList.SetColumnWidth(currencyColumn, float32(math.Max(
		float64(testCurrencyLabelSize),
		float64(currencyLabel.MinSize().Width))),
	)
	accountList.SetColumnWidth(lastUpdateColumn, float32(math.Max(
		float64(testLastUpdateLabelSize),
		float64(lastUpdateLabel.MinSize().Width))),
	)
	accountList.SetColumnWidth(typeColumn, float32(math.Max(
		float64(testTypeLabelSize),
		float64(typeLabel.MinSize().Width))),
	)
	accountList.SetColumnWidth(usageColumn, float32(math.Max(
		float64(testUsageLabelSize),
		float64(usageLabel.MinSize().Width))),
	)
	accountList.SetColumnWidth(IBANColumn, float32(math.Max(
		float64(testIBANLabelSize),
		float64(IBANLabel.MinSize().Width))),
	)
	accountList.SetColumnWidth(accountNumberColumn, float32(math.Max(
		float64(testNumberLabelSize),
		float64(numberLabel.MinSize().Width))),
	)

	accountList.StickyRowCount = 1 // Basically, we are setting a table header because the first row contains special data

	return accountList
}

// Add spacing to IBAN to make it more easily readable
// Ex: From FR7630001007941234567890185 to FR76 3000 1007 9412 3456 7890 185
func ibanSpacer(IBAN string) string {
	var modifiedIban string
	for pos, char := range IBAN {
		if pos%4 == 0 && pos != 0 {
			modifiedIban = modifiedIban + " "
		}
		modifiedIban = modifiedIban + string(char)
	}
	return modifiedIban
}

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/accounts" and retrieve bank accounts
func GetBankAccounts(app fyne.App, accountType string) []BankAccount {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	var url string
	if accountType == "" {
		url = fmt.Sprintf("%s://%s:%s/bank_account/", backendProtocol, backendIp, backendPort)
	} else {
		url = fmt.Sprintf("%s://%s:%s/bank_account/?type=%s", backendProtocol, backendIp, backendPort, accountType)
	}
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

	var accounts []BankAccount
	if err := json.Unmarshal(body, &accounts); err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot unmarshal bank accounts")
		return nil

	}

	return accounts
}

// ToDo: modify the function to return an error and display it if sth went wrong in the backend
// Call the backend endpoint "/webview/manageConnectionLink" and get the webview URL
func getWebviewManageConnexionLink(app fyne.App) *url.URL {

	backendIp := app.Preferences().StringWithFallback(settings.PreferenceBackendIP, settings.BackendIPDefault)
	backendProtocol := app.Preferences().StringWithFallback(settings.PreferenceBackendProtocol, settings.BackendProtocolDefault)
	backendPort := app.Preferences().StringWithFallback(settings.PreferenceBackendPort, settings.BackendPortDefault)

	targetURL := fmt.Sprintf("%s://%s:%s/webview/manageConnectionLink/", backendProtocol, backendIp, backendPort)
	resp, err := http.Get(targetURL)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("Cannot run http get request")
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		helper.Logger.Error().Err(err).Msg("ReadAll error")
		return nil
	}

	webviewURL, err := url.Parse(string(body))

	if err != nil {
		helper.Logger.Error().Err(err).Msg("URL parse error")
		return nil
	}

	return webviewURL
}
