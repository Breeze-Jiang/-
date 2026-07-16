//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code string `json:"code"`
	} `json:"error"`
}

type placePage struct {
	Items []struct {
		ID string `json:"id"`
	} `json:"items"`
}

func TestGuestEmergencyPlaceFlowAgainstComposeStack(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("E2E_API_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}
	client := &http.Client{Timeout: 3 * time.Second}
	waitForReady(t, client, baseURL)

	nearby := getEnvelope(t, client, baseURL+"/api/v1/places/nearby?latitude=39.908823&longitude=116.397470&coordinateSystem=gcj02&category=toilet")
	var page placePage
	if err := json.Unmarshal(nearby.Data, &page); err != nil {
		t.Fatalf("decode nearby response: %v", err)
	}
	if len(page.Items) == 0 || page.Items[0].ID == "" {
		t.Fatalf("nearby response has no materialized platform place: %s", nearby.Data)
	}
	placeID := page.Items[0].ID

	detail := getEnvelope(t, client, baseURL+"/api/v1/places/"+placeID)
	if !strings.Contains(string(detail.Data), `"id":"`+placeID+`"`) || !strings.Contains(string(detail.Data), `"trust"`) {
		t.Fatalf("unexpected detail response: %s", detail.Data)
	}

	routeBody := `{"origin":{"latitude":39.908823,"longitude":116.397470,"coordinateSystem":"gcj02"},"destination":{"latitude":39.918823,"longitude":116.407470,"coordinateSystem":"gcj02"},"mode":"walking"}`
	route := postEnvelope(t, client, baseURL+"/api/v1/navigation/routes", routeBody)
	if !strings.Contains(string(route.Data), `"distanceMeters":1200`) || !strings.Contains(string(route.Data), `"dataFreshness":"live"`) {
		t.Fatalf("unexpected route response: %s", route.Data)
	}

	links := getEnvelope(t, client, baseURL+"/api/v1/navigation/external-links?originLatitude=39.908823&originLongitude=116.397470&destinationLatitude=39.918823&destinationLongitude=116.407470&destinationName=%E6%B5%8B%E8%AF%95%E5%8E%95%E6%89%80&coordinateSystem=gcj02&mode=walking")
	for _, provider := range []string{"amap", "baidu", "tencent"} {
		if !strings.Contains(string(links.Data), `"provider":"`+provider+`"`) {
			t.Fatalf("external navigation response lacks %s: %s", provider, links.Data)
		}
	}
}

func waitForReady(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health/ready", nil)
		if err != nil {
			t.Fatalf("create health request: %v", err)
		}
		response, err := client.Do(request)
		if err == nil && response.StatusCode == http.StatusOK {
			response.Body.Close()
			return
		}
		if response != nil {
			response.Body.Close()
		}
		select {
		case <-ctx.Done():
			t.Fatalf("API did not become ready at %s: %v", baseURL, ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func getEnvelope(t *testing.T, client *http.Client, url string) envelope {
	t.Helper()
	response, err := client.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()
	return decodeEnvelope(t, response, url)
}

func postEnvelope(t *testing.T, client *http.Client, url, body string) envelope {
	t.Helper()
	response, err := client.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer response.Body.Close()
	return decodeEnvelope(t, response, url)
}

func decodeEnvelope(t *testing.T, response *http.Response, target string) envelope {
	t.Helper()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("%s status=%d", target, response.StatusCode)
	}
	var decoded envelope
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode %s: %v", target, err)
	}
	if decoded.Error != nil || len(decoded.Data) == 0 {
		t.Fatalf("unexpected error envelope for %s: %+v", target, decoded.Error)
	}
	return decoded
}
