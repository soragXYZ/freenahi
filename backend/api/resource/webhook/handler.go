package webhook

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"financialApp/config"
)

func ConnectionSynced(w http.ResponseWriter, r *http.Request) {

	// // Display the JSON body in plain text, only for debug
	// buf, _ := io.ReadAll(r.Body)
	// rdr1 := io.NopCloser(bytes.NewBuffer(buf))
	// rdr2 := io.NopCloser(bytes.NewBuffer(buf))
	// io.Copy(os.Stdout, rdr1)
	// r.Body = rdr2

	var conn Conn_synced

	err := json.NewDecoder(r.Body).Decode(&conn)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot decode r.Body")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	for _, account := range conn.Connection.Accounts {

		config.Logger.Trace().
			Str("Connector name", conn.Connection.Bank_connector.Name).
			Int("account_id", account.Account_id).
			Str("account_name", account.Original_name).
			Str("last_update", account.Last_update).
			Int("user_id", account.User_id).
			Msg("Account Update")

		// Create bank account if it does not exists. Otherwise, update last_update value
		var query string = "INSERT INTO bankAccount (account_id, user_id, bank_original_name, bank_number, original_name, balance, last_update, iban, currency, account_type, usage_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		query += "ON DUPLICATE KEY UPDATE balance=?, last_update=?, bank_original_name=?"
		_, err = config.DB.Exec(
			query, account.Account_id, account.User_id, conn.Connection.Bank_connector.Name, account.Number, account.Original_name, account.Balance, account.Last_update, account.Iban, account.Currency.Id, account.Account_type, account.Usage,
			account.Balance, account.Last_update, conn.Connection.Bank_connector.Name,
		)
		if err != nil {
			config.Logger.Error().Err(err).Msg(query)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Add the current value of the account to history (used to draw graphs with historical data)
		vals := []any{}
		query = "INSERT INTO historyValue (bank_account_id, valuation, date_valuation) VALUES "
		query += "(?, ?, ?),"
		vals = append(vals, account.Account_id, account.Balance, account.Last_update)
		query = query[0 : len(query)-1]
		_, err = config.DB.Exec(query, vals...)
		if err != nil {
			config.Logger.Error().Err(err).Msg(query)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		// Proceed with loan
		if account.Loan.Total_amount != 0 {
			config.Logger.Trace().
				Str("last_payment_date", account.Loan.Last_payment_date).
				Uint("nb_payments_done", account.Loan.Nb_payments_done).
				Uint("nb_payments_left", account.Loan.Nb_payments_left).
				Str("next_payment_date", account.Loan.Next_payment_date).
				Float32("total_loan_amount", account.Loan.Total_amount).
				Str("loan_type", account.Loan.Loan_type).
				Msg("Loan update")

			// Sometimes Powens does not calculate some values so we do it manually
			if account.Loan.Nb_payments_total == 0 { // Manually calculate how much payments are left to pay

				// Manually calculate duration the ugly way
				t1, err := time.Parse("2006-01-02 15:04:05", account.Loan.Maturity_date)
				if err != nil {
					config.Logger.Error().Err(err).Msgf("Cannot parse date %s", account.Loan.Maturity_date)
				}

				t2, err := time.Parse("2006-01-02 15:04:05", account.Loan.Subscription_date)
				if err != nil {
					config.Logger.Error().Err(err).Msgf("Cannot parse date %s", account.Loan.Subscription_date)
				}

				yearT1, _ := strconv.Atoi(t1.Format("2006"))
				monthT1, _ := strconv.Atoi(t1.Format("01"))
				yearT2, _ := strconv.Atoi(t2.Format("2006"))
				monthT2, _ := strconv.Atoi(t2.Format("01"))

				account.Loan.Nb_payments_total = uint((yearT1-yearT2)*12 + monthT1 - monthT2)
				config.Logger.Trace().
					Str("Subscription_date", account.Loan.Subscription_date).
					Str("Maturity_date", account.Loan.Maturity_date).
					Msgf("Manually calculated Nb_payments_total: %d", account.Loan.Nb_payments_total)

			}

			if account.Loan.Nb_payments_done == 0 {
				account.Loan.Nb_payments_done = account.Loan.Nb_payments_total - account.Loan.Nb_payments_left
				config.Logger.Trace().
					Uint("Nb_payments_total", account.Loan.Nb_payments_total).
					Uint("Nb_payments_left", account.Loan.Nb_payments_left).
					Msgf("Manually calculated Nb_payments_done: %d", account.Loan.Nb_payments_done)
			}

			if account.Loan.Total_amount < 0 {
				account.Loan.Total_amount = -account.Loan.Total_amount
				config.Logger.Trace().Msg("Total_amount was negative. Reverted")
			}

			if account.Loan.Duration == 0 {
				account.Loan.Duration = account.Loan.Nb_payments_total
				config.Logger.Trace().Msg("Duration was not set. Set with Nb_payments_total value")
			}

			// Update the loan table
			query = "INSERT INTO loan (loan_account_id, total_amount, available_amount, used_amount, subscription_date, maturity_date, start_repayment_date, is_deferred, next_payment_amount, next_payment_date, rate, nb_payments_left, nb_payments_done, nb_payments_total, last_payment_amount, last_payment_date, account_label, insurance_label, insurance_amount, insurance_rate, duration, loan_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE total_amount=?, available_amount=?, used_amount=?, next_payment_amount=?, next_payment_date=?, nb_payments_left=?, nb_payments_done=?, nb_payments_total=?, duration=?"
			_, err = config.DB.Exec(query, account.Account_id, account.Loan.Total_amount, account.Loan.Available_amount, account.Loan.Used_amount, account.Loan.Subscription_date, account.Loan.Maturity_date, account.Loan.Start_repayment_date, account.Loan.Deferred, account.Loan.Next_payment_amount, account.Loan.Next_payment_date, account.Loan.Rate, account.Loan.Nb_payments_left, account.Loan.Nb_payments_done, account.Loan.Nb_payments_total, account.Loan.Last_payment_amount, account.Loan.Last_payment_date, account.Loan.Account_label, account.Loan.Insurance_label, account.Loan.Insurance_amount, account.Loan.Insurance_rate, account.Loan.Duration, account.Loan.Loan_type,
				account.Loan.Total_amount, account.Loan.Available_amount, account.Loan.Used_amount, account.Loan.Next_payment_amount, account.Loan.Next_payment_date, account.Loan.Nb_payments_left, account.Loan.Nb_payments_done, account.Loan.Nb_payments_total, account.Loan.Duration)
			if err != nil {
				config.Logger.Error().Err(err).Msg(query)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}

		// Proceed with transactions
		if len(account.Transactions) != 0 {

			// Bulk insert txs
			query = "INSERT INTO tx (tx_id, user_id, account_id, tx_date, tx_value, tx_type, original_wording) VALUES "
			vals := []any{}
			for _, tx := range account.Transactions {

				config.Logger.Trace().
					Int("account_id", tx.Account_id).
					Str("date", tx.Date).
					Str("original_wording", tx.Original_wording).
					Int("tx_id", tx.Id).
					Float32("value", tx.Value).
					Msg("Tx update")

				query += "(?, ?, ?, ?, ?, ?, ?),"
				vals = append(vals, tx.Id, account.User_id, tx.Account_id, tx.Date, tx.Value, tx.Transaction_type, tx.Original_wording)
			}

			// remove last comma
			query = query[0 : len(query)-1]

			_, err := config.DB.Exec(query, vals...)
			if err != nil {
				config.Logger.Error().Err(err).Msg(query)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}

		// Proceed with invests
		if len(account.Investments) != 0 {

			// Bulk insert invests
			query = "INSERT INTO invest (invest_id, account_id, invest_label, invest_code, invest_code_type, stock_symbol, quantity, unit_price, unit_value, valuation, diff, diff_percent, last_update) VALUES "
			vals := []any{}

			for _, invest := range account.Investments {

				config.Logger.Trace().
					Str("account_name", account.Original_name).
					Str("code", invest.Code).
					Int("invest_id", invest.Invest_id).
					Str("label", invest.Label).
					Float32("unit_price", invest.Unit_price).
					Float32("unit_value", invest.Unit_value).
					Float32("valuation", invest.Valuation).
					Msg("Investment update")

				query += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"

				vals = append(vals, invest.Invest_id, invest.Account_id, invest.Label, invest.Code, invest.Code_type, invest.Stock_symbol, invest.Quantity, invest.Unit_price, invest.Unit_value, invest.Valuation, invest.Diff, invest.Diff_percent, invest.Last_update)
			}

			// remove last comma
			query = query[0 : len(query)-1]

			// if duplicate entry, update the field by the new value
			query += "AS new(a, b, c, d, e, f, Nquantity, Nunit_price, Nunit_value, Nvaluation, Ndiff, Ndiff_percent, Nlast_update)"
			query += "ON DUPLICATE KEY UPDATE quantity=Nquantity, unit_price=Nunit_price, unit_value=Nunit_value, valuation=Nvaluation, diff=Ndiff, diff_percent=Ndiff_percent, last_update=Nlast_update"

			_, err := config.DB.Exec(query, vals...)
			if err != nil {
				config.Logger.Error().Err(err).Msg(query)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}
	}
}
