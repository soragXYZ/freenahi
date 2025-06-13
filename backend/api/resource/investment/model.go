package investment

import "time"

// Models taken from https://docs.powens.com/api-reference/products/wealth-aggregation/investments#data-model

// https://docs.powens.com/api-reference/products/wealth-aggregation/investments#investment-object
type Investment struct {
	Invest_id        int     `json:"id"`
	Account_id       int     `json:"id_account"`
	Label            string  `json:"label"`
	Code             string  `json:"code"`
	Code_type        string  `json:"code_type"`
	Stock_symbol     string  `json:"stock_symbol"`
	Quantity         float32 `json:"quantity"`
	Unit_price       float32 `json:"unitprice"`
	Unit_value       float32 `json:"unitvalue"`
	Valuation        float32 `json:"valuation"`
	Diff             float32 `json:"diff"`
	Diff_percent     float32 `json:"diff_percent"`
	Last_update      string  `json:"last_update"`
	BankOriginalName string  `json:"bank_original_name"` // not present in base data, field added for simplicity
	OriginalName     string  `json:"original_name"`      // not present in base data, field added for simplicity
}

type HistoryValue struct {
	History_id    int
	BankAccountId int
	Valuation     float32
	DateValuation string
}

type HistoryValuePoint struct {
	Valuation     float32
	DateValuation time.Time
}
