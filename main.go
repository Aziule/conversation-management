package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aziule/conversation-management/core"
	"github.com/aziule/conversation-management/core/nlu"
	"github.com/aziule/conversation-management/bot/facebook"
	"github.com/aziule/conversation-management/nlu/rasa"
	"github.com/go-chi/chi"
)

var configFlagPath = flag.String("config", "config.json", "Config file path")

func main() {
	flag.Parse()

	config, err := core.LoadConfig(*configFlagPath)

	if err != nil {
		panic(err)
	}

	bot := facebook.NewFacebookBot(config)

	nlu.RegisterFactory("rasa_nlu", rasa.NewRasaNluParser)

	r := chi.NewRouter()

	// Automatically set the bot's webhooks routes
	for _, webhook := range bot.Webhooks() {
		fmt.Println(webhook.Method(), webhook.Path())

		switch webhook.Method() {
		case core.HTTP_METHOD_GET:
			r.Get(webhook.Path(), webhook.Handler())
		case core.HTTP_METHOD_POST:
			r.Post(webhook.Path(), webhook.Handler())
		}
	}

	fmt.Println("Listening on port", config.ListeningPort)
	http.ListenAndServe(":"+strconv.Itoa(config.ListeningPort), r)
}
