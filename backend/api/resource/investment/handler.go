package investment

import (
	"encoding/json"
	"net/http"

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

func GetInvestmentsHistory(w http.ResponseWriter, r *http.Request) {

	var historyInvestments []HistoryInvestment

	var query string = "SELECT invest_id, valuation, date_valuation FROM historyInvest"
	rows, err := config.DB.Query(query)
	if err != nil {
		config.Logger.Error().Err(err).Msg(query)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var historyInvestment HistoryInvestment
		if err := rows.Scan(&historyInvestment.Invest_id, &historyInvestment.Valuation, &historyInvestment.Date_valuation); err != nil {
			config.Logger.Error().Err(err).Msg("Cannot scan row")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		historyInvestments = append(historyInvestments, historyInvestment)
	}
	if err := rows.Err(); err != nil {
		config.Logger.Error().Err(err).Msg("Error in rows")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	jsonBody, err := json.Marshal(historyInvestments)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Cannot marshal historyInvestments")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBody)
}
