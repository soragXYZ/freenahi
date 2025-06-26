* Backend
- DB datetime datetime auto / TIMESTAMP ? DEFAULT CURRENT_TIMESTAMP ???
- Write unit tests
- Form validation ? "github.com/go-playground/validator/v10"
- godoc / swagger ? postman collection ?
- Verify Bearer token sent from webhooks to authenticate (we only use IP for now)
- See what s happening when an invest is deleted ? For example, swap from 1 ETF to another
- Subrouting for endpoints ?
    https://codewithflash.com/advanced-routing-with-go-122
    https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e
- Need to modify history invest table => use bank account id and not invest id
- Is model bankAccountWebhook usefull ?

* Documentation
- Complete it

* Frontend
- Loans: deal with revolving credit
- Tx: add filter and reload options
- Do not load everything as start up (Tabs mechanism), use tab on selected to fill the data ?
- Add a possibility to export as PDF ?
- Add tooltips ? https://github.com/dweymouth/fyne-tooltip
- Refactor backend calls ? Always same structure, just URL and data type changing
- Wealth view: replace bank original name by its icon
- Reload button for asset should reload graphs
- If backend is down or no value to display, display sth to indicate no value
- Finish general screen: add display value for each type (list / table ?)
