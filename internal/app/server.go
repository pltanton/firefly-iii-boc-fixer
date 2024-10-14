package app

import (
	"net/http"

	"github.com/pltanton/firefly-iii-boc-fixer/internal/config"
	"github.com/pltanton/firefly-iii-boc-fixer/internal/firefly"
)

func NewServer(
	c config.Config,
	fireflyClient *firefly.FireflyClient,
) http.Handler {
	mux := &http.ServeMux{}

	addRoutes(mux, c, fireflyClient)

	var handler http.Handler = mux

	// Set middlewares

	return handler
}

func addRoutes(
	mux *http.ServeMux,
	c config.Config,
	fireflyClient *firefly.FireflyClient,
) {
	mux.Handle("/webhook", handleWebhook(c, fireflyClient))
}
