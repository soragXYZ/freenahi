package bank

import (
	"financialApp/api/resource/investment"
	"financialApp/api/resource/loan"
	"financialApp/api/resource/miscellaneous"
	"financialApp/api/resource/transaction"
)

// Models taken from https://docs.powens.com/api-reference/products/data-aggregation/bank-accounts#data-model

// Time sent by Powens API is not RFC3339
// so we store it as string
// https://stackoverflow.com/questions/25087960/json-unmarshal-time-that-isnt-in-rfc-3339-format

// https://docs.powens.com/api-reference/products/data-aggregation/bank-accounts#bankaccount-object
type BankAccount struct {
	Account_id         int     `json:"id"`
	User_id            int     `json:"id_user"`
	Number             string  `json:"number"`
	Bank_Original_name string  `json:"bank_original_name"`
	Original_name      string  `json:"original_name"`
	Balance            float32 `json:"balance"`
	Last_update        string  `json:"last_update"`
	Iban               string  `json:"iban"`
	Currency           string  `json:"currency"`
	Account_type       string  `json:"type"`
	Error              string  `json:"error"` // not needed ?
	Usage              string  `json:"usage"`
}

type BankAccountWebhook struct {
	Account_id    int                       `json:"id"`
	User_id       int                       `json:"id_user"`
	Number        string                    `json:"number"`
	Original_name string                    `json:"original_name"`
	Balance       float32                   `json:"balance"`
	Last_update   string                    `json:"last_update"`
	Iban          string                    `json:"iban"`
	Currency      miscellaneous.Currency    `json:"currency"`
	Account_type  string                    `json:"type"`
	Error         string                    `json:"error"` // not needed ?
	Usage         string                    `json:"usage"`
	Loan          loan.Loan                 `json:"loan"`
	Investments   []investment.Investment   `json:"investments"`
	Transactions  []transaction.Transaction `json:"transactions"`
}

type Connector struct {
	Bank_id int    `json:"id"`
	Name    string `json:"name"`
}
