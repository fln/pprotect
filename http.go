// Package pprotect provides a helper functions to handle runtime panics.
package pprotect

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// HTTPPanicHandler is a callback function type for handling panics in HTTP
// handlers. It receives HTTP request object, panic value and stack trace in
// byte slice.
type HTTPPanicHandler func(r *http.Request, val interface{}, stack []byte)

// StdoutHTTPPanicHandler is an instance of NewPrintHTTPPanicHandler which
// prints panics to STDOUT.
var StdoutHTTPPanicHandler = NewPrintHTTPPanicHandler(os.Stdout)

// HTTPRecoveryHandler creates a new HTTP middleware that recovers from panics.
// In case of panic this handler will call each panic handler in handlers slice
// and then return HTTP response with status code 500 (internal server error),
// body of the panic will be a panic value converted to string.
func HTTPRecoveryHandler(next http.Handler, handlers ...HTTPPanicHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, stack := Call(func() {
			next.ServeHTTP(w, r)
		})

		if val == nil {
			return
		}

		for _, hph := range handlers {
			hph(r, val, stack)
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", val)
	}
}

// NewPrintHTTPPanicHandler creates a new HTTP handler panic handler that will
// print panic object and stack trace to the given io.Writer. This panic handler
// is intender to be used in production and should only be used when
// experimenting with this package or as a quick demo on how to implement HTTP
// panic handlers.
func NewPrintHTTPPanicHandler(w io.Writer) HTTPPanicHandler {
	return func(r *http.Request, val interface{}, stack []byte) {
		fmt.Fprintf(w, "PANIC: %s %s %s\n%s\n%s\n", r.RemoteAddr, r.Method, r.URL.String(), val, stack)
	}
}
