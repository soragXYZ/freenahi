package investment

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"time"

	"financialApp/config"
)

// Get invests ordered by valuation (DESC)
func GetInvestments(w http.ResponseWriter, r *http.Request) {

	var investments []Investment

	// Invest_id, Account_id, Label, Code, Code_type, Stock_symbol, Quantity, Unit_price, Unit_value, Valuation, Diff, Diff_percent, Last_update
	var query string = "SELECT invest.invest_id, invest.account_id, invest.invest_label, invest.invest_code, invest.invest_code_type, invest.stock_symbol, invest.quantity, invest.unit_price, invest.unit_value, invest.valuation, invest.diff, invest.diff_percent, invest.last_update, bankAccount.bank_original_name, bankAccount.original_name FROM invest INNER JOIN bankAccount ON invest.account_id = bankAccount.account_id ORDER BY valuation DESC"
	rows, err := config.DB.Query(query)
	if err != nil {
		config.Logger.Error().Err(err).Msg(query)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var investment Investment
		if err := rows.Scan(&investment.Invest_id, &investment.Account_id, &investment.Label, &investment.Code, &investment.Code_type, &investment.Stock_symbol, &investment.Quantity, &investment.Unit_price, &investment.Unit_value, &investment.Valuation, &investment.Diff, &investment.Diff_percent, &investment.Last_update, &investment.BankOriginalName, &investment.OriginalName); err != nil {
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

	// ToDo: parse and add validation with "github.com/go-playground/validator/v10"
	period := r.URL.Query().Get("period")
	accountType := r.URL.Query().Get("type")

	var rows *sql.Rows
	var err error
	var query string

	switch period {
	case "", "all": // Get each point
		query = "SELECT historyValue.bank_account_id, historyValue.valuation, historyValue.date_valuation FROM historyValue INNER JOIN bankAccount ON historyValue.bank_account_id = bankAccount.account_id AND bankAccount.account_type=? ORDER BY historyValue.date_valuation"
		rows, err = config.DB.Query(query, accountType)

	case "month": // Get values which are 1 month old MAX
		query = "SELECT historyValue.bank_account_id, historyValue.valuation, historyValue.date_valuation FROM historyValue INNER JOIN bankAccount ON historyValue.bank_account_id = bankAccount.account_id AND bankAccount.account_type=? WHERE historyValue.date_valuation > ? ORDER BY historyValue.date_valuation"
		rows, err = config.DB.Query(query, accountType, time.Now().Add(-31*24*time.Hour))

	case "year": // Get values which are 1 year old MAX
		query = "SELECT historyValue.bank_account_id, historyValue.valuation, historyValue.date_valuation FROM historyValue INNER JOIN bankAccount ON historyValue.bank_account_id = bankAccount.account_id AND bankAccount.account_type=? WHERE historyValue.date_valuation > ? ORDER BY historyValue.date_valuation"
		rows, err = config.DB.Query(query, accountType, time.Now().Add(-365*24*time.Hour))
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

	// Get every bank account id registered and remove duplicate values
	var bankAccountIds []int
	for _, point := range historyValues {
		bankAccountIds = append(bankAccountIds, point.BankAccountId)
	}

	slices.Sort(bankAccountIds)
	bankAccountIds = slices.Compact(bankAccountIds)

	// Create a map: key = date, value = bank / stock / savings valuation :
	// Points are sorted by date in the previous SQL query, so the first item has the longest history
	// We create a firt
	// Then, we loop with every other bankAccountId, and we sum the value for the current date
	// The goal is to know the global (ie summed) valuation for every date
	// MAybe issue if they don't have the same length, ie some bank accounts points do not end the same day

	constructedHistoryValues := make(map[time.Time]float32)

	for _, bankAccountId := range bankAccountIds {

		nextHistoryValues := generateInitialValueDatePairs(bankAccountId, historyValues)

		for key, value := range nextHistoryValues {

			if constructedHistoryValues[key] != 0 {
				constructedHistoryValues[key] += value
			} else {
				constructedHistoryValues[key] = value
			}
		}
	}

	var pointValues []HistoryValuePoint

	// Reorder the maping by date
	var keys []time.Time

	for key := range constructedHistoryValues {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	for _, value := range keys {
		pointValues = append(pointValues, HistoryValuePoint{
			DateValuation: value,
			Valuation:     constructedHistoryValues[value],
		})
	}

	jsonBody, err := json.Marshal(pointValues)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal pointValues")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)

}

// Returns a list of value/Date pairs for a given account. Used to display graphs
func ReadHistoryValue(w http.ResponseWriter, r *http.Request) {

	// ToDo: parse and add validation with "github.com/go-playground/validator/v10"
	bankAccountId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot convert id parameter to from str to int")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	period := r.URL.Query().Get("period")

	var rows *sql.Rows
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

	// Create a list of value/date pairs.
	constructedHistoryValues := generateInitialValueDatePairs(bankAccountId, historyValues)

	var pointValues []HistoryValuePoint

	// Reorder the maping by date
	var keys []time.Time

	for key := range constructedHistoryValues {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Before(keys[j])
	})

	for _, value := range keys {
		pointValues = append(pointValues, HistoryValuePoint{
			DateValuation: value,
			Valuation:     constructedHistoryValues[value],
		})
	}

	jsonBody, err := json.Marshal(pointValues)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal constructedHistoryValues")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)

}

// Create a list of value/date pairs from historical data for the specified bank account.
func generateInitialValueDatePairs(bankAccountId int, historyValues []HistoryValue) map[time.Time]float32 {

	mappedValues := make(map[time.Time]float32)

	// Get the first value
	var previousPoint HistoryValue
	for index, point := range historyValues {
		if point.BankAccountId != bankAccountId {
			continue
		}
		previousPoint = historyValues[index]
		break
	}

	for _, point := range historyValues {

		// Only take into account the requested bank account id
		if point.BankAccountId != bankAccountId {
			continue
		}

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

			mappedValues[previousTime.Add(24*time.Hour*time.Duration(daysDiff))] = previousPoint.Valuation

			daysDiff += 1
		}

		previousPoint = point
	}

	// Add last value of the given bankAccountId, since we miss 1 iteration in the previous loop
	for i := len(historyValues) - 1; i >= 0; i-- {

		// We need to add the last element. Loop from the end of the list and add the correct element
		if historyValues[i].BankAccountId != bankAccountId {
			continue
		}

		lastTime, err := time.Parse("2006-01-02", historyValues[i].DateValuation)
		if err != nil {
			config.Logger.Error().Err(err).Msgf("Cannot parse date %s", historyValues[i].DateValuation)
		}
		mappedValues[lastTime] = historyValues[i].Valuation

		extendMapToToday(lastTime, historyValues[i].Valuation, mappedValues)

		break
	}

	return mappedValues
}

// Extend the mapping to today if some points are missing.
// Extend with the last known value for every point
func extendMapToToday(lastDate time.Time, lastValuation float32, mappedValues map[time.Time]float32) map[time.Time]float32 {

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	daysDiff := 0

	for lastDate.Add(24 * time.Hour * time.Duration(daysDiff)).Before(today) {

		mappedValues[lastDate.Add(24*time.Hour*time.Duration(daysDiff+1))] = lastValuation
		daysDiff += 1
	}

	return mappedValues
}
