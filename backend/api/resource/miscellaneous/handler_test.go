package miscellaneous

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {

	req, err := http.NewRequest("GET", "/health/", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()

	HealthCheck(resp, req)

	if status := resp.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if body := resp.Body.String(); body != "Healthy" {
		t.Errorf("Handler returned wrong body: got %v want %v", body, "Healthy")
	}
}

func TestVersion(t *testing.T) {

	req, err := http.NewRequest("GET", "/version/", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp := httptest.NewRecorder()

	Version(resp, req)

	if status := resp.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if body := resp.Body.String(); body != "0.0.1" {
		t.Errorf("Handler returned wrong body: got %v want %v", body, "0.0.1")
	}
}
