package telemetry

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

var (
	NRLicense = "Unknown" // The NewRelic ingest key to use
	NRAppName = "pack-assist.cagedtornado.com"
	NRApp     = &newrelic.Application{}
)

func init() {

	//	If we have NR environment variables, use them:
	NRLicense = os.Getenv("NEW_RELIC_LICENSE_KEY")

	err := *new(error)
	NRApp, err = newrelic.NewApplication(
		newrelic.ConfigAppName(NRAppName),
		newrelic.ConfigLicense(NRLicense),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		log.Err(err).
			Str("NRAppName", NRAppName).
			Str("NRLicense", NRLicense).
			Msg("Problem trying to create a New Relic connection")
	}
}

// NewTransactionContext creates a NewRelic transaction,
// and returns the transaction -- so you can defer End() it
// and a context with the transaction
func NewTransactionContext(ctx context.Context, transactionName string) (context.Context, *newrelic.Transaction) {

	//	Start a new transaction
	txn := NRApp.StartTransaction(transactionName)

	//	Add it to the context
	cx := newrelic.NewContext(ctx, txn)

	return cx, txn
}

// Middleware to create new newrelic http middleware.
func Middleware(nrApp *newrelic.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get route pattern.
			path, ok := getRoutePattern(r)
			if !ok {
				path = r.RequestURI
			}

			// Start newrelic transaction.
			tx := nrApp.StartTransaction(r.Method + " " + path)
			defer tx.End()

			w = tx.SetWebResponse(w)
			tx.SetWebRequestHTTP(r)

			r = newrelic.RequestWithTransactionContext(r, tx)

			next.ServeHTTP(w, r)
		})
	}
}

func getRoutePattern(r *http.Request) (string, bool) {
	routePath := r.URL.Path
	if r.URL.RawPath != "" {
		routePath = r.URL.RawPath
	}

	rctx := chi.RouteContext(r.Context())
	tctx := chi.NewRouteContext()
	if rctx.Routes.Match(tctx, r.Method, routePath) {
		return tctx.RoutePattern(), true
	}

	return "", false
}
