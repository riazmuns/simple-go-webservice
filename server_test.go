// server_test.go

package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEchoHandler(t *testing.T) {
	// Create a request to the http endpoint
	req, err := http.NewRequest("GET", "/echo", nil)
	if err != nil {
		t.Fatal(err)
	}

	// create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call your HTTP handler function (in this case, it's the root handler)
	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call your handler function
		echoDemo(w, r)
	}).ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	expected := fmt.Sprintf("Hello you requested %q", req.URL)

	// use go test ./... -v
	// to print the Log
	//t.Log(rr.Body.String(), expected)

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
