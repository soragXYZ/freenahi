package router

import (
	"net/http"

	"financialApp/api/resource/auth"
	"financialApp/api/resource/bank"
	"financialApp/api/resource/investment"
	"financialApp/api/resource/loan"
	"financialApp/api/resource/miscellaneous"
	"financialApp/api/resource/transaction"
	"financialApp/api/resource/webhook"
	"financialApp/api/resource/webview"

	"financialApp/api/router/middleware"
)

func New() *http.ServeMux {

	// to do: dispatch routes in submodules
	// https://dev.to/kengowada/go-routing-101-handling-and-grouping-routes-with-nethttp-4k0e

	router := http.NewServeMux()

	router.HandleFunc("GET /health/", middleware.Log(middleware.Whitelisted(miscellaneous.HealthCheck)))
	router.HandleFunc("/", middleware.Log(middleware.Whitelisted(miscellaneous.NotFound)))

	router.HandleFunc("POST /webhook/connection_synced/", middleware.Log(middleware.Whitelisted(webhook.ConnectionSynced)))

	router.HandleFunc("GET /bank_account/", middleware.Log(middleware.Whitelisted(bank.GetAccounts)))

	router.HandleFunc("GET /investment/", middleware.Log(middleware.Whitelisted(investment.GetInvestments)))
	router.HandleFunc("GET /investment/history/", middleware.Log(middleware.Whitelisted(investment.GetInvestmentsHistory)))

	router.HandleFunc("GET /loan/", middleware.Log(middleware.Whitelisted(loan.GetLoans)))

	router.HandleFunc("GET /transaction/", middleware.Log(middleware.Whitelisted(transaction.GetTransactions)))

	router.HandleFunc("POST /auth/permanentUserToken/", middleware.Log(middleware.Whitelisted(auth.CreatePermanentUserToken)))
	router.HandleFunc("GET /auth/permanentUserToken/", middleware.Log(middleware.Whitelisted(auth.GetPermanentUserToken)))
	router.HandleFunc("DELETE /auth/permanentUserToken/", middleware.Log(middleware.Whitelisted(auth.DeletePermanentUserToken)))

	router.HandleFunc("GET /webview/manageConnectionLink/", middleware.Log(middleware.Whitelisted(webview.GetManageLink)))

	return router
}
