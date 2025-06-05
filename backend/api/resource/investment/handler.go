package investment

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"financialApp/config"
)

func GetInvestments(w http.ResponseWriter, r *http.Request) {

	var investments []Investment

	var query string = "SELECT * FROM invest ORDER BY valuation DESC"
	rows, err := config.DB.Query(query)
	if err != nil {
		config.Logger.Error().Err(err).Msg(query)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var investment Investment
		if err := rows.Scan(&investment.Invest_id, &investment.Account_id, &investment.Label, &investment.Code, &investment.Code_type, &investment.Stock_symbol, &investment.Quantity, &investment.Unit_price, &investment.Unit_value, &investment.Valuation, &investment.Diff, &investment.Diff_percent, &investment.Last_update); err != nil {
			config.Logger.Error().Err(err).Msg("Cannot scan row")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		investments = append(investments, investment)
	}
	if err := rows.Err(); err != nil {
		config.Logger.Error().Err(err).Msg("Error in rows")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	jsonBody, err := json.Marshal(investments)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal investments")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)
}

func ReadHistoryValues(w http.ResponseWriter, r *http.Request) {

	var historyValues []HistoryValue

	var query string = "SELECT bank_account_id, valuation, date_valuation FROM historyValue"
	rows, err := config.DB.Query(query)
	if err != nil {
		config.Logger.Error().Err(err).Msg(query)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var historyValue HistoryValue
		if err := rows.Scan(&historyValue.BankAccountId, &historyValue.Valuation, &historyValue.DateValuation); err != nil {
			config.Logger.Error().Err(err).Msg("Cannot scan row")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		historyValues = append(historyValues, historyValue)
	}
	if err := rows.Err(); err != nil {
		config.Logger.Error().Err(err).Msg("Error in rows")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	jsonBody, err := json.Marshal(historyValues)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal historyValues")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)
}

// Returns a list of value/Date pairs for a given account. Used to display graphs
func ReadHistoryValue(w http.ResponseWriter, r *http.Request) {

	// ToDo: parse and add validation with "github.com/go-playground/validator/v10"
	bankAccountId := r.PathValue("id")
	period := r.URL.Query().Get("period")

	var rows *sql.Rows
	var err error
	var query string

	switch period {
	case "", "all": // Get each point
		query = "SELECT bank_account_id, valuation, date_valuation FROM historyValue WHERE bank_account_id = ? ORDER BY date_valuation"
		rows, err = config.DB.Query(query, bankAccountId)

	case "month": // Get values which are 1 month old MAX
		query = "SELECT bank_account_id, valuation, date_valuation FROM historyValue WHERE bank_account_id = ? AND date_valuation > ? ORDER BY date_valuation"
		rows, err = config.DB.Query(query, bankAccountId, time.Now().Add(-31*24*time.Hour))

	case "year": // Get values which are 1 year old MAX
		query = "SELECT bank_account_id, valuation, date_valuation FROM historyValue WHERE bank_account_id = ? AND date_valuation > ? ORDER BY date_valuation"
		rows, err = config.DB.Query(query, bankAccountId, time.Now().Add(-365*24*time.Hour))
	}

	if err != nil {
		config.Logger.Error().Err(err).Msg(query)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var historyValues []HistoryValue

	for rows.Next() {
		var historyValue HistoryValue
		if err := rows.Scan(&historyValue.BankAccountId, &historyValue.Valuation, &historyValue.DateValuation); err != nil {
			config.Logger.Error().Err(err).Msg("Cannot scan row")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		historyValues = append(historyValues, historyValue)
	}

	if err := rows.Err(); err != nil {
		config.Logger.Error().Err(err).Msg("Error in rows")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if len(historyValues) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Next, we create a list of value/date pairs for the asked duration.
	var constructedHistoryValues []HistoryValuePoint

	previousPoint := historyValues[0]
	for _, point := range historyValues {

		previousTime, err := time.Parse("2006-01-02", previousPoint.DateValuation)
		if err != nil {
			config.Logger.Error().Err(err).Msgf("Cannot parse date %s", point.DateValuation)
		}

		parsedTime, err := time.Parse("2006-01-02", point.DateValuation)
		if err != nil {
			config.Logger.Error().Err(err).Msgf("Cannot parse date %s", point.DateValuation)
		}

		daysDiff := 0

		// If some days are missing (for example Powens was down for some time, or if we missed the wehbook),
		// we fill the gap with the previously known value
		for previousTime.Add(24 * time.Hour * time.Duration(daysDiff)).Before(parsedTime) {

			constructedHistoryValues = append(constructedHistoryValues, HistoryValuePoint{
				Valuation:     point.Valuation,
				DateValuation: previousTime.Add(24 * time.Hour * time.Duration(daysDiff)),
			})
			daysDiff += 1
		}

		previousPoint = point
	}

	// Add last value
	lastTime, err := time.Parse("2006-01-02", historyValues[len(historyValues)-1].DateValuation)
	if err != nil {
		config.Logger.Error().Err(err).Msgf("Cannot parse date %s", historyValues[len(historyValues)-1].DateValuation)
	}
	constructedHistoryValues = append(constructedHistoryValues, HistoryValuePoint{
		Valuation:     historyValues[len(historyValues)-1].Valuation,
		DateValuation: lastTime,
	})

	jsonBody, err := json.Marshal(constructedHistoryValues)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal constructedHistoryValues")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)

}
