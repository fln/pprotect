package pprotect

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func expectToPanic(t *testing.T, f func(), shouldPanic bool) {
	val, stack := Call(f)

	if shouldPanic && val == nil {
		t.Errorf("Call() did not captured panic value")
	}
	if !shouldPanic && val != nil {
		t.Errorf("Call() returned not nil on a non panicking funcion")
	}
	if val != nil && !bytes.HasPrefix(stack, []byte("goroutine ")) {
		t.Error("Call() did not returned a stack trace")
	}
}

func TestCall(t *testing.T) {
	goodFn := func() {}
	panicFn := func() { panic("Welp me!") }

	expectToPanic(t, goodFn, false)
	expectToPanic(t, panicFn, true)
}

type nPanicker int

func (n *nPanicker) do() {
	if *n > 0 {
		*n--
		panic("still panicking")
	}
	// return without panic
}

func expectNPanics(t *testing.T, n int) {
	obj := nPanicker(n)
	count := 0
	panicFn := func(i interface{}, stack []byte) {
		count++
	}

	CallLoop(obj.do, panicFn, time.Millisecond)
	if count != n {
		t.Errorf("CallLoop() called panic handler %d times, expected %d", count, n)
	}
}

func TestCallLoop(t *testing.T) {
	expectNPanics(t, 0)
	expectNPanics(t, 1)
	expectNPanics(t, 5)
}

func TestNewPrintPanicHandler(t *testing.T) {
	buf := new(bytes.Buffer)
	handler := NewPrintPanicHandler(buf)

	handler("panic message", []byte("stack trace"))
	expected := "PANIC: panic message\nstack trace"
	if buf.String() != expected {
		t.Errorf("NewPrintPanicHandler() produced invalid output \"%s\", expected \"%s\"", buf.String(), expected)
	}
}

func ExampleCallLoop() {
	myFunc := func() {
		// Some long running job, that might panic and should be
		//restarted
	}
	go CallLoop(myFunc, StdoutPanicHandler, time.Second)
}

func ExampleCallLoop_blocking() {
	myFunc := func() {
		// Some code that might panic and should be restarted
	}
	// Call to CallLoop will block until myFunc is finished without panics
	CallLoop(myFunc, StdoutPanicHandler, time.Second)
}

func ExampleCallLoop_closure() {
	go CallLoop(func() {
		fmt.Println("some", "arguments")
	}, StdoutPanicHandler, time.Second)
}
