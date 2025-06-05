* Backend
- DB datetime datetime auto / TIMESTAMP ? DEFAULT CURRENT_TIMESTAMP ???
- Write unit tests
- Form validation ? "github.com/go-playground/validator/v10"
- godoc / swagger ?
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
- Add a tutorial to create the backend and Powens account
- Tx: add filter and reload options
- Do not load everything as start up (Tabs mechanism), use tab on selected to fill the data ?
- Add a possibility to export as PDF ?
- Add tooltips ? https://github.com/dweymouth/fyne-tooltip
- Refactor backend calls ? Always same structure, just URL and data type changing
