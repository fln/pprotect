package pprotect

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPanicHandler struct {
	callCount int
}

type mockHTTPHandler struct {
	shouldPanic bool
	panicVal    interface{}
}

func (h *mockPanicHandler) handle(r *http.Request, val interface{}, stack []byte) {
	h.callCount++
}

func (h *mockHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.shouldPanic {
		panic(h.panicVal)
	}
}

func TestHTTPRecoveryHandlerIfPanics(t *testing.T) {
	next := mockHTTPHandler{shouldPanic: true, panicVal: "oops"}
	handler := &mockPanicHandler{}

	recovery := HTTPRecoveryHandler(&next, handler.handle, handler.handle)
	response := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		recovery.ServeHTTP(response, &http.Request{})
	})
	assert.Equal(t, 2, handler.callCount)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	require.NotNil(t, response.Body)
	assert.Equal(t, fmt.Sprintf("%s\n", next.panicVal), response.Body.String())
}

func TestHTTPRecoveryHandlerIfNotsPanics(t *testing.T) {
	next := mockHTTPHandler{shouldPanic: false}
	handler := &mockPanicHandler{}

	recovery := HTTPRecoveryHandler(&next, handler.handle, handler.handle)
	response := httptest.NewRecorder()

	recovery.ServeHTTP(response, &http.Request{})
	assert.Equal(t, 0, handler.callCount)
	assert.Equal(t, http.StatusOK, response.Code)
	require.NotNil(t, response.Body)
	assert.Equal(t, "", response.Body.String())
}
