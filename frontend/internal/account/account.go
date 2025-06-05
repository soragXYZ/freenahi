package account

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"time"

	"freenahiFront/internal/helper"
	"freenahiFront/internal/settings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Const which are used in order to increase code readability
// These consts are the name of the columns and used later in a switch case
const (
	bankNameColumn int = iota
	accountNameColumn
	valueColumn
	currencyColumn
	lastUpdateColumn
	typeColumn
	usageColumn
	IBANColumn
	accountNumberColumn
	numberOfColumns

	unselectTime = 200 * time.Millisecond
)

const (
	sortOff = iota // Used to sort data when clicking on a table header
	sortAsc
	sortDesc
	numberOfSorts
)

// Var holding the sort type for each
var columnSort = [numberOfColumns]int{}

type BankAccount struct {
	Id                 int     `json:"id"`
	Number             string  `json:"number"`
	Bank_Original_name string  `json:"bank_original_name"`
	Original_name      string  `json:"original_name"`
	Balance            float32 `json:"balance"`
	Last_update        string  `json:"last_update"`
	Iban               string  `json:"iban"`
	Currency           string  `json:"currency"`
	Account_type       string  `json:"type"`
	Usage              string  `json:"usage"`
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
func createAccountTable(app fyne.App) *fyne.Container {

	// These values are used later to set column width sizes, which are the max between the header and an actual value
	testIconSize := widget.NewIcon(theme.RadioButtonCheckedIcon()).MinSize().Width

	bankNameHeader := widget.NewLabel(lang.L("Name"))
	bankNameHeader.TextStyle.Bold = true
	bankNameHeaderSize := bankNameHeader.MinSize().Width + testIconSize
	testBankNameLabelSize := widget.NewLabel("Connecteur de test").MinSize().Width

	accountNameHeader := widget.NewLabel(lang.L("Account name"))
	accountNameHeader.TextStyle.Bold = true
	accountNameHeaderSize := accountNameHeader.MinSize().Width + testIconSize
	testAccountNameLabelSize := widget.NewLabel("COMPTE COURANT NUMERO XXX").MinSize().Width

	valueHeader := widget.NewLabel(lang.L("Value"))
	valueHeader.TextStyle.Bold = true
	valueHeaderSize := valueHeader.MinSize().Width + testIconSize
	testValueLabelSize := widget.NewLabel("-123456123.00").MinSize().Width

	currencyHeader := widget.NewLabel(lang.L("Currency"))
	currencyHeader.TextStyle.Bold = true
	currencyHeaderSize := currencyHeader.MinSize().Width + testIconSize
	testCurrencyLabelSize := widget.NewLabel("EUR").MinSize().Width

	lastUpdateHeader := widget.NewLabel(lang.L("Last update"))
	lastUpdateHeader.TextStyle.Bold = true
	lastUpdateHeaderSize := lastUpdateHeader.MinSize().Width + testIconSize
	testLastUpdateLabelSize := widget.NewLabel("XXXX-YY-ZZ").MinSize().Width

	typeHeader := widget.NewLabel(lang.L("Type"))
	typeHeader.TextStyle.Bold = true
	typeHeaderSize := typeHeader.MinSize().Width + testIconSize
	testTypeLabelSize := widget.NewLabel(lang.L("capitalisation")).MinSize().Width

	usageHeader := widget.NewLabel(lang.L("Usage"))
	usageHeader.TextStyle.Bold = true
	usageHeaderSize := usageHeader.MinSize().Width + testIconSize
	testUsageLabelSize := widget.NewLabel(lang.L("PRIV")).MinSize().Width

	IBANHeader := widget.NewLabel(lang.L("IBAN"))
	IBANHeader.TextStyle.Bold = true
	IBANHeaderSize := IBANHeader.MinSize().Width + testIconSize
	testIBANLabelSize := widget.NewLabel("FR76 3000 1007 9412 3456 7890 185").MinSize().Width

	numberHeader := widget.NewLabel(lang.L("Account number"))
	numberHeader.TextStyle.Bold = true
	numberHeaderSize := numberHeader.MinSize().Width + testIconSize
	testNumberLabelSize := widget.NewLabel("550e8400-e29b-41d4-a716-446655440000").MinSize().Width

	// Fill bank accounts. Backend call
	bankAccounts := GetBankAccounts(app, "")

	accountTable := widget.NewTable(
		func() (int, int) {
			return len(bankAccounts), numberOfColumns
		},
		func() fyne.CanvasObject {
			scrollerBankNameLabel := widget.NewLabel("Template")
			scrollerBankNameLabel.Alignment = fyne.TextAlignCenter
			bankNameItem := container.NewHScroll(scrollerBankNameLabel)

			scrollerAccountNameLabel := widget.NewLabel("Template")
			scrollerAccountNameLabel.Alignment = fyne.TextAlignCenter
			accountNameItem := container.NewHScroll(scrollerAccountNameLabel)

			valueItem := widget.NewLabel("Template")
			valueItem.Alignment = fyne.TextAlignCenter
			currencyItem := widget.NewLabel("Template")
			currencyItem.Alignment = fyne.TextAlignCenter
			lastUpdateItem := widget.NewLabel("Template")
			lastUpdateItem.Alignment = fyne.TextAlignCenter
			typeItem := widget.NewLabel("Template")
			typeItem.Alignment = fyne.TextAlignCenter
			usageItem := widget.NewLabel("Template")
			usageItem.Alignment = fyne.TextAlignCenter
			ibanItem := widget.NewLabel("Template")
			ibanItem.Alignment = fyne.TextAlignCenter
			accountNumberItem := widget.NewLabel("Template")
			accountNumberItem.Alignment = fyne.TextAlignCenter

			return container.NewStack(
				bankNameItem,
				accountNameItem,
				valueItem,
				currencyItem,
				lastUpdateItem,
				typeItem,
				usageItem,
				ibanItem,
				accountNumberItem,
			)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			// We parse the previously created items (order is important and defined in the create function and iota)
			bankNameItem := o.(*fyne.Container).Objects[bankNameColumn].(*container.Scroll)
			accountNameItem := o.(*fyne.Container).Objects[accountNameColumn].(*container.Scroll)
			valueItem := o.(*fyne.Container).Objects[valueColumn].(*widget.Label)
			currencyItem := o.(*fyne.Container).Objects[currencyColumn].(*widget.Label)
			lastUpdateItem := o.(*fyne.Container).Objects[lastUpdateColumn].(*widget.Label)
			typeItem := o.(*fyne.Container).Objects[typeColumn].(*widget.Label)
			usageItem := o.(*fyne.Container).Objects[usageColumn].(*widget.Label)
			ibanItem := o.(*fyne.Container).Objects[IBANColumn].(*widget.Label)
			accountNumberItem := o.(*fyne.Container).Objects[accountNumberColumn].(*widget.Label)

			bankNameItem.Hide()
			accountNameItem.Hide()
			valueItem.Hide()
			currencyItem.Hide()
			lastUpdateItem.Hide()
			typeItem.Hide()
			usageItem.Hide()
			ibanItem.Hide()
			accountNumberItem.Hide()

			// Update the cell by adding content according to its "type" (IBAN, accountName, value, currency, lastUpdate, type, usage, number)
			switch id.Col {
			case bankNameColumn:
				accountNameItem.Show()
				accountNameItem.Content.(*widget.Label).SetText(bankAccounts[id.Row].Bank_Original_name)

			case accountNameColumn:
				accountNameItem.Show()
				accountNameItem.Content.(*widget.Label).SetText(bankAccounts[id.Row].Original_name)

			case valueColumn:
				value := fmt.Sprintf("%.2f", bankAccounts[id.Row].Balance)

				if bankAccounts[id.Row].Balance < 0 {
					valueItem.Importance = widget.WarningImportance
				} else {
					valueItem.Importance = widget.MediumImportance
				}
				valueItem.Show()
				valueItem.SetText(helper.ValueSpacer(value))

			case currencyColumn:
				currencyItem.Show()
				currencyItem.SetText(bankAccounts[id.Row].Currency)

			case lastUpdateColumn:
				parsedTxDate, err := time.Parse("2006-01-02 15:04:05", bankAccounts[id.Row].Last_update)
				if err != nil {
					helper.Logger.Error().Err(err).Msgf("Cannot parse date %s", bankAccounts[id.Row].Last_update)
				}

				lastUpdateItem.Show()
				lastUpdateItem.SetText(parsedTxDate.Format("2006-01-02"))

			case typeColumn:
				// Each type is here: https://docs.powens.com/api-reference/products/data-aggregation/bank-account-types#accounttypename-values
				typeItem.Show()
				typeItem.SetText(lang.L(bankAccounts[id.Row].Account_type))

			case usageColumn:
				if bankAccounts[id.Row].Usage != "" {
					usageItem.Show()
					usageItem.SetText(lang.L(bankAccounts[id.Row].Usage))
				}

			case IBANColumn:
				ibanItem.Show()
				ibanItem.SetText(ibanSpacer(bankAccounts[id.Row].Iban))

			case accountNumberColumn:
				accountNumberItem.Show()
				accountNumberItem.SetText(bankAccounts[id.Row].Number)

			default:
				helper.Logger.Fatal().Msg("Too much column in the account grid")
			}
		},
	)

	// Set column header, sortable when taped https://fynelabs.com/2023/10/05/user-data-sorting-with-a-fyne-table-widget/
	accountTable.ShowHeaderRow = true
	accountTable.CreateHeader = func() fyne.CanvasObject {
		return widget.NewButton("000", func() {})
	}

	accountTable.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		b := o.(*widget.Button)

		switch id.Col {
		case bankNameColumn:
			b.SetText(lang.L("Name"))
			helper.SetColumnHeaderIcon(columnSort[bankNameColumn], b, sortAsc, sortDesc)

		case accountNameColumn:
			b.SetText(lang.L("Account name"))
			helper.SetColumnHeaderIcon(columnSort[accountNameColumn], b, sortAsc, sortDesc)

		case valueColumn:
			b.SetText(lang.L("Value"))
			helper.SetColumnHeaderIcon(columnSort[valueColumn], b, sortAsc, sortDesc)

		case currencyColumn:
			b.SetText(lang.L("Currency"))
			helper.SetColumnHeaderIcon(columnSort[currencyColumn], b, sortAsc, sortDesc)

		case lastUpdateColumn:
			b.SetText(lang.L("Last update"))
			helper.SetColumnHeaderIcon(columnSort[lastUpdateColumn], b, sortAsc, sortDesc)

		case typeColumn:
			b.SetText(lang.L("Type"))
			helper.SetColumnHeaderIcon(columnSort[typeColumn], b, sortAsc, sortDesc)

		case usageColumn:
			b.SetText(lang.L("Usage"))
			helper.SetColumnHeaderIcon(columnSort[usageColumn], b, sortAsc, sortDesc)

		case IBANColumn:
			b.SetText(lang.L("IBAN"))
			helper.SetColumnHeaderIcon(columnSort[IBANColumn], b, sortAsc, sortDesc)

		case accountNumberColumn:
			b.SetText(lang.L("Account number"))
			helper.SetColumnHeaderIcon(columnSort[accountNumberColumn], b, sortAsc, sortDesc)

		default:
			helper.Logger.Fatal().Msg("Too much column in the grid for account header")
		}

		b.OnTapped = func() {
			applySort(id.Col, accountTable, bankAccounts)
		}

		b.Refresh()
	}

	accountTable.OnSelected = func(id widget.TableCellID) {
		go func() {
			time.Sleep(unselectTime)
			fyne.Do(func() {
				accountTable.Unselect(id)
			})
		}()
	}

	// We set the width of the columns, ie the max between the language name header size and actual value
	// For example, the max between "Value" and "-123456123.00", or "Montant" and "-123456123.00" in french
	accountTable.SetColumnWidth(bankNameColumn, float32(math.Max(float64(testBankNameLabelSize), float64(bankNameHeaderSize))))
	accountTable.SetColumnWidth(accountNameColumn, float32(math.Max(float64(testAccountNameLabelSize), float64(accountNameHeaderSize))))
	accountTable.SetColumnWidth(valueColumn, float32(math.Max(float64(testValueLabelSize), float64(valueHeaderSize))))
	accountTable.SetColumnWidth(currencyColumn, float32(math.Max(float64(testCurrencyLabelSize), float64(currencyHeaderSize))))
	accountTable.SetColumnWidth(lastUpdateColumn, float32(math.Max(float64(testLastUpdateLabelSize), float64(lastUpdateHeaderSize))))
	accountTable.SetColumnWidth(typeColumn, float32(math.Max(float64(testTypeLabelSize), float64(typeHeaderSize))))
	accountTable.SetColumnWidth(usageColumn, float32(math.Max(float64(testUsageLabelSize), float64(usageHeaderSize))))
	accountTable.SetColumnWidth(IBANColumn, float32(math.Max(float64(testIBANLabelSize), float64(IBANHeaderSize))))
	accountTable.SetColumnWidth(accountNumberColumn, float32(math.Max(float64(testNumberLabelSize), float64(numberHeaderSize))))

	// Reload button reloads data by querying the backend
	reloadButton := widget.NewButton("", func() {
		bankAccounts = GetBankAccounts(app, "")
		accountTable.Refresh()

		// Reset header sorting if any
		columnSort[0] = numberOfSorts
		applySort(0, accountTable, bankAccounts)
	})

	reloadButton.Icon = theme.ViewRefreshIcon()

	return container.NewBorder(nil, container.NewBorder(nil, nil, nil, reloadButton, nil), nil, nil, accountTable)

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

// Sort table data
func applySort(col int, t *widget.Table, data []BankAccount) {

	// Circle sorting: off => asc => desc => off => etc...
	order := columnSort[col]
	order++
	if order >= numberOfSorts {
		order = sortOff
	}

	// Reset all and assign tapped sort
	for i := range numberOfColumns {
		columnSort[i] = sortOff
	}

	columnSort[col] = order

	sort.Slice(data, func(i, j int) bool {
		a := data[i]
		b := data[j]

		// re-sort with no sort selected
		if order == sortOff {
			return a.Bank_Original_name < b.Bank_Original_name
		}

		switch col {
		case bankNameColumn:
			if order == sortAsc {
				return a.Bank_Original_name < b.Bank_Original_name
			}
			return a.Bank_Original_name > b.Bank_Original_name

		case accountNameColumn:
			if order == sortAsc {
				return a.Original_name < b.Original_name
			}
			return a.Original_name > b.Original_name

		case valueColumn:
			if order == sortAsc {
				return a.Balance > b.Balance
			}
			return a.Balance < b.Balance

		case currencyColumn:
			if order == sortDesc {
				return a.Currency > b.Currency
			}
			return a.Currency < b.Currency

		case lastUpdateColumn:
			if order == sortDesc {
				return a.Last_update > b.Last_update
			}
			return a.Last_update < b.Last_update

		case typeColumn:
			if order == sortDesc {
				return a.Account_type > b.Account_type
			}
			return a.Account_type < b.Account_type

		case usageColumn:
			if order == sortDesc {
				return a.Usage > b.Usage
			}
			return a.Usage < b.Usage

		case IBANColumn:
			if order == sortDesc {
				return a.Iban > b.Iban
			}
			return a.Iban < b.Iban

		case accountNumberColumn:
			if order == sortDesc {
				return a.Number > b.Number
			}
			return a.Number < b.Number
		default:
			return false
		}
	})
	t.Refresh()
}
