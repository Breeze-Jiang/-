package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRejectsTrailingJSONValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"first"}{"name":"second"}`))
	response := httptest.NewRecorder()
	var payload struct {
		Name string `json:"name"`
	}
	if Decode(response, request, &payload) {
		t.Fatal("multiple JSON values were accepted")
	}
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("unexpected invalid JSON response: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestDecodeAcceptsOneJSONValueWithWhitespace(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("  {\"name\":\"one\"} \n"))
	response := httptest.NewRecorder()
	var payload struct {
		Name string `json:"name"`
	}
	if !Decode(response, request, &payload) || payload.Name != "one" {
		t.Fatalf("single JSON value rejected: payload=%+v body=%s", payload, response.Body.String())
	}
}
