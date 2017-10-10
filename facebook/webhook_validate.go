package facebook

import (
	"errors"
	"net/url"
	"net/http"
)

// @todo: use a custom handler that passes the bot as well
// HandleValidateWebhook tries to validate the Facebook webhook
// More information here: https://developers.facebook.com/docs/messenger-platform/getting-started/quick-start
func HandleValidateWebhook(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	hubMode, err := getSingleQueryParam(queryParams, "hub.mode");

	if err != nil || hubMode != "subscribe" {
		return
	}

	verifyToken, err := getSingleQueryParam(queryParams, "hub.verify_token");

	// @todo: use config here
	if err != nil || verifyToken != "app_verify_token" {
		return
	}

	challenge, err := getSingleQueryParam(queryParams, "hub.challenge");

	if err != nil {
		return
	}

	// Validate the webhook by writing back the "hub.challenge" query param
	w.Write([]byte(challenge))
}

// getSingleQueryParam fetches a single query param using the given url values
func getSingleQueryParam(values url.Values, key string) (string, error) {
	params, ok := values[key]

	if (!ok || len(params) != 1) {
		return "", errors.New("Could not fetch param")
	}

	return params[0], nil
}
