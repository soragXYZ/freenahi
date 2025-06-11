package webview

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"financialApp/config"
)

func GetManageLink(w http.ResponseWriter, r *http.Request) {

	var code authCode = getTemporaryToken()
	if code.Error != nil {
		config.Logger.Error().Err(code.Error).Msg(code.ErrorString)

		if code.ErrorString == "Permanent user token does not exist" {
			http.Error(w, code.ErrorString, http.StatusNotFound)
			return
		}

		http.Error(w, code.ErrorString, http.StatusInternalServerError)
		return
	}

	// The URL is builded as indicated here
	// https://docs.powens.com/api-reference/overview/webview#manage-connections
	var powensURL string = config.Conf.Powens.WebviewUrl + config.Conf.Other.Language + "/manage"
	var domain string = "domain=" + config.Conf.Powens.Domain + "-sandbox"
	var client_id string = "client_id=" + config.Conf.Powens.ClientId
	var redirect_uri string = "redirect_uri=" + config.Conf.Powens.RedirectUrl
	var connector_capabilities string = "connector_capabilities=bank,bankwealth"

	var connexionUrl string = powensURL + "?" + domain + "&" + client_id + "&" + redirect_uri + "&" + connector_capabilities + "&code=" + code.Auth_Code

	w.Write([]byte(connexionUrl))
}

func getTemporaryToken() authCode {

	// Check if one permanent user token already exists in DB or not, as it s needed to create a temporary token
	var permanentUserToken string
	var query string = "SELECT auth_token FROM authToken LIMIT 1"
	row := config.DB.QueryRow(query)
	if err := row.Scan(&permanentUserToken); err != nil {

		if err == sql.ErrNoRows {
			return authCode{
				Auth_Code:   "",
				Error:       errors.New("no permanent user token"),
				ErrorString: "Permanent user token does not exist",
			}
		}

		return authCode{Auth_Code: "", Error: err, ErrorString: query}
	}

	config.Logger.Trace().Str("permanent_user_code", permanentUserToken).Msg("")

	// Get a temporary user token from Powens API
	var url string = "https://" + config.Conf.Powens.Domain + "-sandbox.biapi.pro/2.0/auth/token/code"
	var bearer string = "Bearer " + permanentUserToken

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return authCode{Auth_Code: "", Error: err, ErrorString: "Cannot create new get request"}
	}
	req.Header.Add("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return authCode{Auth_Code: "", Error: err, ErrorString: "Cannot execute request"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorString string = "Powens did not answer correctly: " + string(resp.Status)
		return authCode{Auth_Code: "", Error: nil, ErrorString: errorString}
	}

	var code authCode

	err = json.NewDecoder(resp.Body).Decode(&code)
	if err != nil {
		return authCode{Auth_Code: "", Error: err, ErrorString: "Cannot decode resp.Body"}
	}

	config.Logger.Trace().Str("temporary_code", code.Auth_Code).Msg("")
	return code
}
