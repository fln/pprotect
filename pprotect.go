// Package pprotect provides a helper functions to handle runtime panics.
package pprotect

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// PanicHandler is a callback function type. It gets called with has panic value
// and stack trace dump.
type PanicHandler func(val interface{}, stack []byte)

// StackSize is the upper limit of bytes of stacktrace that will be passed to
// the panic handler. Default value of 8 kb should be suitable in most cases.
// If there is a need for larger stack traces this value can be changed during
// application execution. Note: this variable is not protected by any
// synchronization primitive, thus this value should be set before using any of
// the functions in this package.
var StackSize = 1024 * 8

// StdoutPanicHandler is an instance of NewPrintPanicHandler which prints panics
// to STDOUT.
var StdoutPanicHandler = NewPrintPanicHandler(os.Stdout)

// Call executes funcion f and converts panic to a return argument val. It also
// returns a stack trace as a byte slice. If function f does not panic nil, nil
// is returned. Call is useful for converting panics to return agruments.
func Call(f func()) (val interface{}, stack []byte) {
	defer func() {
		if val = recover(); val != nil {
			stack = make([]byte, StackSize)
			stack = stack[:runtime.Stack(stack, false)]
		}
	}()
	f()
	return
}

// CallLoop executes a function f. If function panics it will call a panic
// handler with panic value and stack trace and restart the function after time
// specified in restartAfter. It is useful for protecting long runing global
// goroutines from bringing whole application down in case of unexpected panic.
// If function f exits without panicking, function will not be restarted.
//
// To wrap function with arguments use a closure.
func CallLoop(f func(), handler PanicHandler, restartAfter time.Duration) {
	for {
		val, stack := Call(f)
		if val == nil {
			return
		}
		handler(val, stack)
		time.Sleep(restartAfter)
	}
}

// NewPrintPanicHandler creates a new panic handler that will print panic object
// and stack trace to the given io.Writer. This panic handler is not production
// ready and should only be used when experimenting with this package or as a
// quick demo on how to implement panic handlers.
func NewPrintPanicHandler(w io.Writer) PanicHandler {
	return func(val interface{}, stack []byte) {
		fmt.Fprintf(w, "PANIC: %v\n%s", val, stack)
	}
}
