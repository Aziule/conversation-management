package bot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/aziule/conversation-management/conversation"
)

// ReceiveMessage is called when a new message is sent by the user to the page
func (bot *Bot) HandleMessageReceived(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Could not parse the request body", 500)
		return
	}

	message, err := conversation.NewMessageFromJson(body)
	fmt.Println(message)
	fmt.Println(message.SenderId())
	fmt.Println(message.RecipientId())

	//parser := &nlu.Parser{}
	//parsed, _ := parser.ParseText(m.Text)
	//
	//user := &FacebookUser{
	//	uuid: "uuid",
	//	fbid: "fbid",
	//	name: "Raoul",
	//}

	//entrypoint := data.GetDummyEntrypoint()
	//
	//for _, startingStep := range entrypoint.Stories()[0].StartingSteps() {
	//	fmt.Println(startingStep.Name())
	//}

	//conversation.Progress(user, parsed)
}

// Validate tries to validate the Facebook webhook
// More information here: https://developers.facebook.com/docs/messenger-platform/getting-started/quick-start
func (bot *Bot) HandleValidateWebhook(w http.ResponseWriter, r *http.Request) {
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