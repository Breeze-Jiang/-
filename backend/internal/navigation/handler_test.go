package navigation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExternalLinksRejectsEmptyDestinationName(t *testing.T) {
	handler := NewHandler(NewService(nil, ""))
	request := httptest.NewRequest(http.MethodGet, "/navigation/external-links?originLatitude=39.9&originLongitude=116.4&destinationLatitude=39.91&destinationLongitude=116.41&destinationName=%20%20&coordinateSystem=gcj02&mode=walking", nil)
	response := httptest.NewRecorder()
	handler.ExternalLinks(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_navigation_request"`) {
		t.Fatalf("unexpected response status=%d body=%s", response.Code, response.Body.String())
	}
}
