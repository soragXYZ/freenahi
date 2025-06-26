package bank

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"financialApp/config"
)

func GetAccounts(w http.ResponseWriter, r *http.Request) {

	var accounts []BankAccount

	accountType := r.URL.Query().Get("type")

	var rows *sql.Rows
	var err error
	var query string

	if accountType == "" {
		query = "SELECT * FROM bankAccount ORDER BY original_name"
		rows, err = config.DB.Query(query)

	} else { // filter by account type if parameter is set

		switch accountType { // https://docs.powens.com/api-reference/products/data-aggregation/bank-account-types#accounttypename-values
		case "article83", "capitalisation", "card", "checking",
			"crowdlending", "deposit", "ldds", "lifeinsurance",
			"loan", "madelin", "market", "pea", "pee", "per",
			"perco", "perp", "real_estate", "rsp", "savings", "unknown":

			query = "SELECT * FROM bankAccount WHERE account_type=? ORDER BY balance DESC"
			rows, err = config.DB.Query(query, accountType)

		default:
			config.Logger.Warn().Str("type", accountType).Msg("Unsupported Powens account type")
			http.Error(w,
				"Unsupported account type. Must be: article83, capitalisation, card, checking,"+
					"crowdlending, deposit, ldds, lifeinsurance,"+
					"loan, madelin, market, pea, pee, per,"+
					"perco, perp, real_estate, rsp, savings, unknown",
				http.StatusBadRequest)
			return
		}
	}

	if err != nil {
		config.Logger.Error().Err(err).Msg(query)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var account BankAccount
		if err := rows.Scan(&account.Account_id, &account.User_id, &account.Bank_Original_name, &account.Number, &account.Original_name, &account.Balance, &account.Last_update, &account.Iban, &account.Currency, &account.Account_type, &account.Usage); err != nil {
			config.Logger.Error().Err(err).Msg("Cannot scan row")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		config.Logger.Error().Err(err).Msg("Error in rows")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	jsonBody, err := json.Marshal(accounts)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal accounts")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)
}

// Return the total amount for every type of bankAccount
func GetAccountSum(w http.ResponseWriter, r *http.Request) {

	var accountSums []BankAccountSum

	var rows *sql.Rows
	var err error

	query := "SELECT account_type, SUM(balance) FROM bankAccount GROUP BY account_type"
	rows, err = config.DB.Query(query)

	if err != nil {
		config.Logger.Error().Err(err).Msg(query)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var account BankAccountSum
		if err := rows.Scan(&account.Account_type, &account.Value); err != nil {
			config.Logger.Error().Err(err).Msg("Cannot scan row")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		accountSums = append(accountSums, account)
	}
	if err := rows.Err(); err != nil {
		config.Logger.Error().Err(err).Msg("Error in rows")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// Some types are identical for market, group them
	groupedAccountSums := []BankAccountSum{
		{Account_type: "Stocks and funds"},
		{Account_type: "Bank accounts"},
		{Account_type: "Savings books"},
	}

	config.Logger.Trace().Msg("Grouping sums")

	for _, accountSum := range accountSums {
		config.Logger.Trace().Str("type", accountSum.Account_type).Float32("value", accountSum.Value).Msg("")

		switch accountSum.Account_type {
		case "article83", "capitalisation", "crowdlending", "lifeinsurance", "madelin", "market", "pea", "pee", "per", "perco", "perp", "rsp":
			groupedAccountSums[0].Value += accountSum.Value
		case "checking":
			groupedAccountSums[1].Value += accountSum.Value
		case "savings":
			groupedAccountSums[2].Value += accountSum.Value
		}
	}

	jsonBody, err := json.Marshal(groupedAccountSums)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal accounts")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)
}
