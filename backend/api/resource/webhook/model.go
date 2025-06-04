package webhook

import (
	"financialApp/api/resource/bank"
	"financialApp/api/resource/miscellaneous"
)

// Models taken from https://docs.powens.com/api-reference/user-connections/connections#connection-synced

type Connection struct {
	Id             int                       `json:"id"`
	Id_user        int                       `json:"id_user"`
	Id_connector   int                       `json:"id_connector"`
	Bank_connector bank.Connector            `json:"connector"`
	Accounts       []bank.BankAccountWebhook `json:"accounts"`
}

type Conn_synced struct {
	User       miscellaneous.User `json:"user"`
	Connection Connection         `json:"connection"`
}
