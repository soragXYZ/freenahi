package loan

// Models taken from https://docs.powens.com/api-reference/products/data-aggregation/bank-accounts#data-model

// Time sent by Powens API is not RFC3339
// so we store it as string
// https://stackoverflow.com/questions/25087960/json-unmarshal-time-that-isnt-in-rfc-3339-format

// https://docs.powens.com/api-reference/products/data-aggregation/bank-accounts#loan-object
type Loan struct {
	Loan_account_id      int     `json:"-"` // absent in base data, field added for simplicity
	Total_amount         float32 `json:"total_amount"`
	Available_amount     float32 `json:"available_amount"`
	Used_amount          float32 `json:"used_amount"`
	Subscription_date    string  `json:"subscription_date"`
	Maturity_date        string  `json:"maturity_date"`
	Start_repayment_date string  `json:"start_repayment_date"`
	Deferred             bool    `json:"deferred"`
	Next_payment_amount  float32 `json:"next_payment_amount"`
	Next_payment_date    string  `json:"next_payment_date"`
	Rate                 float32 `json:"rate"`
	Nb_payments_left     uint    `json:"nb_payments_left"`
	Nb_payments_done     uint    `json:"nb_payments_done"`
	Nb_payments_total    uint    `json:"nb_payments_total"`
	Last_payment_amount  float32 `json:"last_payment_amount"`
	Last_payment_date    string  `json:"last_payment_date"`
	Account_label        string  `json:"account_label"`
	Insurance_label      string  `json:"insurance_label"`
	Insurance_amount     float32 `json:"insurance_amount"`
	Insurance_rate       float32 `json:"insurance_rate"`
	Duration             uint    `json:"duration"`
	Loan_type            string  `json:"type"`
}
