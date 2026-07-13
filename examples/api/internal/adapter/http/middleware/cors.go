package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

func Cors(
	allowedOrigins []string,
	allowedHeaders []string,
	exposedHeaders []string,
	debug ...bool,
) func(handler http.Handler) http.Handler {
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"*"}
	}

	options := cors.Options{
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   allowedHeaders,
		AllowCredentials: true,
		MaxAge:           300,
	}

	if len(debug) > 0 {
		options.Debug = debug[0]
	}

	if len(allowedOrigins) > 0 {
		options.AllowedOrigins = allowedOrigins
	}

	if len(exposedHeaders) > 0 {
		options.ExposedHeaders = exposedHeaders
	}

	return cors.Handler(options)
}
