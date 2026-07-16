//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

type smsChallenge struct {
	ChallengeID string `json:"challengeId"`
}

type sessionTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func TestAuthenticatedConfirmationFlowAgainstComposeStack(t *testing.T) {
	baseURL := e2eBaseURL()
	client := &http.Client{}
	code := os.Getenv("E2E_SMS_CODE")
	if code == "" {
		code = "123456"
	}

	challengeResponse := postJSON(t, client, baseURL+"/api/v1/auth/sms/send", map[string]any{
		"phone":    "13800138001",
		"deviceId": "compose-auth-test-device",
	}, nil, http.StatusAccepted)
	var challenge smsChallenge
	decodeE2EData(t, challengeResponse.Data, &challenge)
	if challenge.ChallengeID == "" {
		t.Fatal("SMS challenge response did not contain a challenge ID")
	}

	verifyResponse := postJSON(t, client, baseURL+"/api/v1/auth/sms/verify", map[string]any{
		"challengeId": challenge.ChallengeID,
		"code":        code,
		"deviceId":    "compose-auth-test-device",
	}, nil, http.StatusOK)
	var tokens sessionTokens
	decodeE2EData(t, verifyResponse.Data, &tokens)
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("SMS verification did not issue both access and refresh tokens")
	}

	nearby := getEnvelope(t, client, baseURL+"/api/v1/places/nearby?latitude=39.908823&longitude=116.397470&coordinateSystem=gcj02&category=toilet&limit=20")
	var page placePage
	decodeE2EData(t, nearby.Data, &page)
	if len(page.Items) == 0 || page.Items[0].ID == "" {
		t.Fatal("nearby API did not materialize a place for confirmation")
	}
	placeID := page.Items[0].ID

	headers := map[string]string{
		"Authorization":     "Bearer " + tokens.AccessToken,
		"Idempotency-Key": "compose-auth-confirmation-key",
	}
	confirmation := map[string]any{"kind": "exists", "result": "confirmed"}
	first := postJSON(t, client, baseURL+"/api/v1/places/"+placeID+"/confirmations", confirmation, headers, http.StatusCreated)
	second := postJSON(t, client, baseURL+"/api/v1/places/"+placeID+"/confirmations", confirmation, headers, http.StatusCreated)
	if !bytes.Equal(first.Data, second.Data) {
		t.Fatalf("same idempotency key did not return the original confirmation: first=%s second=%s", first.Data, second.Data)
	}

	postJSON(t, client, baseURL+"/api/v1/places/"+placeID+"/confirmations", map[string]any{"kind": "exists", "result": "incorrect"}, headers, http.StatusConflict)
	postJSON(t, client, baseURL+"/api/v1/auth/logout", map[string]any{"refreshToken": tokens.RefreshToken}, nil, http.StatusNoContent)
	postJSON(t, client, baseURL+"/api/v1/places/"+placeID+"/confirmations", confirmation, map[string]string{
		"Authorization":     "Bearer " + tokens.AccessToken,
		"Idempotency-Key": "compose-auth-revoked-key",
	}, http.StatusUnauthorized)
}

func e2eBaseURL() string {
	if baseURL := os.Getenv("E2E_API_BASE_URL"); baseURL != "" {
		return baseURL
	}
	return "http://127.0.0.1:8080"
}

func decodeE2EData(t *testing.T, raw json.RawMessage, target any) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode response data: %v; data=%s", err, raw)
	}
}

func postJSON(t *testing.T, client *http.Client, url string, body any, headers map[string]string, wantStatus int) envelope {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request for %s: %v", url, err)
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create request for %s: %v", url, err)
	}
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		t.Fatalf("POST %s status=%d want=%d", url, response.StatusCode, wantStatus)
	}
	if wantStatus == http.StatusNoContent {
		return envelope{}
	}
	var decoded envelope
	if err = json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
	return decoded
}
