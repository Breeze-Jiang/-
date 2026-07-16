package app

import (
	"os"
	"strings"
	"testing"
)

func TestOpenAPICoversPublicRoutesAndResponseSchemas(t *testing.T) {
	raw, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}
	document := string(raw)
	operations := []struct {
		path   string
		method string
		ref    string
	}{
		{"/auth/sms/send", "post", "ChallengeEnvelope"},
		{"/auth/sms/verify", "post", "TokensEnvelope"},
		{"/auth/refresh", "post", "TokensEnvelope"},
		{"/places/nearby", "get", "PlacePageEnvelope"},
		{"/places/search", "get", "PlacePageEnvelope"},
		{"/places/{placeId}", "get", "PlaceEnvelope"},
		{"/places/{placeId}/confirmations", "post", "ConfirmationEnvelope"},
		{"/places/{placeId}/feedback", "post", "FeedbackEnvelope"},
		{"/navigation/external-links", "get", "ExternalLinksEnvelope"},
		{"/navigation/routes", "post", "RouteEnvelope"},
	}
	for _, operation := range operations {
		block := openAPIOperationBlock(t, document, operation.path, operation.method)
		if !strings.Contains(block, "#/components/schemas/"+operation.ref) {
			t.Errorf("%s %s has no typed success response %s", operation.method, operation.path, operation.ref)
		}
	}
	logout := openAPIOperationBlock(t, document, "/auth/logout", "post")
	if !strings.Contains(logout, "'204':") {
		t.Error("POST /auth/logout does not declare its 204 response")
	}
	confirmation := openAPIOperationBlock(t, document, "/places/{placeId}/confirmations", "post")
	for _, required := range []string{"requestBody:", "ConfirmationRequest", "Idempotency-Key", "'404':", "'409':", "'422':", "'500':"} {
		if !strings.Contains(confirmation, required) {
			t.Errorf("confirmation contract is missing %q", required)
		}
	}
	feedback := openAPIOperationBlock(t, document, "/places/{placeId}/feedback", "post")
	for _, required := range []string{"FeedbackRequest", "Idempotency-Key", "'404':", "'409':", "'422':", "'500':"} {
		if !strings.Contains(feedback, required) {
			t.Errorf("feedback contract is missing %q", required)
		}
	}
	if !strings.Contains(document, "schema: { $ref: '#/components/schemas/ErrorEnvelope' }") {
		t.Error("shared error response is not bound to ErrorEnvelope")
	}
	for _, path := range []string{"/auth/sms/send", "/auth/sms/verify", "/auth/refresh", "/auth/logout"} {
		block := openAPIOperationBlock(t, document, path, "post")
		if !strings.Contains(block, "additionalProperties: false") {
			t.Errorf("%s request contract must reject unknown JSON fields", path)
		}
	}
	routeRequest := openAPIComponentBlock(t, document, "RouteRequest")
	if !strings.Contains(routeRequest, "additionalProperties: false") {
		t.Error("RouteRequest contract must reject unknown JSON fields")
	}
}

func openAPIOperationBlock(t *testing.T, document, path, method string) string {
	t.Helper()
	pathMarker := "  " + path + ":\n"
	pathStart := strings.Index(document, pathMarker)
	if pathStart < 0 {
		t.Fatalf("OpenAPI is missing path %s", path)
	}
	pathRest := document[pathStart+len(pathMarker):]
	pathEnd := strings.Index(pathRest, "\n  /")
	if pathEnd >= 0 {
		pathRest = pathRest[:pathEnd]
	}
	methodMarker := "    " + method + ":\n"
	methodStart := strings.Index(pathRest, methodMarker)
	if methodStart < 0 {
		t.Fatalf("OpenAPI is missing %s operation for %s", method, path)
	}
	return pathRest[methodStart+len(methodMarker):]
}

func openAPIComponentBlock(t *testing.T, document, name string) string {
	t.Helper()
	marker := "    " + name + ":\n"
	start := strings.Index(document, marker)
	if start < 0 {
		t.Fatalf("OpenAPI is missing schema %s", name)
	}
	rest := document[start+len(marker):]
	var block strings.Builder
	for _, line := range strings.Split(rest, "\n") {
		if strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "     ") {
			break
		}
		block.WriteString(line)
		block.WriteByte('\n')
	}
	return block.String()
}
