// Package pprotect provides a helper functions to handle runtime panics.
package pprotect

import (
	"fmt"
	"net/http"
)

// HTTPPanicHandler is a callback function type for handling panics in HTTP
// handlers. It receives HTTP request object, panic value and stack trace in
// byte slice.
type HTTPPanicHandler func(r *http.Request, val interface{}, stack []byte)

// HTTPRecoveryHandler creates a new HTTP middleware that recovers from panics.
// In case of panic this handler will call each panic handler in hphs slice and
// then return HTTP response with status code 500 (internal server error), body
// of the panic will be a panic value converted to string.
func HTTPRecoveryHandler(next http.Handler, hphs ...HTTPPanicHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, stack := Call(func() {
			next.ServeHTTP(w, r)
		})

		if val == nil {
			return
		}

		for _, hph := range hphs {
			hph(r, val, stack)
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", val)
	}
}
